package controllers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	o "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/shipwright-io/operator/api/v1alpha1"
	tektonoperatorv1alpha1 "github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	tektonoperatorv1alpha1client "github.com/tektoncd/operator/pkg/client/clientset/versioned/fake"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func init() {
	// exporting the environment variable which points Manifestival to the release.yaml file,
	// containing all resources managed by it
	_ = os.Setenv("KO_DATA_PATH", "../cmd/operator/kodata")
}

// bootstrapShipwrightBuildReconciler start up a new instance of ShipwrightBuildReconciler which is
// ready to interact with Manifestival, returning the Manifestival instance and the client.
func bootstrapShipwrightBuildReconciler(
	t *testing.T,
	b *v1alpha1.ShipwrightBuild,
	tcfg *tektonoperatorv1alpha1.TektonConfig,
	tcrds []*crdv1.CustomResourceDefinition,
) (client.Client, *crdclientv1.Clientset, *tektonoperatorv1alpha1client.Clientset, *ShipwrightBuildReconciler) {
	g := o.NewGomegaWithT(t)

	s := runtime.NewScheme()
	s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Namespace{})
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.Deployment{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ShipwrightBuild{})

	logger := zap.New()

	c := fake.NewFakeClientWithScheme(s, b)
	crdClient := &crdclientv1.Clientset{}
	toClient := &tektonoperatorv1alpha1client.Clientset{}
	if tcrds != nil && len(tcrds) > 0 {
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
		err := r.setupManifestival(logger)
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

// TestShipwrightBuildReconciler_ProvisionTekton tests provisioning of the Tekton Operator resources and config
func TestShipwrightBuildReconciler_ProvisionTekton(t *testing.T) {
	g := o.NewGomegaWithT(t)

	namespacedName := types.NamespacedName{Namespace: "default", Name: "name"}
	deploymentName := types.NamespacedName{
		Namespace: defaultTargetNamespace,
		Name:      "shipwright-build-controller",
	}
	req := reconcile.Request{NamespacedName: namespacedName}

	b := &v1alpha1.ShipwrightBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: v1alpha1.ShipwrightBuildSpec{
			TargetNamespace: defaultTargetNamespace,
		},
	}

	t.Logf("Deploying Shipwright Controller against '%s' namespace", defaultTargetNamespace)

	t.Run("provision tekton valid operator version, create TektonConfig object", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{"version": "v0.49.0"}
		crds := []*crdv1.CustomResourceDefinition{crd}
		c, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())

		err = c.Get(ctx, deploymentName, &appsv1.Deployment{})
		g.Expect(err).To(o.BeNil())

		tkcfg, err := r.TektonOperatorClient.TektonConfigs().Get(ctx, "config", metav1.GetOptions{})
		g.Expect(err).To(o.BeNil())
		g.Expect(tkcfg.Spec.TargetNamespace).To(o.Equal("tekton-pipelines"))
		g.Expect(tkcfg.Spec.Profile).To(o.Equal("lite"))
	})

	t.Run("provision tekton nil operator label", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crds := []*crdv1.CustomResourceDefinition{crd}
		_, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())

		_, err = r.TektonOperatorClient.TektonConfigs().Get(ctx, "config", metav1.GetOptions{})
		g.Expect(err).NotTo(o.BeNil())
	})

	t.Run("provision tekton too old operator label", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{"version": "v0.29.0"}
		crds := []*crdv1.CustomResourceDefinition{crd}
		_, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(res.Requeue).To(o.BeTrue())

		_, err = r.TektonOperatorClient.TektonConfigs().Get(ctx, "config", metav1.GetOptions{})
		g.Expect(err).NotTo(o.BeNil())
	})

	t.Run("provision tekton missing version label", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{}
		crds := []*crdv1.CustomResourceDefinition{crd}
		_, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())

		_, err = r.TektonOperatorClient.TektonConfigs().Get(ctx, "config", metav1.GetOptions{})
		g.Expect(err).NotTo(o.BeNil())
	})

	t.Run("provision tekton error on TektonConfig list", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{"version": "v0.49.0"}
		crds := []*crdv1.CustomResourceDefinition{crd}
		_, _, toClient, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)
		toClient.PrependReactor("list", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("error on list")
		})

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(err.Error()).To(o.ContainSubstring("error on list"))
		g.Expect(res.Requeue).To(o.BeTrue())

		_, err = r.TektonOperatorClient.TektonConfigs().Get(ctx, "config", metav1.GetOptions{})
		g.Expect(err).NotTo(o.BeNil())
	})

	t.Run("provision tekton error on TektonConfig create", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{"version": "v0.49.0"}
		crds := []*crdv1.CustomResourceDefinition{crd}
		_, _, toClient, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)
		toClient.PrependReactor("create", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("error on create")
		})

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(err.Error()).To(o.ContainSubstring("error on create"))
		g.Expect(res.Requeue).To(o.BeTrue())

		_, err = r.TektonOperatorClient.TektonConfigs().Get(ctx, "config", metav1.GetOptions{})
		g.Expect(err).NotTo(o.BeNil())
	})

	t.Run("provision tekton, TektonConfig object exists, error on TaskRun CRD", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{"version": "v0.49.0"}
		crds := []*crdv1.CustomResourceDefinition{crd}
		_, crdClient, toClient, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)
		toClient.PrependReactor("get", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			tcfg := &tektonoperatorv1alpha1.TektonConfig{}
			tcfg.Name = "config"
			return true, tcfg, nil
		})
		toClient.PrependReactor("list", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			tcfg := tektonoperatorv1alpha1.TektonConfig{}
			tcfg.Name = "config"
			tlist := &tektonoperatorv1alpha1.TektonConfigList{}
			tlist.Items = append(tlist.Items, tcfg)
			return true, tlist, nil
		})
		returnedTOCRD := false
		crdClient.PrependReactor("get", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			getAction, _ := action.(k8stesting.GetAction)
			if getAction.GetName() == "taskruns.tekton.dev" {
				return true, nil, fmt.Errorf("error on get")
			}
			crd := &crdv1.CustomResourceDefinition{}
			crd.Name = getAction.GetName()
			crd.Labels = map[string]string{"version": "v0.49.0"}
			returnedTOCRD = true
			return true, crd, nil
		})

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(err.Error()).To(o.ContainSubstring("error on get"))
		g.Expect(res.Requeue).To(o.BeTrue())
		g.Expect(returnedTOCRD).To(o.BeTrue())
	})

	t.Run("provision tekton, neither CRD found", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{"version": "v0.49.0"}
		_, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, nil)

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())
	})

	t.Run("provision tekton, both have errors other than not found", func(t *testing.T) {
		ctx := context.TODO()
		crd := &crdv1.CustomResourceDefinition{}
		crd.Name = "tektonconfigs.operator.tekton.dev"
		crd.Labels = map[string]string{"version": "v0.49.0"}
		_, crdClient, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, nil)
		crdClient.PrependReactor("get", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("error on all crd gets")
		})

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(o.BeNil())
		g.Expect(err.Error()).To(o.ContainSubstring("error on all crd gets"))
		g.Expect(res.Requeue).To(o.BeTrue())
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
	crd2.Labels = map[string]string{"version": "v0.49.0"}
	crds := []*crdv1.CustomResourceDefinition{crd1, crd2}
	c, _, _, r := bootstrapShipwrightBuildReconciler(t, b, nil, crds)

	t.Logf("Deploying Shipwright Controller against '%s' namespace", targetNamespace)

	// rolling out all manifests on the desired namespace, making sure the deployment for Shipwright
	// Build Controller is created accordingly
	t.Run("rollout-manifests", func(t *testing.T) {
		ctx := context.TODO()
		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())
		err = c.Get(ctx, deploymentName, &appsv1.Deployment{})
		g.Expect(err).To(o.BeNil())
		err = c.Get(ctx, namespacedName, b)
		g.Expect(err).To(o.BeNil())
		g.Expect(b.Status.IsReady()).To(o.BeTrue())
	})

	// rolling back all changes, making sure the main deployment is also not found afterwards
	t.Run("rollback-manifests", func(t *testing.T) {
		ctx := context.TODO()

		err := r.Get(ctx, namespacedName, b)
		g.Expect(err).To(o.BeNil())

		// setting a deletion timestemp on the build object, it triggers the rollback logic so the
		// reconciliation should remove the objects previously deployed
		b.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
		err = r.Update(ctx, b, &client.UpdateOptions{})
		g.Expect(err).To(o.BeNil())

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())

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
