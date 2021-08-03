// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	mfc "github.com/manifestival/controller-runtime-client"
	"github.com/manifestival/manifestival"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/shipwright-io/operator/api/v1alpha1"
)

const (
	// FinalizerAnnotation annotation string appended on finalizer slice.
	FinalizerAnnotation = "finalizer.operator.shipwright.io"
	// defaultTargetNamespace fallback namespace when `.spec.namepace` is not informed.
	defaultTargetNamespace = "shipwright-build"
)

// ShipwrightBuildReconciler reconciles a ShipwrightBuild object
type ShipwrightBuildReconciler struct {
	client.Client // controller kubernetes client

	Logger   logr.Logger           // decorated logger
	Scheme   *runtime.Scheme       // runtime scheme
	Manifest manifestival.Manifest // release manifests render
}

// setFinalizer append finalizer on the resource, and uses local client to update it immediately.
func (r *ShipwrightBuildReconciler) setFinalizer(ctx context.Context, b *v1alpha1.ShipwrightBuild) error {
	if contains(b.GetFinalizers(), FinalizerAnnotation) {
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

	// retrieving the ShipwrightBuild instance requested for reconciliation
	b := &v1alpha1.ShipwrightBuild{}
	if err := r.Get(ctx, req.NamespacedName, b); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource is not found!")
			return NoRequeue()
		}
		logger.Error(err, "Retrieving ShipwrightBuild object from cache")
		return RequeueOnError(err)
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

	// filtering out namespace resource, so it does not create new namespaces accidentally, and
	// transforming object to target the namespace informed on the CRD (.spec.namespace)
	manifest, err := r.Manifest.
		Filter(manifestival.Not(manifestival.ByKind("Namespace"))).
		Transform(manifestival.InjectNamespace(targetNamespace))
	if err != nil {
		logger.Error(err, "Transforming manifests, injecting namespace")
		return RequeueWithError(err)
	}

	// when deletion-timestamp is set, the reconciliation process is in fact deleting the resources
	// previously deployed. To mark the deletion process as done, it needs to clean up the
	// finalizers, and thus the ShipwrightBuild is removed from cache
	if !b.GetDeletionTimestamp().IsZero() {
		logger.Info("DeletionTimestamp is set...")
		if !contains(b.GetFinalizers(), FinalizerAnnotation) {
			logger.Info("Finalizers removed, deletion of manifests completed!")
			return NoRequeue()
		}

		logger.Info("Deleting manifests...")
		if err := manifest.Delete(); err != nil {
			logger.Error(err, "Deleting manifest's resources")
			return RequeueWithError(err)
		}
		logger.Info("Removing finalizers...")
		if err := r.unsetFinalizer(ctx, b); err != nil {
			logger.Error(err, "Removing the finalizer")
			return RequeueWithError(err)
		}
		logger.Info("All removed!")
		return NoRequeue()
	}

	// rolling out the resources described on the manifests, it should create a new Shipwright Build
	// instance with required dependencies
	logger.Info("Applying manifest's resources...")
	if err := manifest.Apply(); err != nil {
		logger.Error(err, "Rolling out manifest's resources")
		return RequeueWithError(err)
	}
	if err := r.setFinalizer(ctx, b); err != nil {
		logger.Info(fmt.Sprintf("%#v", b))
		logger.Error(err, "Setting the finalizer")
		return RequeueWithError(err)
	}
	logger.Info("All done!")
	return NoRequeue()
}

// setupManifestival instantiate manifestival with local controller attributes.
func (r *ShipwrightBuildReconciler) setupManifestival(managerLogger logr.Logger) error {
	client := mfc.NewClient(r.Client)
	logger := managerLogger.WithName("manifestival")

	dataPath, err := koDataPath()
	if err != nil {
		return err
	}
	buildManifest := filepath.Join(dataPath, "release.yaml")

	r.Manifest, err = manifestival.NewManifest(
		buildManifest,
		manifestival.UseClient(client),
		manifestival.UseLogger(logger),
	)
	return err
}

// SetupWithManager sets up the controller with the Manager, by instantiating Manifestival and
// setting up watch and predicate rules for ShipwrightBuild objects.
func (r *ShipwrightBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.setupManifestival(mgr.GetLogger()); err != nil {
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
