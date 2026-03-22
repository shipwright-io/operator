package triggers

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/shipwright-io/operator/pkg/common"
)

func TestReconcileTriggers(t *testing.T) {
	cases := []struct {
		name                     string
		installCRDs              bool
		expectRequeue            bool
		expectResourcesInstalled bool
	}{
		{
			name:          "no builds CRD",
			installCRDs:   false,
			expectRequeue: true,
		},
		{
			name:                     "builds CRD exists",
			installCRDs:              true,
			expectRequeue:            false,
			expectResourcesInstalled: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := NewWithT(t)
			ctx := context.Background()

			objects := []runtime.Object{}
			if tc.installCRDs {
				objects = append(objects, &crdv1.CustomResourceDefinition{
					ObjectMeta: v1.ObjectMeta{
						Name: buildsCRD,
					},
				})
			}
			crdClient := apiextensionsfake.NewSimpleClientset(objects...)

			k8sScheme := runtime.NewScheme()
			err := scheme.AddToScheme(k8sScheme)
			o.Expect(err).NotTo(HaveOccurred())

			k8sClient := fake.NewClientBuilder().WithScheme(k8sScheme).Build()
			log := zap.New()

			manifests, err := common.SetupManifestival(k8sClient, "triggers-release.yaml", false, log)
			o.Expect(err).NotTo(HaveOccurred(), "setting up Manifestival")

			requeue, err := ReconcileTriggers(ctx, crdClient.ApiextensionsV1(), log, manifests)
			o.Expect(err).NotTo(HaveOccurred(), "reconciling triggers")
			o.Expect(requeue).To(BeEquivalentTo(tc.expectRequeue), "check reconcile requeue")

			if tc.expectResourcesInstalled {
				deployment := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "shipwright-build", Name: "shipwright-triggers"}, deployment)
				o.Expect(err).NotTo(HaveOccurred(), "triggers deployment should exist")

				service := &corev1.Service{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: "shipwright-build", Name: "shipwright-triggers"}, service)
				o.Expect(err).NotTo(HaveOccurred(), "triggers service should exist")

				sa := &corev1.ServiceAccount{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: "shipwright-build", Name: "shipwright-triggers"}, sa)
				o.Expect(err).NotTo(HaveOccurred(), "triggers service account should exist")
			}
		})
	}
}
