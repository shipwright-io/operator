// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	mfc "github.com/manifestival/controller-runtime-client"
	"github.com/manifestival/manifestival"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shipwright-io/operator/api/v1alpha1"
)

// ShipwrightBuildReconciler reconciles a ShipwrightBuild object
type ShipwrightBuildReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Manifest manifestival.Manifest
}

// Declare RBAC needed to reconcile the release manifest YAML
// To minimize the risk of privilege escalation or destructive behavior, the controller is only
// allowed to modify named resources that deploy Shipwright Build.
// This is especially true for the cluster roles and custom resource definitions included in the
// release manifest.

// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/finalizers,verbs=update
// +kubebuilder:rbac:groups=shipwright.io,resources=*,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=core,resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=apps,resourceNames=shipwright-build,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,resourceNames=builds.shipwright.io;buildruns.shipwright.io;buildstrategies.shipwright.io;clusterbuildstrategies.shipwright.io,verbs=update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=tekton.dev,resources=tasks;taskruns,verbs=create;delete;get;list;patch;update;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ShipwrightBuild object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ShipwrightBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("shipwrightbuild", req.NamespacedName)

	build := &v1alpha1.ShipwrightBuild{}
	// Remove Namespaces from the manifest - cluster admins must provision the shipwright-build namespace
	manifest := r.Manifest.Filter(manifestival.Not(manifestival.ByKind("Namespace")))
	err := r.Client.Get(ctx, req.NamespacedName, build)
	if errors.IsNotFound(err) {
		log.Info("object not found, deleting Shipwright Build from the cluster")
		err = manifest.Delete()
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("reconciling ShipwrightBuild with manifest")
	err = manifest.Apply()
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ShipwrightBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mfclient := mfc.NewClient(mgr.GetClient())
	mflogger := mgr.GetLogger().WithName("manifestival")
	dataPath, exists := os.LookupEnv("KO_DATA_PATH")
	if !exists {
		return fmt.Errorf("KO_DATA_PATH is not set - cannot set up reconciler")
	}
	buildManifest := filepath.Join(dataPath, "release.yaml")

	mf, err := manifestival.NewManifest(buildManifest, manifestival.UseClient(mfclient), manifestival.UseLogger(mflogger))
	if err != nil {
		return err
	}
	r.Manifest = mf
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ShipwrightBuild{}).
		Complete(r)
}
