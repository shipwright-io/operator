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

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	tektonoperatorv1alpha1client "github.com/tektoncd/operator/pkg/client/clientset/versioned/typed/operator/v1alpha1"

	operatorv1alpha1 "github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
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

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: webhookPort,
		}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "01a9b2d1.shipwright.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	crdClient, err := crdclientv1.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to get crd client")
		os.Exit(1)
	}
	tektonOperatorClient, err := tektonoperatorv1alpha1client.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to get tekton operator client")
		os.Exit(1)
	}

	if err = (&controllers.ShipwrightBuildReconciler{
		CRDClient:            crdClient,
		TektonOperatorClient: tektonOperatorClient,
		Client:               mgr.GetClient(),
		Scheme:               mgr.GetScheme(),
		Logger:               ctrl.Log.WithName("controllers").WithName("ShipwrightBuild"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ShipwrightBuild")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
