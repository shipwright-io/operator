package certificates

import (
	"context"
	"testing"

	certmanager "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	o "github.com/onsi/gomega"
	"github.com/shipwright-io/operator/api/v1alpha1"
	commonctrl "github.com/shipwright-io/operator/controllers/common"
	corev1 "k8s.io/api/core/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// bootstrapShipwrightBuildReconciler start up a new instance of ShipwrightBuildReconciler which is
// ready to interact with Manifestival, returning the Manifestival instance and the client.
func bootstrapCertificatesReconciler(
	t *testing.T,
	b *v1alpha1.ShipwrightBuild,
	tcrds []*crdv1.CustomResourceDefinition,
) (client.Client, *crdclientv1.Clientset, *CertificatesReconciler) {
	g := o.NewGomegaWithT(t)

	s := runtime.NewScheme()
	s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Namespace{})
	s.AddKnownTypes(certmanager.SchemeGroupVersion, &certmanager.Certificate{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ShipwrightBuild{})

	logger := zap.New()

	c := fake.NewClientBuilder().WithScheme(s).WithObjects(b).WithStatusSubresource(b).Build()

	var crdClient *crdclientv1.Clientset
	if len(tcrds) > 0 {
		objs := []runtime.Object{}
		for _, obj := range tcrds {
			objs = append(objs, obj)
		}
		crdClient = crdclientv1.NewSimpleClientset(objs...)
	} else {
		crdClient = crdclientv1.NewSimpleClientset()
	}

	r := &CertificatesReconciler{CRDClient: crdClient.ApiextensionsV1(), Client: c, Scheme: s, Logger: logger}

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

	return c, crdClient, r
}

// testCertificatesReconcilerReconcile simulates the reconciliation process for rolling out and
// rolling back manifests in the informed target namespace name.
func testCertificatesReconcilerReconcile(t *testing.T, targetNamespace string) {
	g := o.NewGomegaWithT(t)

	namespacedName := types.NamespacedName{Namespace: "default", Name: "name"}
	certificateName := types.NamespacedName{
		Namespace: targetNamespace,
		Name:      "shipwright-build-webhook-cert",
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
	crd1.Name = "certificates.cert-manager.io"
	crd2 := &crdv1.CustomResourceDefinition{}
	crd2.Name = "certmanagers.operator.openshift.io"
	crds := []*crdv1.CustomResourceDefinition{crd1, crd2}
	_, _, r := bootstrapCertificatesReconciler(t, b, crds)

	t.Logf("Deploying Certificates Controller against '%s' namespace", targetNamespace)
	t.Setenv(commonctrl.UseManagedWebhookCerts, "true")

	// rolling out all manifests on the desired namespace, making sure the webhook certificate is created
	t.Run("rollout-manifests", func(t *testing.T) {
		ctx := context.TODO()
		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())
		err = r.Get(ctx, certificateName, &certmanager.Certificate{})
		g.Expect(err).To(o.BeNil())
	})

	// rolling back all changes, making sure the webhook certificate is also not found afterwards
	t.Run("rollback-manifests", func(t *testing.T) {
		ctx := context.TODO()

		err := r.Get(ctx, namespacedName, b)
		g.Expect(err).To(o.BeNil())

		/*b.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
		err = r.Update(ctx, b, &client.UpdateOptions{Raw: &metav1.UpdateOptions{}})
		g.Expect(err).To(o.BeNil())

		res, err := r.Reconcile(ctx, req)
		g.Expect(err).To(o.BeNil())
		g.Expect(res.Requeue).To(o.BeFalse())

		err = r.Get(ctx, certificateName, &certmanager.Certificate{})
		g.Expect(err).To(o.BeNil())*/
	})
}

// TestShipwrightBuildReconciler_Reconcile runs rollout/rollback tests against different namespaces.
func TestCertificatesReconciler_Reconcile(t *testing.T) {
	tests := []struct {
		testName        string
		targetNamespace string
	}{{
		testName:        "target namespace is informed",
		targetNamespace: "namespace",
	}, {
		testName:        "target namespace is not informed",
		targetNamespace: commonctrl.DefaultTargetNamespace,
	}}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testCertificatesReconcilerReconcile(t, tt.targetNamespace)
		})
	}
}
