package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/test"
)

var _ = Describe("Install embedded build strategies", func() {

	var build *v1alpha1.ShipwrightBuild

	BeforeEach(func(ctx SpecContext) {
		setupTektonCRDs(ctx)
		build = createShipwrightBuild(ctx, "shipwright")
		test.CRDEventuallyExists(ctx, k8sClient, "clusterbuildstrategies.shipwright.io")
	})

	When("the install build strategies feature is enabled", func() {

		It("applies the embedded build strategy manifests to the cluster", func(ctx SpecContext) {
			expectedBuildStrategies, err := test.ParseBuildStrategyNames()
			Expect(err).NotTo(HaveOccurred())
			for _, strategy := range expectedBuildStrategies {
				strategyObj := &buildv1alpha1.ClusterBuildStrategy{
					ObjectMeta: metav1.ObjectMeta{
						Name: strategy,
					},
				}
				By(fmt.Sprintf("checking for build strategy %q", strategy))
				test.EventuallyExists(ctx, k8sClient, strategyObj)
			}

		})
	})

	AfterEach(func(ctx SpecContext) {
		deleteShipwrightBuild(ctx, build)
	})

})
