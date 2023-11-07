package buildstrategy

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/operator/pkg/common"
	"github.com/shipwright-io/operator/test"
)

func TestReconcileBuildStrategies(t *testing.T) {

	cases := []struct {
		name                      string
		installShipwrightCRDs     bool
		expectRequeue             bool
		expectStrategiesInstalled bool
	}{
		{
			name:                  "no Shipwright CRDs",
			installShipwrightCRDs: false,
			expectRequeue:         true,
		},
		{
			name:                      "install Shipwright CRDs",
			installShipwrightCRDs:     true,
			expectRequeue:             false,
			expectStrategiesInstalled: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := NewWithT(t)
			ctx := context.Background()
			var cancel context.CancelFunc
			deadline, hasDeadline := t.Deadline()
			if hasDeadline {
				ctx, cancel = context.WithDeadline(context.Background(), deadline)
				defer cancel()
			}
			objects := []runtime.Object{}
			if tc.installShipwrightCRDs {
				objects = append(objects, &crdv1.CustomResourceDefinition{
					ObjectMeta: v1.ObjectMeta{
						Name: clusterBuildStrategiesCRD,
					},
				})
			}
			crdClient := apiextensionsfake.NewSimpleClientset(objects...)
			schemeBuilder := runtime.NewSchemeBuilder(scheme.AddToScheme, buildv1alpha1.AddToScheme)
			scheme := runtime.NewScheme()
			err := schemeBuilder.AddToScheme(scheme)
			o.Expect(err).NotTo(HaveOccurred(), "create k8s client scheme")
			k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			log := zap.New()
			manifests, err := common.SetupManifestival(k8sClient, filepath.Join("samples", "buildstrategy"), true, log)
			o.Expect(err).NotTo(HaveOccurred(), "setting up Manifestival")
			requeue, err := ReconcileBuildStrategies(ctx, crdClient.ApiextensionsV1(), log, manifests)
			o.Expect(err).NotTo(HaveOccurred(), "reconciling build strategies")
			o.Expect(requeue).To(BeEquivalentTo(tc.expectRequeue), "check reconcile requeue")

			if tc.expectStrategiesInstalled {
				strategies, err := test.ParseBuildStrategyNames()
				t.Logf("build strategies: %s", strategies)
				o.Expect(err).NotTo(HaveOccurred(), "parse build strategy names")
				for _, strategy := range strategies {
					obj := &buildv1alpha1.ClusterBuildStrategy{
						ObjectMeta: v1.ObjectMeta{
							Name: strategy,
						},
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj)
					o.Expect(err).NotTo(HaveOccurred(), "get ClusterBuildStrategy %s", strategy)
				}
			}
		})
	}
}
