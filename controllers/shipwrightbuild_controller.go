// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"sort"
	errors2 "github.com/pkg/errors"
	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	tektonoperatorv1alpha1client "github.com/tektoncd/operator/pkg/client/clientset/versioned/typed/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/pkg/certmanager"
	"github.com/shipwright-io/operator/pkg/common"
	"github.com/shipwright-io/operator/pkg/tekton"
)

const (
	// FinalizerAnnotation annotation string appended on finalizer slice.
	FinalizerAnnotation = "finalizer.operator.shipwright.io"
	// defaultTargetNamespace fallback namespace when `.spec.namepace` is not informed.
	defaultTargetNamespace = "shipwright-build"

	// Ready object is providing service.
	ConditionReady = "Ready"

	// UseManagedWebhookCerts is an env Var that controls wether we install the webhook certs
	UseManagedWebhookCerts = "USE_MANAGED_WEBHOOK_CERTS"
)

// ShipwrightBuildReconciler reconciles a ShipwrightBuild object
type ShipwrightBuildReconciler struct {
	client.Client        // controller kubernetes client
	CRDClient            crdclientv1.ApiextensionsV1Interface
	TektonOperatorClient tektonoperatorv1alpha1client.OperatorV1alpha1Interface

	Logger         logr.Logger           // decorated logger
	Scheme         *runtime.Scheme       // runtime scheme
	Manifest       manifestival.Manifest // release manifests render
	TektonManifest manifestival.Manifest // Tekton release manifest render
}

// setFinalizer append finalizer on the resource, and uses local client to update it immediately.
func (r *ShipwrightBuildReconciler) setFinalizer(ctx context.Context, b *v1alpha1.ShipwrightBuild) error {
	if common.Contains(b.GetFinalizers(), FinalizerAnnotation) {
		return nil
	}
	b.SetFinalizers(append(b.GetFinalizers(), FinalizerAnnotation))
	return r.Update(ctx, b, &client.UpdateOptions{})
}

// unsetFinalizer remove all instances of local finalizer string, updating the resource immediately.
func (r *ShipwrightBuildReconciler) unsetFinalizer(ctx context.Context, b *v1alpha1.ShipwrightBuild) error {
	finalizers := []string{}
	for _, f := range b.GetFinalizers() {
		if f == FinalizerAnnotation {
			continue
		}
		finalizers = append(finalizers, f)
	}

	b.SetFinalizers(finalizers)
	return r.Update(ctx, b, &client.UpdateOptions{})
}

// Reconcile performs the resource reconciliation steps to deploy or remove Shipwright Build
// instances. When deletion-timestamp is found, the removal of the previously deploy resources is
// executed, otherwise the regular deploy workflow takes place.
func (r *ShipwrightBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("namespace", req.Namespace, "name", req.Name)
	logger.Info("Starting resource reconciliation...")
	// ReconcileTekton
	_, requeue, err := tekton.ReconcileTekton(ctx, r.CRDClient, r.TektonOperatorClient)
	if err != nil {
		return ctrl.Result{Requeue: requeue}, err
	}
	if requeue {
		return Requeue()
	}

	bList := &v1alpha1.ShipwrightBuildList{}
	err = r.Client.List(context.TODO(), bList, &client.ListOptions{})
	if err != nil {
		return ctrl.Result{}, errors2.Wrap(err, "failed listing all Shipwright instances")
	}

	// retrieving the ShipwrightBuild instance requested for reconcile
	b := &v1alpha1.ShipwrightBuild{}
	if err := r.Get(ctx, req.NamespacedName, b); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource is not found!")
			return NoRequeue()
		}
		logger.Error(err, "retrieving ShipwrightBuild object from cache")
		return RequeueOnError(err)
	}
	if len(bList.Items) > 0 {
		if len(bList.Items) > 1 {
			sort.Slice(bList.Items, func(i, j int) bool {
				return bList.Items[j].CreationTimestamp.After(bList.Items[i].CreationTimestamp.Time)
			})
		}
		if bList.Items[0].Name != req.Name {
			logger.Info("Ignoring ShipwrightBuild.operator.shipwright.io because one already exists and does not match existing name")
			err = r.Client.Delete(context.TODO(), b, &client.DeleteOptions{})
			if err != nil {
				logger.Error(err, "failed to remove ShipwrightBuild.operator.shipwright.io instance")
			}
			return ctrl.Result{}, nil
		}
	}
	init := b.Status.Conditions == nil
	if init {
		b.Status.Conditions = make([]metav1.Condition, 0)
		apimeta.SetStatusCondition(&b.Status.Conditions, metav1.Condition{
			Type:    ConditionReady,
			Status:  metav1.ConditionUnknown, // we just started trying to reconcile
			Reason:  "Init",
			Message: "Initializing Shipwright Operator",
		})
		if err := r.Client.Status().Update(ctx, b); err != nil {
			return RequeueWithError(err)
		}
	}

	// selecting the target namespace based on the CRD information, when not informed using the
	// default namespace instead
	targetNamespace := b.Spec.TargetNamespace
	if targetNamespace == "" {
		logger.Info(
			"Namespace is not informed! Target namespace is selected from default settings instead",
			"defaultTargetNamespace", defaultTargetNamespace,
		)
		targetNamespace = defaultTargetNamespace
	}
	logger = logger.WithValues("targetNamespace", targetNamespace)
	// create if it does not exist
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, types.NamespacedName{Name: targetNamespace}, ns); err != nil {
		if !errors.IsNotFound(err) {
			logger.Info("retrieving target namespace %s error: %s", targetNamespace, err.Error())
			return RequeueOnError(err)
		}
		ns.Name = targetNamespace

		if err = r.Create(ctx, ns, &client.CreateOptions{Raw: &metav1.CreateOptions{}}); err != nil {
			if !errors.IsAlreadyExists(err) {
				logger.Info("creating target namespace %s error: %s", targetNamespace, err.Error())
				return RequeueOnError(err)
			}
		}
		logger.Info("created target namespace")
	}

	// ReconcileCertManager
	if common.BoolFromEnvVar(UseManagedWebhookCerts) {
		requeue, err = certmanager.ReconcileCertManager(ctx, r.CRDClient, r.Client, r.Logger, targetNamespace)
		if err != nil {
			return ctrl.Result{Requeue: requeue}, err
		}
		if requeue {
			return Requeue()
		}
	}

	// filtering out namespace resource, so it does not create new namespaces accidentally, and
	// transforming object to target the namespace informed on the CRD (.spec.namespace)
	images := common.ToLowerCaseKeys(common.ImagesFromEnv(common.ShipwrightImagePrefix))
	manifest, err := r.Manifest.
	Filter(manifestival.Not(manifestival.ByKind("Namespace"))).
	Transform(manifestival.InjectNamespace(targetNamespace), common.DeploymentImages(images))
	if err != nil {
		logger.Error(err, "transforming manifests, injecting namespace")
		return RequeueWithError(err)
	}

	// when deletion-timestamp is set, the reconciliation process is in fact deleting the resources
	// previously deployed. To mark the deletion process as done, it needs to clean up the
	// finalizers, and thus the ShipwrightBuild is removed from cache
	if !b.GetDeletionTimestamp().IsZero() {
		logger.Info("DeletionTimestamp is set...")
		if !common.Contains(b.GetFinalizers(), FinalizerAnnotation) {
			logger.Info("Finalizers removed, deletion of manifests completed!")
			return NoRequeue()
		}

		logger.Info("Deleting manifests...")
		if err := manifest.Delete(); err != nil {
			logger.Error(err, "deleting manifest's resources")
			return RequeueWithError(err)
		}
		logger.Info("Removing finalizers...")
		if err := r.unsetFinalizer(ctx, b); err != nil {
			logger.Error(err, "removing the finalizer")
			return RequeueWithError(err)
		}
		logger.Info("All removed!")
		return NoRequeue()
	}

	// rolling out the resources described on the manifests, it should create a new Shipwright Build
	// instance with required dependencies
	logger.Info("Applying manifest's resources...")
	if err := manifest.Apply(); err != nil {
		logger.Error(err, "rolling out manifest's resources")
		apimeta.SetStatusCondition(&b.Status.Conditions, metav1.Condition{
			Type:    ConditionReady,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: fmt.Sprintf("Reconciling ShipwrightBuild failed: %v", err),
		})
		err = r.Client.Status().Update(ctx, b)
		return RequeueWithError(err)
	}
	if err := r.setFinalizer(ctx, b); err != nil {
		logger.Info(fmt.Sprintf("%#v", b))
		logger.Error(err, "setting the finalizer")
		return RequeueWithError(err)
	}
	apimeta.SetStatusCondition(&b.Status.Conditions, metav1.Condition{
		Type:    ConditionReady,
		Status:  metav1.ConditionTrue,
		Reason:  "Success",
		Message: "Reconciled ShipwrightBuild successfully",
	})
	if err := r.Client.Status().Update(ctx, b); err != nil {
		logger.Error(err, "updating ShipwrightBuild status")
		RequeueWithError(err) //nolint:errcheck
	}
	logger.Info("All done!")
	return NoRequeue()
}

// setupManifestival instantiate manifestival with local controller attributes, as well as tekton prereqs.
func (r *ShipwrightBuildReconciler) setupManifestival() error {
	var err error
	r.Manifest, err = common.SetupManifestival(r.Client, "release.yaml", r.Logger)
	return err
}

// SetupWithManager sets up the controller with the Manager, by instantiating Manifestival and
// setting up watch and predicate rules for ShipwrightBuild objects.
func (r *ShipwrightBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.setupManifestival(); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
	For(&v1alpha1.ShipwrightBuild{}, builder.WithPredicates(predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			// all new objects must be subject to reconciliation
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// objects that haven't been confirmed deleted must be subject to reconciliation
			return !e.DeleteStateUnknown
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// objects that have updated generation must be subject to reconciliation
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	})).
	Complete(r)
}
