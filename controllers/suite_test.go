// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	tektonoperatorv1alpha1client "github.com/tektoncd/operator/pkg/client/clientset/versioned/typed/operator/v1alpha1"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	operatorv1alpha1 "github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/pkg/common"
	"github.com/shipwright-io/operator/test"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient   client.Client
	ctx         context.Context
	cancel      context.CancelFunc
	testEnv     *envtest.Environment
	restTimeout = 5 * time.Second
	restRetry   = 100 * time.Millisecond
)

func TestAPIs(t *testing.T) {
	skip := os.Getenv("SKIP_ENVTEST")
	shouldSkip, _ := strconv.ParseBool(skip)
	if shouldSkip {
		t.Skip("bypassing Ginkgo-driven tests")
	}
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(restTimeout)
	SetDefaultEventuallyPollingInterval(restRetry)
	RunSpecs(t, "Controller Suite")
}

// setupTektonCRDs mocks out the CRD definition for Tekton TaskRuns and TektonConfig
func setupTektonCRDs(ctx context.Context) {
	truePtr := true
	By("does tekton taskrun crd exist")
	err := k8sClient.Get(ctx, types.NamespacedName{Name: "taskruns.tekton.dev"}, &crdv1.CustomResourceDefinition{})
	if errors.IsNotFound(err) {
		By("creating tekton taskrun crd")
		taskRunCRD := &crdv1.CustomResourceDefinition{}
		taskRunCRD.Name = "taskruns.tekton.dev"
		taskRunCRD.Spec.Group = "tekton.dev"
		taskRunCRD.Spec.Scope = crdv1.NamespaceScoped
		taskRunCRD.Spec.Versions = []crdv1.CustomResourceDefinitionVersion{
			{
				Name:    "v1beta1",
				Storage: true,
				Schema: &crdv1.CustomResourceValidation{
					OpenAPIV3Schema: &crdv1.JSONSchemaProps{
						Type:                   "object",
						XPreserveUnknownFields: &truePtr,
					},
				},
			},
		}
		taskRunCRD.Spec.Names.Plural = "taskruns"
		taskRunCRD.Spec.Names.Singular = "taskrun"
		taskRunCRD.Spec.Names.Kind = "TaskRun"
		taskRunCRD.Spec.Names.ListKind = "TaskRunList"
		taskRunCRD.Status.StoredVersions = []string{"v1beta1"}
		err = k8sClient.Create(ctx, taskRunCRD, &client.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

	}
	Expect(err).NotTo(HaveOccurred())

	By("does tektonconfig crd exist")
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "tektonconfigs.operator.tekton.dev"}, &crdv1.CustomResourceDefinition{})
	if errors.IsNotFound(err) {
		tektonOpCRD := &crdv1.CustomResourceDefinition{}
		tektonOpCRD.Name = "tektonconfigs.operator.tekton.dev"
		tektonOpCRD.Labels = map[string]string{"operator.tekton.dev/release": common.TektonOpMinSupportedVersion}
		tektonOpCRD.Spec.Group = "operator.tekton.dev"
		tektonOpCRD.Spec.Scope = crdv1.ClusterScoped
		tektonOpCRD.Spec.Versions = []crdv1.CustomResourceDefinitionVersion{
			{
				Name:    "v1alpha1",
				Storage: true,
				Schema: &crdv1.CustomResourceValidation{
					OpenAPIV3Schema: &crdv1.JSONSchemaProps{
						Type:                   "object",
						XPreserveUnknownFields: &truePtr,
					},
				},
			},
		}
		tektonOpCRD.Spec.Names.Plural = "tektonconfigs"
		tektonOpCRD.Spec.Names.Singular = "tektonconfig"
		tektonOpCRD.Spec.Names.Kind = "TektonConfig"
		tektonOpCRD.Spec.Names.ListKind = "TektonConfigList"
		tektonOpCRD.Status.StoredVersions = []string{"v1alpha1"}
		err = k8sClient.Create(ctx, tektonOpCRD, &client.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	}
	Expect(err).NotTo(HaveOccurred())
}

// createShipwrightBuild creates an instance of the ShipwrightBuild object with the given target
// namespace.
func createShipwrightBuild(ctx context.Context, targetNamespace string) *operatorv1alpha1.ShipwrightBuild {
	By("creating a ShipwrightBuild instance")
	build := &operatorv1alpha1.ShipwrightBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: operatorv1alpha1.ShipwrightBuildSpec{
			TargetNamespace: targetNamespace,
		},
	}
	err := k8sClient.Create(ctx, build, &client.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	// when the finalizer is in place, the deployment of manifest elements is done, and therefore
	// functional testing can proceed
	By("waiting for the finalizer to be set")
	test.EventuallyContainFinalizer(ctx, k8sClient, build, FinalizerAnnotation)
	return build
}

// deleteShipwrightBuild tears down the given ShipwrightBuild instance.
func deleteShipwrightBuild(ctx context.Context, build *operatorv1alpha1.ShipwrightBuild) {
	By("deleting the ShipwrightBuild instance")
	namespacedName := types.NamespacedName{Name: build.Name}
	err := k8sClient.Get(ctx, namespacedName, build)
	if errors.IsNotFound(err) {
		return
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Delete(ctx, build, &client.DeleteOptions{})
	// the delete e2e's can delete this object before this AfterEach runs
	if errors.IsNotFound(err) {
		return
	}
	Expect(err).NotTo(HaveOccurred())

	By("waiting for ShipwrightBuild instance to be completely removed")
	test.EventuallyRemoved(ctx, k8sClient, build)
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = operatorv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = crdv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = buildv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).NotTo(HaveOccurred())
	crdClient, err := crdclientv1.NewForConfig(mgr.GetConfig())
	Expect(err).NotTo(HaveOccurred())
	toClient, err := tektonoperatorv1alpha1client.NewForConfig(mgr.GetConfig())
	Expect(err).NotTo(HaveOccurred())
	err = (&ShipwrightBuildReconciler{
		CRDClient:            crdClient,
		TektonOperatorClient: toClient,
		Client:               mgr.GetClient(),
		Scheme:               scheme.Scheme,
		Logger:               ctrl.Log.WithName("controllers").WithName("shipwrightbuild"),
	}).SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		// The controller needs its own context to run in a separate goroutine
		// This needs to be lifecycled independently of the context that ginkgo/v2 passes in
		ctx, cancel = context.WithCancel(context.Background())
		err := mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
