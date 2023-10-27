// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package buildstrategy

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	corev1 "k8s.io/api/core/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/operator/api/v1alpha1"
	commonctrl "github.com/shipwright-io/operator/controllers/common"
	"github.com/shipwright-io/operator/pkg/reconciler/build"
)

// BuildStrategyReconciler reconciles a ShipwrightBuild object
type BuildStrategyReconciler struct {
	client.Client // controller kubernetes client
	CRDClient     crdclientv1.ApiextensionsV1Interface

	Logger   logr.Logger           // decorated logger
	Scheme   *runtime.Scheme       // runtime scheme
	Manifest manifestival.Manifest // release manifests render
}

// Add creates a new buildStrategy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (*BuildStrategyReconciler, error) {
	c := mgr.GetClient()
	scheme := mgr.GetScheme()
	logger := ctrl.Log.WithName("controllers").WithName("buildstrategy")

	crdClient, err := crdclientv1.NewForConfig(mgr.GetConfig())
	if err != nil {
		logger.Error(err, "unable to get crd client")
		return nil, err
	}

	return &BuildStrategyReconciler{
		CRDClient: crdClient,
		Client:    c,
		Scheme:    scheme,
		Logger:    logger,
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *BuildStrategyReconciler) error {

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

// Reconcile performs the resource reconciliation steps to deploy or remove Shipwright Build
// instances. When deletion-timestamp is found, the removal of the previously deploy resources is
// executed, otherwise the regular deploy workflow takes place.
func (r *BuildStrategyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("namespace", req.Namespace, "name", req.Name)

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

	// Check targetNamespace is created
	targetNamespace := b.Spec.TargetNamespace
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, types.NamespacedName{Name: targetNamespace}, ns); err != nil {
		if !errors.IsNotFound(err) {
			logger.Info("retrieving target namespace %s error: %s", targetNamespace, err.Error())
			return commonctrl.RequeueAfterWithError(err)
		}
	}

	// Reconcile BuildStrategy
	requeue, err := build.ReconcileBuildStrategy(ctx, r.CRDClient, r.Client, r.Logger, targetNamespace)
	if err != nil {
		return commonctrl.RequeueAfterWithError(err)
	}
	if requeue {
		return commonctrl.Requeue()
	}

	logger.Info("All done!")
	return commonctrl.NoRequeue()
}

func (r *BuildStrategyReconciler) GetBuildStrategy(namespaced types.NamespacedName) (*v1beta1.ClusterBuildStrategy, error) {
	// look up storage class by name
	cls := &v1beta1.ClusterBuildStrategy{}
	if err := r.Get(context.TODO(), namespaced, cls); err != nil {
		return nil, fmt.Errorf("Unable to retrieve cls class")
	}
	return cls, nil
}
