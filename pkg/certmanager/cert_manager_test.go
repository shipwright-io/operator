package certmanager

import (
	"context"
	"testing"

	o "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestReconcileCertManager(t *testing.T) {
	cases := []struct {
		name            string
		certificatesCRD *apiextensionsv1.CustomResourceDefinition
		certmanagersCRD *apiextensionsv1.CustomResourceDefinition
		expectError     bool
		expectRequeue   bool
	}{
		{
			name:        "No cert-manager Objects",
			expectError: true,
		},
		{
			name: "cert-manager certificates defined",
			certificatesCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "certificates.cert-manager.io",
				},
			},
			expectError:   false,
			expectRequeue: false,
		},
		{
			name: "cert-manager crd defined",
			certmanagersCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "certmanagers.operator.openshift.io",
				},
			},
			expectError:   false,
			expectRequeue: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := o.NewWithT(t)
			ctx := context.TODO()
			crds := []runtime.Object{}
			if tc.certificatesCRD != nil {
				crds = append(crds, tc.certificatesCRD)
			}
			if tc.certmanagersCRD != nil {
				crds = append(crds, tc.certmanagersCRD)
			}
			crdClient := apiextensionsfake.NewSimpleClientset(crds...)
			c := fake.NewClientBuilder().Build()
			requeue, err := ReconcileCertManager(ctx, crdClient.ApiextensionsV1(), c, zap.New(), "shipwright-build")
			if tc.expectError {
				g.Expect(err).To(o.HaveOccurred())
			} else {
				g.Expect(err).NotTo(o.HaveOccurred())
			}
			g.Expect(requeue).To(o.Equal(tc.expectRequeue))
		})
	}

}
