package buildstrategy

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/manifestival/manifestival"
	. "github.com/onsi/gomega"

	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/operator/internal/pkg/common"
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
			schemeBuilder := runtime.NewSchemeBuilder(scheme.AddToScheme, buildv1beta1.AddToScheme)
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
				strategies, err := ParseBuildStrategyNames()
				t.Logf("build strategies: %s", strategies)
				o.Expect(err).NotTo(HaveOccurred(), "parse build strategy names")
				for _, strategy := range strategies {
					obj := &buildv1beta1.ClusterBuildStrategy{
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

// ParseBuildStrategyNames returns a list of object names from the embedded build strategy
// manifests.
func ParseBuildStrategyNames() ([]string, error) {
	koDataPath, err := common.KoDataPath()
	if err != nil {
		return nil, err
	}
	strategyPath := filepath.Join(koDataPath, "samples", "buildstrategy")
	sampleNames := []string{}
	manifest, err := manifestival.ManifestFrom(manifestival.Recursive(strategyPath))
	if err != nil {
		return sampleNames, err
	}
	for _, obj := range manifest.Resources() {
		if obj.GetKind() == "ClusterBuildStrategy" {
			sampleNames = append(sampleNames, obj.GetName())
		}

	}
	if len(sampleNames) == 0 {
		return sampleNames, fmt.Errorf("no ClusterBuildStrategy objects found in %s", strategyPath)
	}
	return sampleNames, nil
}
