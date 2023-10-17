// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"

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
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/shipwright-io/operator/api/v1alpha1"
	commonctrl "github.com/shipwright-io/operator/controllers/common"
	"github.com/shipwright-io/operator/pkg/reconciler/common"
	"github.com/shipwright-io/operator/pkg/reconciler/tekton"
)

// BuildReconciler reconciles a ShipwrightBuild object
type BuildReconciler struct {
	client.Client        // controller kubernetes client
	CRDClient            crdclientv1.ApiextensionsV1Interface
	TektonOperatorClient tektonoperatorv1alpha1client.OperatorV1alpha1Interface

	Logger         logr.Logger           // decorated logger
	Scheme         *runtime.Scheme       // runtime scheme
	Manifest       manifestival.Manifest // release manifests render
	TektonManifest manifestival.Manifest // Tekton release manifest render
}

// Add creates a new build Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (*BuildReconciler, error) {
	c := mgr.GetClient()
	scheme := mgr.GetScheme()
	logger := ctrl.Log.WithName("controllers").WithName("build")

	crdClient, err := crdclientv1.NewForConfig(mgr.GetConfig())
	if err != nil {
		logger.Error(err, "unable to get crd client")
		return nil, err
	}

	tektonOperatorClient, err := tektonoperatorv1alpha1client.NewForConfig(mgr.GetConfig())
	if err != nil {
		logger.Error(err, "unable to get tekton operator client")
		return nil, err
	}

	return &BuildReconciler{
		CRDClient:            crdClient,
		TektonOperatorClient: tektonOperatorClient,
		Client:               c,
		Scheme:               scheme,
		Logger:               logger,
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *BuildReconciler) error {

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

// setFinalizer append finalizer on the resource, and uses local client to update it immediately.
func (r *BuildReconciler) setFinalizer(ctx context.Context, b *v1alpha1.ShipwrightBuild) error {
	if common.Contains(b.GetFinalizers(), commonctrl.FinalizerAnnotation) {
		return nil
	}
	b.SetFinalizers(append(b.GetFinalizers(), commonctrl.FinalizerAnnotation))
	return r.Update(ctx, b, &client.UpdateOptions{})
}

// unsetFinalizer remove all instances of local finalizer string, updating the resource immediately.
func (r *BuildReconciler) unsetFinalizer(ctx context.Context, b *v1alpha1.ShipwrightBuild) error {
	finalizers := []string{}
	for _, f := range b.GetFinalizers() {
		if f == commonctrl.FinalizerAnnotation {
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
func (r *BuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("namespace", req.Namespace, "name", req.Name)
	// ReconcileTekton
	_, requeue, err := tekton.ReconcileTekton(ctx, r.CRDClient, r.TektonOperatorClient)
	if err != nil {
		return commonctrl.RequeueWithError(err)
	}
	if requeue {
		return commonctrl.Requeue()
	}

	// retrieving the ShipwrightBuild instance requested for reconcile
	b := &v1alpha1.ShipwrightBuild{}
	if err := r.Get(ctx, req.NamespacedName, b); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource is not found!")
			return commonctrl.NoRequeue()
		}
		logger.Error(err, "retrieving ShipwrightBuild object from cache")
		return commonctrl.RequeueOnError(err)
	}
	init := b.Status.Conditions == nil
	if init {
		b.Status.Conditions = make([]metav1.Condition, 0)
		apimeta.SetStatusCondition(&b.Status.Conditions, metav1.Condition{
			Type:    commonctrl.ConditionReady,
			Status:  metav1.ConditionUnknown, // we just started trying to reconcile
			Reason:  "Init",
			Message: "Initializing Shipwright Operator",
		})
		if err := r.Client.Status().Update(ctx, b); err != nil {
			return commonctrl.RequeueWithError(err)
		}
	}

	// selecting the target namespace based on the CRD information, when not informed using the
	// default namespace instead
	targetNamespace := b.Spec.TargetNamespace
	if targetNamespace == "" {
		logger.Info(
			"Namespace is not informed! Target namespace is selected from default settings instead",
			"defaultTargetNamespace", commonctrl.DefaultTargetNamespace,
		)
		targetNamespace = commonctrl.DefaultTargetNamespace
	}
	logger = logger.WithValues("targetNamespace", targetNamespace)
	// create if it does not exist
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, types.NamespacedName{Name: targetNamespace}, ns); err != nil {
		if !errors.IsNotFound(err) {
			logger.Info("retrieving target namespace %s error: %s", targetNamespace, err.Error())
			return commonctrl.RequeueOnError(err)
		}
		ns.Name = targetNamespace

		if err = r.Create(ctx, ns, &client.CreateOptions{Raw: &metav1.CreateOptions{}}); err != nil {
			if !errors.IsAlreadyExists(err) {
				logger.Info("creating target namespace %s error: %s", targetNamespace, err.Error())
				return commonctrl.RequeueOnError(err)
			}
		}
		logger.Info("created target namespace")
	}

	// filtering out namespace resource, so it does not create new namespaces accidentally, and
	// transforming object to target the namespace informed on the CRD (.spec.namespace)
	images := common.ToLowerCaseKeys(common.ImagesFromEnv(common.ShipwrightImagePrefix))

	transformerfncs := []manifestival.Transformer{}
	if common.IsOpenShiftPlatform() {
		transformerfncs = append(transformerfncs, manifestival.InjectNamespace(targetNamespace))
		transformerfncs = append(transformerfncs, common.DeploymentImages(images))
	} else {
		transformerfncs = append(transformerfncs, manifestival.InjectNamespace(targetNamespace))
		transformerfncs = append(transformerfncs, common.DeploymentImages(images))
		// Add Transformer to inject cert-manager annotation in crds to handle conversion spec
		transformerfncs = append(transformerfncs, common.InjectAnnotations(commonctrl.CertManagerInjectAnnotationKey, fmt.Sprintf(commonctrl.CertManagerInjectAnnotationValueTemplate, targetNamespace), common.Overwrite, "CustomResourceDefinition"))
	}

	manifest, err := r.Manifest.
		Filter(manifestival.Not(manifestival.ByKind("Namespace"))).
		Transform(transformerfncs...)

	if err != nil {
		logger.Error(err, "transforming manifests, injecting namespace")
		return commonctrl.RequeueWithError(err)
	}

	// when deletion-timestamp is set, the reconciliation process is in fact deleting the resources
	// previously deployed. To mark the deletion process as done, it needs to clean up the
	// finalizers, and thus the ShipwrightBuild is removed from cache
	if !b.GetDeletionTimestamp().IsZero() {
		logger.Info("DeletionTimestamp is set...")
		if !common.Contains(b.GetFinalizers(), commonctrl.FinalizerAnnotation) {
			logger.Info("Finalizers removed, deletion of manifests completed!")
			return commonctrl.NoRequeue()
		}

		logger.Info("Deleting manifests...")
		if err := manifest.Delete(); err != nil {
			logger.Error(err, "deleting manifest's resources")
			return commonctrl.RequeueWithError(err)
		}
		logger.Info("Removing finalizers...")
		if err := r.unsetFinalizer(ctx, b); err != nil {
			logger.Error(err, "removing the finalizer")
			return commonctrl.RequeueWithError(err)
		}
		logger.Info("All removed!")
		return commonctrl.NoRequeue()
	}

	// rolling out the resources described on the manifests, it should create a new Shipwright Build
	// instance with required dependencies
	logger.Info("Applying manifest's resources...")
	if err := manifest.Apply(); err != nil {
		logger.Error(err, "rolling out manifest's resources")
		apimeta.SetStatusCondition(&b.Status.Conditions, metav1.Condition{
			Type:    commonctrl.ConditionReady,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: fmt.Sprintf("Reconciling ShipwrightBuild failed: %v", err),
		})
		err = r.Client.Status().Update(ctx, b)
		return commonctrl.RequeueWithError(err)
	}
	if err := r.setFinalizer(ctx, b); err != nil {
		logger.Info(fmt.Sprintf("%#v", b))
		logger.Error(err, "setting the finalizer")
		return commonctrl.RequeueWithError(err)
	}
	apimeta.SetStatusCondition(&b.Status.Conditions, metav1.Condition{
		Type:    commonctrl.ConditionReady,
		Status:  metav1.ConditionTrue,
		Reason:  "Success",
		Message: "Reconciled ShipwrightBuild successfully",
	})
	if err := r.Client.Status().Update(ctx, b); err != nil {
		logger.Error(err, "updating ShipwrightBuild status")
		commonctrl.RequeueWithError(err) //nolint:errcheck
	}
	logger.Info("All done!")
	return commonctrl.NoRequeue()
}

// setupManifestival instantiate manifestival with local controller attributes, as well as tekton prereqs.
func (r *BuildReconciler) setupManifestival() error {
	var err error
	r.Manifest, err = common.SetupManifestival(r.Client, []string{common.BuildDataDir}, r.Logger)
	return err
}
