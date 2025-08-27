package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/shipwright-io/operator/pkg/common"

	o "github.com/onsi/gomega"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/operator/api/v1alpha1"
	tektonoperatorv1alpha1 "github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	tektonoperatorv1alpha1client "github.com/tektoncd/operator/pkg/client/clientset/versioned/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// bootstrapShipwrightBuildReconciler start up a new instance of ShipwrightBuildReconciler which is
// ready to interact with Manifestival, returning the Manifestival instance and the client.
func bootstrapShipwrightBuildReconciler(
	t *testing.T,
	b *v1alpha1.ShipwrightBuild,
	tcfg *tektonoperatorv1alpha1.TektonConfig,
	tcrds []*crdv1.CustomResourceDefinition,
	statusObjects ...client.Object,
) (client.Client, *crdclientv1.Clientset, *tektonoperatorv1alpha1client.Clientset, *ShipwrightBuildReconciler) {
	g := o.NewGomegaWithT(t)

	s := runtime.NewScheme()
	s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Namespace{})
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.Deployment{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ShipwrightBuild{})
	s.AddKnownTypes(rbacv1.SchemeGroupVersion, &rbacv1.ClusterRoleBinding{})
	s.AddKnownTypes(rbacv1.SchemeGroupVersion, &rbacv1.ClusterRole{})
	s.AddKnownTypes(tektonoperatorv1alpha1.SchemeGroupVersion, &tektonoperatorv1alpha1.TektonConfig{})
	s.AddKnownTypes(buildv1alpha1.SchemeGroupVersion, &buildv1alpha1.ClusterBuildStrategy{})

	logger := zap.New()
	clientBuilder := fake.NewClientBuilder().WithScheme(s).WithObjects(b)
	if len(statusObjects) > 0 {
		// the fake client does not support the status subresource by default.
		clientBuilder = clientBuilder.WithStatusSubresource(statusObjects...)
	}
	c := clientBuilder.Build()

	var crdClient *crdclientv1.Clientset
	var toClient *tektonoperatorv1alpha1client.Clientset
	if len(tcrds) > 0 {
		objs := []runtime.Object{}
		for _, obj := range tcrds {
			objs = append(objs, obj)
		}
		crdClient = crdclientv1.NewSimpleClientset(objs...)
	} else {
		crdClient = crdclientv1.NewSimpleClientset()
	}
	if tcfg == nil {
		toClient = tektonoperatorv1alpha1client.NewSimpleClientset()
	} else {
		toClient = tektonoperatorv1alpha1client.NewSimpleClientset(tcfg)
	}
	r := &ShipwrightBuildReconciler{CRDClient: crdClient.ApiextensionsV1(), TektonOperatorClient: toClient.OperatorV1alpha1(), Client: c, Scheme: s, Logger: logger}

	// creating targetNamespace on which Shipwright-Build will be deployed against, before the other
	// tests takes place
	if b.Spec.TargetNamespace != "" {
		t.Logf("Creating test namespace '%s'", b.Spec.TargetNamespace)
		t.Run("create-test-namespace", func(t *testing.T) {
			err := c.Create(
				context.TODO(),
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: b.Spec.TargetNamespace}},
				&client.CreateOptions{},
			)
			g.Expect(err).To(o.BeNil())
		})
	}

	// manifestival instance is setup as part of controller-=runtime's SetupWithManager, thus calling
	// the setup before all other methods
	t.Run("setupManifestival", func(t *testing.T) {
		err := r.setupManifestival()
		g.Expect(err).To(o.BeNil())
	})

	return c, crdClient, toClient, r
}

// TestShipwrightBuildReconciler_Finalizers testing adding and removing finalizers on the resource.
func TestShipwrightBuildReconciler_Finalizers(t *testing.T) {
	g := o.NewGomegaWithT(t)

	b := &v1alpha1.ShipwrightBuild{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "default"}}
	_, _, _, r := bootstrapShipwrightBuildReconciler(t, b, &tektonoperatorv1alpha1.TektonConfig{}, []*crdv1.CustomResourceDefinition{})

	// adding one entry on finalizers slice, making sure it's registered
	t.Run("setFinalizer", func(t *testing.T) {
		err := r.setFinalizer(context.TODO(), b)

		g.Expect(err).To(o.BeNil())
		g.Expect(b.GetFinalizers()).To(o.Equal([]string{FinalizerAnnotation}))
	})

	// removing previously added finalizer entry, making sure slice it's empty afterwards
	t.Run("unsetFinalizer", func(t *testing.T) {
		err := r.unsetFinalizer(context.TODO(), b)

		g.Expect(err).To(o.BeNil())
		g.Expect(b.GetFinalizers()).To(o.Equal([]string{}))
	})
}

// testShipwrightBuildReconcilerReconcile simulates the reconciliation process for rolling out and
// rolling back manifests in the informed target namespace name.
func testShipwrightBuildReconcilerReconcile(t *testing.T, targetNamespace string) {
	g := o.NewGomegaWithT(t)

	namespacedName := types.NamespacedName{Namespace: "default", Name: "name"}
	deploymentName := types.NamespacedName{
		Namespace: targetNamespace,
		Name:      "shipwright-build-controller",
	}
	req := reconcile.Request{NamespacedName: namespacedName}

	b := &v1alpha1.ShipwrightBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: v1alpha1.ShipwrightBuildSpec{
			TargetNamespace: targetNamespace,
		},
	}
	crd1 := &crdv1.CustomResourceDefinition{}
	crd1.Name = "taskruns.tekton.dev"
	crd2 := &crdv1.CustomResourceDefinition{}
	crd2.Name = "tektonconfigs.operator.tekton.dev"
	crd2.Labels = map[string]string{"operator.tekton.dev/release": common.TektonOpMinSupportedVersion}
	crds := []*crdv1.CustomResourceDefinition{crd1, crd2}
	c, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)

	images := []struct {
		key, value string
	}{
		{"IMAGE_SHIPWRIGHT_SHIPWRIGHT_BUILD", "ghcr.io/shipwright-io/build/shipwright-build-controller:nightly-2023-05-05-1683263383"},
		{"IMAGE_SHIPWRIGHT_GIT_CONTAINER_IMAGE", "ghcr.io/shipwright-io/build/git:nightly-2023-05-02-1683004171"},
		{"IMAGE_SHIPWRIGHT_WAITER_CONTAINER_IMAGE", "ghcr.io/shipwright-io/build/waiter:nightly-2023-05-05-1683263383"},
		{"IMAGE_SHIPWRIGHT_MUTATE_IMAGE_CONTAINER_IMAGE", "ghcr.io/shipwright-io/build/mutate-image:nightly-2023-04-18-1681794585"},
		{"IMAGE_SHIPWRIGHT_BUNDLE_CONTAINER_IMAGE", "ghcr.io/shipwright-io/build/bundle:nightly-2023-05-05-1683263383"},
	}

	t.Logf("Deploying Shipwright Controller against '%s' namespace", targetNamespace)

	// rolling out all manifests on the desired namespace, making sure the deployment for Shipwright
	// Build Controller is created accordingly
	t.Run("rollout-manifests", func(t *testing.T) {
		t.Skip("k8s 1.27 upgrade led to test clients returning false not found errors")
		ctx := context.TODO()
		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		// TODO: Code technically uses two different clientsets that don't talk to each other.
		// This makes testing brittle and unable to capture the behavior on a real cluster.
		// Requeue can return "true" because the tests think the CRD for ClusterBuildStrategies
		// do not exist yet.
		g.Expect(res.RequeueAfter).NotTo(o.BeZero(), "checking requeue for Reconcile")
		err = c.Get(ctx, deploymentName, &appsv1.Deployment{})
		g.Expect(err).To(o.BeNil())
		err = c.Get(ctx, namespacedName, b)
		g.Expect(err).To(o.BeNil())
		// Likewise, the ShipwrightBuild object will not report itself ready because it is waiting
		// for the ClusterBuildStrategy CRD to be created first.
		g.Expect(b.Status.IsReady()).To(o.BeFalse(), "checking ShipwrightBuild readiness")
	})

	t.Run("rollout-manifests-with-images-env-vars", func(t *testing.T) {
		t.Skip("k8s 1.27 upgrade led to test clients returning false not found errors")
		ctx := context.TODO()
		for _, v := range images {
			t.Setenv(v.key, v.value)
		}
		deployment := &appsv1.Deployment{}
		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		// TODO: Code technically uses two different clientsets that don't talk to each other.
		// This makes testing brittle and unable to capture the behavior on a real cluster.
		// Requeue can return "true" because the tests think the CRD for ClusterBuildStrategies
		// do not exist yet.
		g.Expect(res.RequeueAfter).NotTo(o.BeZero())
		err = c.Get(ctx, deploymentName, deployment)
		g.Expect(err).To(o.BeNil())
		containers := deployment.Spec.Template.Spec.Containers
		g.Expect(containers[0].Image).To(o.Equal("ghcr.io/shipwright-io/build/shipwright-build-controller:nightly-2023-05-05-1683263383"))
		err = c.Get(ctx, namespacedName, b)
		g.Expect(err).To(o.BeNil())
		// Likewise, the ShipwrightBuild object will not report itself ready because it is waiting
		// for the ClusterBuildStrategy CRD to be created first.
		g.Expect(b.Status.IsReady()).To(o.BeFalse())
	})

	// rolling back all changes, making sure the main deployment is also not found afterwards
	t.Run("rollback-manifests", func(t *testing.T) {
		t.Skip("k8s 1.27 upgrade led to test clients returning false not found errors")
		ctx := context.TODO()

		err := r.Get(ctx, namespacedName, b)
		g.Expect(err).To(o.BeNil())

		// setting a deletion timestemp on the build object, it triggers the rollback logic so the
		// reconciliation should remove the objects previously deployed

		// TODO: Refactor to use owner references so the rollback is handled by Kubernetes itself.
		b.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
		err = r.Update(ctx, b, &client.UpdateOptions{})
		g.Expect(err).To(o.BeNil())

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		g.Expect(res.RequeueAfter).To(o.BeZero())

		err = c.Get(ctx, deploymentName, &appsv1.Deployment{})
		g.Expect(errors.IsNotFound(err)).To(o.BeTrue())
	})
}

// TestShipwrightBuildReconciler_Reconcile runs rollout/rollback tests against different namespaces.
func TestShipwrightBuildReconciler_Reconcile(t *testing.T) {
	tests := []struct {
		testName        string
		targetNamespace string
	}{{
		testName:        "target namespace is informed",
		targetNamespace: "namespace",
	}, {
		testName:        "target namespace is not informed",
		targetNamespace: defaultTargetNamespace,
	}}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testShipwrightBuildReconcilerReconcile(t, tt.targetNamespace)
		})
	}
}

func TestShipwrightBuildReconciler_OperandReadiness(t *testing.T) {
	g := o.NewGomegaWithT(t)
	ctx := context.TODO()

	// ShipwrightBuild object
	b := &v1alpha1.ShipwrightBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "default",
		},
		Spec: v1alpha1.ShipwrightBuildSpec{
			TargetNamespace: "namespace",
		},
		Status: v1alpha1.ShipwrightBuildStatus{
			Conditions: []metav1.Condition{
				{
					Type:   ConditionReady,
					Status: metav1.ConditionTrue,
				},
			},
		},
	}

	// Mock TektonConfig to simulate Tekton Operator installation
	tektonConfig := &tektonoperatorv1alpha1.TektonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "config",
		},
		Status: tektonoperatorv1alpha1.TektonConfigStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					{
						Type:    ConditionReady,
						Status:  corev1.ConditionFalse,
						Reason:  "Installed",
						Message: "TektonConfig is not ready",
					},
				},
			},
		},
	}
	// Preload the fake client with a valid ClusterBuildStrategy to pass buildstrategy reconciliation
	cbs := &buildv1alpha1.ClusterBuildStrategy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "buildah",
		},
		Spec: buildv1alpha1.BuildStrategySpec{
			BuildSteps: []buildv1alpha1.BuildStep{
				{
					Container: corev1.Container{
						Name:  "build-step",
						Image: "quay.io/buildah/stable",
					},
				},
			},
		},
	}

	// Prepare cr
	crd1 := &crdv1.CustomResourceDefinition{}
	crd1.Name = "taskruns.tekton.dev"
	crd2 := &crdv1.CustomResourceDefinition{}
	crd2.Name = "tektonconfigs.operator.tekton.dev"
	crd2.Labels = map[string]string{"operator.tekton.dev/release": common.TektonOpMinSupportedVersion}
	crd3 := &crdv1.CustomResourceDefinition{}
	crd3.Name = "clusterbuildstrategies.shipwright.io"
	crds := []*crdv1.CustomResourceDefinition{crd1, crd2, crd3}

	// Bootstrap the reconciler with the mock objects
	c, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds, &v1alpha1.ShipwrightBuild{})

	// Inject a pre-created valid tektonconfig
	_, err := r.TektonOperatorClient.TektonConfigs().Create(ctx, tektonConfig, metav1.CreateOptions{})
	g.Expect(err).To(o.BeNil())

	// Verify creation of tektonconfig
	cfg, err := r.TektonOperatorClient.TektonConfigs().Get(ctx, "config", metav1.GetOptions{})
	g.Expect(err).To(o.BeNil())
	g.Expect(cfg.Name).To(o.Equal("config"))
	g.Expect(cfg.Status.Conditions[0].Type).To(o.Equal(apis.ConditionReady))

	// Simulate reconciliation
	namespacedName := types.NamespacedName{Namespace: "default", Name: "name"}
	req := reconcile.Request{NamespacedName: namespacedName}
	res, err := r.Reconcile(ctx, req)
	g.Expect(err).To(o.BeNil())
	g.Expect(res.RequeueAfter).NotTo(o.BeZero(), "Reconciliation should requeue when TektonConfig is not ready")

	// Verify that the ShipwrightBuild is marked as not ready
	updated := &v1alpha1.ShipwrightBuild{}
	err = c.Get(ctx, req.NamespacedName, updated)
	g.Expect(err).To(o.BeNil())
	g.Expect(updated.Status.IsReady()).To(o.BeFalse(), "ShipwrightBuild should not be ready when TektonConfig is not ready")

	// Simulate TektonConfig becoming ready
	tektonConfig.Status.Conditions[0].Status = corev1.ConditionTrue
	tektonConfig.Status.Conditions[0].Reason = "Installed"
	tektonConfig.Status.Conditions[0].Message = "TektonConfig is now ready"
	_, err = r.TektonOperatorClient.TektonConfigs().Update(ctx, tektonConfig, metav1.UpdateOptions{})
	g.Expect(err).To(o.BeNil())

	// Inject a pre-created valid ClusterBuildStrategy
	err = c.Create(ctx, cbs)
	g.Expect(err).To(o.BeNil())
	// Trigger reconciliation again
	res, err = r.Reconcile(ctx, req)
	g.Expect(err).To(o.BeNil())
	g.Expect(res.RequeueAfter).To(o.BeZero(), "Should not requeue after TektonConfig is ready")

	// Fetch and verify ShipwrightBuild is now ready
	err = c.Get(ctx, req.NamespacedName, updated)
	g.Expect(err).To(o.BeNil())
	g.Expect(updated.Status.IsReady()).To(o.BeTrue(), "ShipwrightBuild should be ready when TektonConfig is ready")
}
