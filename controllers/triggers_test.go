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

var _ = g.Describe("Reconcile ShipwrightBuild with Triggers", func() {

	const targetNamespace = "triggers-test-ns"
	var build *v1alpha1.ShipwrightBuild

	triggersDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: targetNamespace,
			Name:      "shipwright-triggers",
		},
	}

	g.When("triggers are enabled", g.Ordered, func() {

		g.BeforeAll(func(ctx g.SpecContext) {
			setupTektonCRDs(ctx)
			createTektonConfig(ctx)
			build = createShipwrightBuildWithTriggers(ctx, "triggers-enabled", targetNamespace)
		})

		g.AfterAll(func(ctx g.SpecContext) {
			deleteShipwrightBuild(ctx, build)
			deleteTektonConfig(ctx)
		})

		g.It("creates triggers resources", func(ctx g.SpecContext) {
			test.EventuallyExists(ctx, k8sClient, triggersDeployment.DeepCopy())
			test.EventuallyExists(ctx, k8sClient, &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Namespace: targetNamespace, Name: "shipwright-triggers"},
			})
			test.EventuallyExists(ctx, k8sClient, &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Namespace: targetNamespace, Name: "shipwright-triggers"},
			})
			test.EventuallyExists(ctx, k8sClient, &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{Name: "shipwright-triggers"},
			})
			test.EventuallyExists(ctx, k8sClient, &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "shipwright-triggers"},
			})
		})
	})

	g.When("triggers are enabled and the ShipwrightBuild is deleted", g.Ordered, func() {

		g.BeforeAll(func(ctx g.SpecContext) {
			setupTektonCRDs(ctx)
			createTektonConfig(ctx)
			build = createShipwrightBuildWithTriggers(ctx, "triggers-delete", targetNamespace)
		})

		g.AfterAll(func(ctx g.SpecContext) {
			deleteTektonConfig(ctx)
		})

		g.It("removes the triggers deployment", func(ctx g.SpecContext) {
			expectedDeployment := triggersDeployment.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedDeployment)

			err := k8sClient.Delete(ctx, build, &client.DeleteOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())

			test.EventuallyRemoved(ctx, k8sClient, expectedDeployment)
		})
	})

	g.When("triggers are disabled after being enabled", g.Ordered, func() {
		const disableNamespace = "triggers-disable-ns"

		g.BeforeAll(func(ctx g.SpecContext) {
			setupTektonCRDs(ctx)
			createTektonConfig(ctx)
			build = createShipwrightBuildWithTriggers(ctx, "triggers-disable", disableNamespace)
		})

		g.AfterAll(func(ctx g.SpecContext) {
			deleteShipwrightBuild(ctx, build)
			deleteTektonConfig(ctx)
		})

		g.It("removes triggers resources when disabled via spec update", func(ctx g.SpecContext) {
			expectedDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: disableNamespace, Name: "shipwright-triggers"},
			}
			test.EventuallyExists(ctx, k8sClient, expectedDeployment)

			g.By("disabling triggers")
			err := k8sClient.Get(ctx, client.ObjectKeyFromObject(build), build)
			o.Expect(err).NotTo(o.HaveOccurred())
			enableFalse := false
			build.Spec.Triggers.Enable = &enableFalse
			err = k8sClient.Update(ctx, build)
			o.Expect(err).NotTo(o.HaveOccurred())

			test.EventuallyRemoved(ctx, k8sClient, expectedDeployment)
		})
	})

	g.When("triggers are not enabled", g.Ordered, func() {
		const noTriggersNamespace = "no-triggers-test-ns"

		g.BeforeAll(func(ctx g.SpecContext) {
			setupTektonCRDs(ctx)
			createTektonConfig(ctx)
			build = createShipwrightBuild(ctx, "no-triggers", noTriggersNamespace)
		})

		g.AfterAll(func(ctx g.SpecContext) {
			deleteShipwrightBuild(ctx, build)
			deleteTektonConfig(ctx)
		})

		g.It("does not create the triggers deployment", func(ctx g.SpecContext) {
			test.EventuallyExists(ctx, k8sClient, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: noTriggersNamespace, Name: "shipwright-build-controller"},
			})

			noTriggersDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: noTriggersNamespace, Name: "shipwright-triggers"},
			}
			o.Consistently(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(noTriggersDeployment), noTriggersDeployment)
				return err != nil
			}, "3s", "500ms").Should(o.BeTrue(), "triggers deployment should not be created")
		})
	})
})
