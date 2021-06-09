package controllers

import (
	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/test"
)

var _ = g.Describe("Reconcile default ShipwrightBuild installation", func() {

	// namespace where ShipwrightBuild instance will be located
	const namespace = "default"
	// namespace where shipwright Controller and dependencies will be located
	const targetNamespace = "namespace"

	var build *v1alpha1.ShipwrightBuild

	g.BeforeEach(func() {
		// setting up the target namespace, where Shipwright Controller will be deployed
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: targetNamespace}}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: ns.Name}, ns)
		if errors.IsNotFound(err) {
			err = k8sClient.Create(ctx, ns, &client.CreateOptions{})
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("creating a ShipwrightBuild instance")
		build = &v1alpha1.ShipwrightBuild{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "cluster",
			},
			Spec: v1alpha1.ShipwrightBuildSpec{
				Namespace: targetNamespace,
			},
		}
		err = k8sClient.Create(ctx, build, &client.CreateOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		// when the finalizer is in place, the deployment of manifest elements is done, and therefore
		// functional testing can proceed
		g.By("waiting for the finalizer to be set")
		test.EventuallyContainFinalizer(ctx, k8sClient, build, FinalizerAnnotation)
	})

	g.AfterEach(func() {
		g.By("deleting the ShipwrightBuild instance")
		err := k8sClient.Get(ctx, types.NamespacedName{Name: build.Name}, build)
		if errors.IsNotFound(err) {
			return
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		err = k8sClient.Delete(ctx, build, &client.DeleteOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("waiting for ShipwrightBuild instance to be completely removed")
		test.EventuallyRemoved(ctx, k8sClient, build)

		g.By("checking that the shipwright-build-controller deployment has been removed")
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: targetNamespace,
				Name:      "shipwright-build-controller",
			},
		}
		test.EventuallyRemoved(ctx, k8sClient, deployment)
	})

	g.When("a ShipwrightBuild object is created", func() {

		g.It("creates RBAC for the Shipwright build controller", func() {
			expectedClusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "shipwright-build-controller",
				},
			}
			test.EventuallyExists(ctx, k8sClient, expectedClusterRole)

			expectedClusterRoleBinding := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "shipwright-build-controller",
				},
			}
			test.EventuallyExists(ctx, k8sClient, expectedClusterRoleBinding)

			expectedServiceAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: targetNamespace,
					Name:      "shipwright-build-controller",
				},
			}
			test.EventuallyExists(ctx, k8sClient, expectedServiceAccount)
		})

		g.It("creates a deployment for the Shipwright build controller", func() {
			expectedDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: targetNamespace,
					Name:      "shipwright-build-controller",
				},
			}
			test.EventuallyExists(ctx, k8sClient, expectedDeployment)
		})

		g.It("creates custom resource definitions for the Shipwright build APIs", func() {
			test.CRDEventuallyExists(ctx, k8sClient, "builds.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildruns.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildstrategies.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "clusterbuildstrategies.shipwright.io")
		})
	})

	g.When("a ShipwrightBuild object is deleted", func() {

		g.It("deletes the RBAC for the Shipwright build controller", func() {
			expectedClusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: targetNamespace,
					Name:      "shipwright-build-controller",
				},
			}
			expectedClusterRoleBinding := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: targetNamespace,
					Name:      "shipwright-build-controller",
				},
			}
			expectedServiceAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: targetNamespace,
					Name:      "shipwright-build-controller",
				},
			}

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

		g.It("deletes the deployment for the Shipwright build controller", func() {
			expectedDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: targetNamespace,
					Name:      "shipwright-build-controller",
				},
			}
			// Setup - ensure the objects we want exist
			test.EventuallyExists(ctx, k8sClient, expectedDeployment)

			// Test - delete the ShipwrightBuild instance
			err := k8sClient.Delete(ctx, build, &client.DeleteOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())

			// Verify - check the behavior
			test.EventuallyRemoved(ctx, k8sClient, expectedDeployment)
		})

		// TODO: Do not delete the CRDs! This is something only a cluster admin should do.
		g.It("deletes the custom resource definitions for the Shipwright build APIs", func() {
			test.CRDEventuallyExists(ctx, k8sClient, "builds.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildruns.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "buildstrategies.shipwright.io")
			test.CRDEventuallyExists(ctx, k8sClient, "clusterbuildstrategies.shipwright.io")

			// Test - delete the ShipwrightBuild instance
			err := k8sClient.Delete(ctx, build, &client.DeleteOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())

			// Verify - check the behavior
			test.CRDEventuallyRemoved(ctx, k8sClient, "builds.shipwright.io")
			test.CRDEventuallyRemoved(ctx, k8sClient, "buildruns.shipwright.io")
			test.CRDEventuallyRemoved(ctx, k8sClient, "buildstrategies.shipwright.io")
			test.CRDEventuallyRemoved(ctx, k8sClient, "clusterbuildstrategies.shipwright.io")
		})
	})

})
