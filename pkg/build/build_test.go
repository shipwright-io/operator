package build

import (
	"context"
	"testing"

	o "github.com/onsi/gomega"
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
		name          string
		cbsCRD        *apiextensionsv1.CustomResourceDefinition
		bsCRD         *apiextensionsv1.CustomResourceDefinition
		webhookpod    *corev1.Pod
		expectError   bool
		expectRequeue bool
	}{
		{
			name:          "No build strategies crds",
			cbsCRD:        &apiextensionsv1.CustomResourceDefinition{},
			webhookpod:    &corev1.Pod{},
			expectError:   true,
			expectRequeue: true,
		},
		{
			name: "cluster build strategy defined but not build strategie crd",
			cbsCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "clusterbuildstrategies.shipwright.io",
				},
			},
			bsCRD:         &apiextensionsv1.CustomResourceDefinition{},
			webhookpod:    &corev1.Pod{},
			expectError:   true,
			expectRequeue: true,
		},
		{
			name: "cluster build strategy and build strategy defined but webhook pod not ready",
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
			webhookpod:    &corev1.Pod{},
			expectError:   true,
			expectRequeue: true,
		},
		{
			name: "cluster build strategy and build strategy defined and webhook pods is ready",
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
			webhookpod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "shipwright-build",
					Name:      "shipwright-build-webhook",
					Labels: map[string]string{
						"name": "shp-build-webhook",
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{{
						Type:   corev1.PodReady,
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
			t.Setenv("TIMEOUT", "30s")
			crdClient := apiextensionsfake.NewSimpleClientset(crds...)
			c := fake.NewClientBuilder().WithObjects(tc.webhookpod).Build()
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
