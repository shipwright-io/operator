// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/shipwright-io/build/pkg/ctxlog"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	operatorv1alpha1 "github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

// variables to hold command-line flags
var (
	// metricsAddr listen address for the prometheus metrics endpoint.
	metricsAddr string
	// probeAddr list address for health-probe endpoint.
	probeAddr string
	// webhookPort port number on which webhook listens to.
	webhookPort int
	// enableLeaderElection enables leader election process, for high-available deployments.
	enableLeaderElection bool
)

func init() {
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	flag.IntVar(&webhookPort, "webhook-port", 9443,
		"Webhook server listen port.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	l := ctxlog.NewLogger("shipwrightBuild-controller")

	ctx := ctxlog.NewParentContext(l)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   webhookPort,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "01a9b2d1.shipwright.io",
	})
	if err != nil {
		ctxlog.Error(ctx, err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ShipwrightBuildReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctxlog.Error(ctx, err, "unable to create controller")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		ctxlog.Error(ctx, err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		ctxlog.Error(ctx, err, "unable to set up ready check")
		os.Exit(1)
	}

	ctxlog.Info(ctx, "starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctxlog.Error(ctx, err, "problem running manager")
		os.Exit(1)
	}
}
