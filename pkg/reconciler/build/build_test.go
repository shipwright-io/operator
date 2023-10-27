package build

import (
	"context"
	"testing"

	o "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestReconcileBuildStrategy(t *testing.T) {
	cases := []struct {
		name              string
		cbsCRD            *apiextensionsv1.CustomResourceDefinition
		bsCRD             *apiextensionsv1.CustomResourceDefinition
		webhookdeployment *appsv1.Deployment
		expectError       bool
		expectRequeue     bool
	}{
		{
			name:              "No build strategies crds",
			cbsCRD:            &apiextensionsv1.CustomResourceDefinition{},
			webhookdeployment: &appsv1.Deployment{},
			expectError:       true,
			expectRequeue:     true,
		},
		{
			name: "cluster build strategy defined but not build strategie crd",
			cbsCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "clusterbuildstrategies.shipwright.io",
				},
			},
			bsCRD:             &apiextensionsv1.CustomResourceDefinition{},
			webhookdeployment: &appsv1.Deployment{},
			expectError:       true,
			expectRequeue:     true,
		},
		{
			name: "cluster build strategy and build strategy defined but webhook deployment not available",
			cbsCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "clusterbuildstrategies.shipwright.io",
				},
			},
			bsCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "buildstrategies.shipwright.io",
				},
			},
			webhookdeployment: &appsv1.Deployment{},
			expectError:       true,
			expectRequeue:     true,
		},
		{
			name: "cluster build strategy and build strategy defined and webhook deployment is available",
			cbsCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "clusterbuildstrategies.shipwright.io",
				},
			},
			bsCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "buildstrategies.shipwright.io",
				},
			},
			webhookdeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: "shipwright-build", Name: "shipwright-build-webhook"},
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{{
						Type:   appsv1.DeploymentAvailable,
						Status: corev1.ConditionTrue,
					}},
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
			if tc.cbsCRD != nil {
				crds = append(crds, tc.cbsCRD)
			}
			if tc.bsCRD != nil {
				crds = append(crds, tc.bsCRD)
			}
			t.Setenv("KO_DATA_PATH", "./testdata")
			crdClient := apiextensionsfake.NewSimpleClientset(crds...)
			c := fake.NewClientBuilder().WithObjects(tc.webhookdeployment).WithStatusSubresource(tc.webhookdeployment).Build()
			requeue, err := ReconcileBuildStrategy(ctx, crdClient.ApiextensionsV1(), c, zap.New(), "shipwright-build")
			if tc.expectError {
				g.Expect(err).To(o.HaveOccurred())
			} else {
				g.Expect(err).NotTo(o.HaveOccurred())
			}
			g.Expect(requeue).To(o.Equal(tc.expectRequeue))
		})
	}

}
