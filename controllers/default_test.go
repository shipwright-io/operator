package controllers

import (
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/test"
)

var _ = g.Describe("Reconcile default ShipwrightBuild installation", func() {

	// targetNamespace namespace where shipwright Controller and dependencies will be located
	const targetNamespace = "target-namespace"
	// build Build instance employed during testing
	var build *v1alpha1.ShipwrightBuild

	baseClusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shipwright-build-controller",
		},
	}
	baseClusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shipwright-build-controller",
		},
	}
	baseServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: targetNamespace,
			Name:      "shipwright-build-controller",
		},
	}
	baseDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: targetNamespace,
			Name:      "shipwright-build-controller",
		},
	}

	g.BeforeEach(func(ctx g.SpecContext) {
		// setting up the namespaces, where Shipwright Controller will be deployed
		setupTektonCRDs(ctx)
		build = createShipwrightBuild(ctx, targetNamespace)
	})

	g.AfterEach(func(ctx g.SpecContext) {
		deleteShipwrightBuild(ctx, build)

		g.By("checking that the shipwright-build-controller deployment has been removed")
		deployment := baseDeployment.DeepCopy()
		test.EventuallyRemoved(ctx, k8sClient, deployment)
	})

	g.When("a ShipwrightBuild object is created", func() {

		g.It("creates RBAC for the Shipwright build controller", func(ctx g.SpecContext) {
			expectedClusterRole := baseClusterRole.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedClusterRole)

			expectedClusterRoleBinding := baseClusterRoleBinding.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedClusterRoleBinding)

			expectedServiceAccount := baseServiceAccount.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedServiceAccount)
		})

		g.It("creates a deployment for the Shipwright build controller", func(ctx g.SpecContext) {
			expectedDeployment := baseDeployment.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedDeployment)
		})

		g.It("creates custom resource definitions for the Shipwright build APIs", func(ctx g.SpecContext) {
			test.CRDEventuallyExists(ctx, k8sClient, "builds.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildruns.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildstrategies.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "clusterbuildstrategies.shipwright.io")
		})
	})

	g.When("a ShipwrightBuild object is deleted", func() {

		g.It("deletes the RBAC for the Shipwright build controller", func(ctx g.SpecContext) {
			expectedClusterRole := baseClusterRole.DeepCopy()
			expectedClusterRoleBinding := baseClusterRoleBinding.DeepCopy()
			expectedServiceAccount := baseServiceAccount.DeepCopy()

			// Setup - ensure the objects we want exist
			test.EventuallyExists(ctx, k8sClient, expectedClusterRole)
			test.EventuallyExists(ctx, k8sClient, expectedClusterRoleBinding)
			test.EventuallyExists(ctx, k8sClient, expectedServiceAccount)

			// Test - delete the ShipwrightBuild instance
			err := k8sClient.Delete(ctx, build, &client.DeleteOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())

			// Verify - check the behavior
			test.EventuallyRemoved(ctx, k8sClient, expectedClusterRole)
			test.EventuallyRemoved(ctx, k8sClient, expectedClusterRoleBinding)
			test.EventuallyRemoved(ctx, k8sClient, expectedServiceAccount)
		})

		g.It("deletes the deployment for the Shipwright build controller", func(ctx g.SpecContext) {
			expectedDeployment := baseDeployment.DeepCopy()
			// Setup - ensure the objects we want exist
			test.EventuallyExists(ctx, k8sClient, expectedDeployment)

			// Test - delete the ShipwrightBuild instance
			err := k8sClient.Delete(ctx, build, &client.DeleteOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())

			// Verify - check the behavior
			test.EventuallyRemoved(ctx, k8sClient, expectedDeployment)
		})

		// Deteling the build instance should not delete CRDS
		g.It("should not delete the custom resource definitions for the Shipwright build APIs", func(ctx g.SpecContext) {
			test.CRDEventuallyExists(ctx, k8sClient, "builds.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildruns.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildstrategies.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "clusterbuildstrategies.shipwright.io")

			// Test - delete the ShipwrightBuild instance
			err := k8sClient.Delete(ctx, build, &client.DeleteOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())

			// Verify - check the behavior
			test.CRDEventuallyExists(ctx, k8sClient, "builds.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildruns.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildstrategies.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "clusterbuildstrategies.shipwright.io")
		})
	})

})
