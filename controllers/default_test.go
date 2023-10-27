package controllers

import (
	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
	commonctrl "github.com/shipwright-io/operator/controllers/common"
	"github.com/shipwright-io/operator/pkg/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shipwright-io/operator/api/v1alpha1"
	"github.com/shipwright-io/operator/test"
)

// createNamespace creates the namespace informed.
func createNamespace(name string) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: ns.Name}, ns)
	if errors.IsNotFound(err) {
		err = k8sClient.Create(ctx, ns, &client.CreateOptions{})
	}
	o.Expect(err).NotTo(o.HaveOccurred())
}

var _ = g.Describe("Reconcile default ShipwrightBuild installation", func() {

	// namespace where ShipwrightBuild instance will be located
	const namespace = "namespace"
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

	truePtr := true
	g.BeforeEach(func() {
		// setting up the namespaces, where Shipwright Controller will be deployed
		createNamespace(namespace)

		g.By("does tekton taskrun crd exist")
		err := k8sClient.Get(ctx, types.NamespacedName{Name: "taskruns.tekton.dev"}, &crdv1.CustomResourceDefinition{})
		if errors.IsNotFound(err) {
			g.By("creating tekton taskrun crd")
			taskRunCRD := &crdv1.CustomResourceDefinition{}
			taskRunCRD.Name = "taskruns.tekton.dev"
			taskRunCRD.Spec.Group = "tekton.dev"
			taskRunCRD.Spec.Scope = crdv1.NamespaceScoped
			taskRunCRD.Spec.Versions = []crdv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1beta1",
					Storage: true,
					Schema: &crdv1.CustomResourceValidation{
						OpenAPIV3Schema: &crdv1.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: &truePtr,
						},
					},
				},
			}
			taskRunCRD.Spec.Names.Plural = "taskruns"
			taskRunCRD.Spec.Names.Singular = "taskrun"
			taskRunCRD.Spec.Names.Kind = "TaskRun"
			taskRunCRD.Spec.Names.ListKind = "TaskRunList"
			taskRunCRD.Status.StoredVersions = []string{"v1beta1"}
			err = k8sClient.Create(ctx, taskRunCRD, &client.CreateOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())

		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("does tektonconfig crd exist")
		err = k8sClient.Get(ctx, types.NamespacedName{Name: "tektonconfigs.operator.tekton.dev"}, &crdv1.CustomResourceDefinition{})
		if errors.IsNotFound(err) {
			tektonOpCRD := &crdv1.CustomResourceDefinition{}
			tektonOpCRD.Name = "tektonconfigs.operator.tekton.dev"
			tektonOpCRD.Labels = map[string]string{"operator.tekton.dev/release": common.TektonOpMinSupportedVersion}
			tektonOpCRD.Spec.Group = "operator.tekton.dev"
			tektonOpCRD.Spec.Scope = crdv1.ClusterScoped
			tektonOpCRD.Spec.Versions = []crdv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Storage: true,
					Schema: &crdv1.CustomResourceValidation{
						OpenAPIV3Schema: &crdv1.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: &truePtr,
						},
					},
				},
			}
			tektonOpCRD.Spec.Names.Plural = "tektonconfigs"
			tektonOpCRD.Spec.Names.Singular = "tektonconfig"
			tektonOpCRD.Spec.Names.Kind = "TektonConfig"
			tektonOpCRD.Spec.Names.ListKind = "TektonConfigList"
			tektonOpCRD.Status.StoredVersions = []string{"v1alpha1"}
			err = k8sClient.Create(ctx, tektonOpCRD, &client.CreateOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("creating a ShipwrightBuild instance")
		build = &v1alpha1.ShipwrightBuild{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "cluster",
			},
			Spec: v1alpha1.ShipwrightBuildSpec{
				TargetNamespace: targetNamespace,
			},
		}
		err = k8sClient.Create(ctx, build, &client.CreateOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())

		// when the finalizer is in place, the deployment of manifest elements is done, and therefore
		// functional testing can proceed
		g.By("waiting for the finalizer to be set")
		test.EventuallyContainFinalizer(ctx, k8sClient, build, commonctrl.FinalizerAnnotation)
	})

	g.AfterEach(func() {
		g.By("deleting the ShipwrightBuild instance")
		namespacedName := types.NamespacedName{Namespace: namespace, Name: build.Name}
		err := k8sClient.Get(ctx, namespacedName, build)
		if errors.IsNotFound(err) {
			return
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		err = k8sClient.Delete(ctx, build, &client.DeleteOptions{})
		// the delete e2e's can delete this object before this AfterEach runs
		if errors.IsNotFound(err) {
			return
		}
		o.Expect(err).NotTo(o.HaveOccurred())

		g.By("waiting for ShipwrightBuild instance to be completely removed")
		test.EventuallyRemoved(ctx, k8sClient, build)

		g.By("checking that the shipwright-build-controller deployment has been removed")
		deployment := baseDeployment.DeepCopy()
		test.EventuallyRemoved(ctx, k8sClient, deployment)
	})

	g.When("a ShipwrightBuild object is created", func() {

		g.It("creates RBAC for the Shipwright build controller", func() {
			expectedClusterRole := baseClusterRole.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedClusterRole)

			expectedClusterRoleBinding := baseClusterRoleBinding.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedClusterRoleBinding)

			expectedServiceAccount := baseServiceAccount.DeepCopy()
			test.EventuallyExists(ctx, k8sClient, expectedServiceAccount)
		})

		g.It("creates a deployment for the Shipwright build controller", func() {
			expectedDeployment := baseDeployment.DeepCopy()
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

		g.It("deletes the deployment for the Shipwright build controller", func() {
			expectedDeployment := baseDeployment.DeepCopy()
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
