package common

import (
	"path"
	"testing"

	mf "github.com/manifestival/manifestival"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestImagesFromEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	t.Setenv("IMAGE_SHIPWRIGHT_CONTROLLER", "docker.io/shipwright-controller")
	data := ImagesFromEnv(ShipwrightImagePrefix)
	Expect(data).To(Equal(map[string]string{"CONTROLLER": "docker.io/shipwright-controller"}))
}

func TestDeploymentImages(t *testing.T) {
	RegisterFailHandler(Fail)
	t.Run("ignore non deployment images", func(t *testing.T) {
		testData := path.Join("testdata", "test-replace-kind.yaml")
		expected, _ := mf.ManifestFrom(mf.Recursive(testData))

		manifest, err := mf.ManifestFrom(mf.Recursive(testData))
		Expect(err).NotTo(HaveOccurred())

		newManifest, err := manifest.Transform(DeploymentImages(map[string]string{}))
		Expect(err).NotTo(HaveOccurred())

		Expect(expected.Resources()).To(Equal(newManifest.Resources()))
	})
	t.Run("replace containers image by name", func(t *testing.T) {
		image := "foo.bar/image"
		images := map[string]string{
			"IMAGE_SHIPWRIGHT_SHIPWRIGHT_BUILD": image,
		}
		testData := path.Join("testdata", "test-replace-image.yaml")

		manifest, err := mf.ManifestFrom(mf.Recursive(testData))
		Expect(err).NotTo(HaveOccurred())
		newManifest, err := manifest.Transform(DeploymentImages(images))
		Expect(err).NotTo(HaveOccurred())
		assertDeployContainersHasImage(t, newManifest.Resources(), "SHIPWRIGHT_BUILD", image)
	})
	t.Run("replace containers env", func(t *testing.T) {
		image := "foo.bar/image/bash"
		images := map[string]string{
			"IMAGE_SHIPWRIGHT_GIT_CONTAINER_IMAGE": image,
		}
		testData := path.Join("testdata", "test-replace-image.yaml")

		manifest, err := mf.ManifestFrom(mf.Recursive(testData))
		Expect(err).NotTo(HaveOccurred())
		newManifest, err := manifest.Transform(DeploymentImages(images))
		Expect(err).NotTo(HaveOccurred())
		assertDeployContainerEnvsHasImage(t, newManifest.Resources(), "IMAGE_SHIPWRIGHT_GIT_CONTAINER_IMAGE", image)
	})
}

func deploymentFor(t *testing.T, unstr unstructured.Unstructured) *appsv1.Deployment {
	deployment := &appsv1.Deployment{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.Object, deployment)
	if err != nil {
		t.Errorf("failed to load deployment yaml")
	}
	return deployment
}

func assertDeployContainersHasImage(t *testing.T, resources []unstructured.Unstructured, name string, image string) {
	t.Helper()

	for _, resource := range resources {
		deployment := deploymentFor(t, resource)
		containers := deployment.Spec.Template.Spec.Containers
		for _, container := range containers {
			if container.Name != name {
				continue
			}
			if container.Image != image {
				t.Errorf("assertion failed; unexpected image: expected %s and got %s", image, container.Image)
			}
		}
	}
}

func assertDeployContainerEnvsHasImage(t *testing.T, resources []unstructured.Unstructured, env string, image string) {
	t.Helper()

	for _, resource := range resources {
		deployment := deploymentFor(t, resource)
		containers := deployment.Spec.Template.Spec.Containers

		for _, container := range containers {
			if len(container.Env) == 0 {
				continue
			}

			for index, envVar := range container.Env {
				if envVar.Name == env && container.Env[index].Value != image {
					t.Errorf("not equal: expected %v, got %v", image, container.Env[index].Value)
				}
			}
		}
	}
}

func TestBoolFromEnvVar(t *testing.T) {
	RegisterFailHandler(Fail)
	cases := []struct {
		name           string
		envVar         string
		expectedResult bool
	}{
		{
			name:           "env var not set",
			envVar:         "",
			expectedResult: false,
		},
		{
			name:           "env var set",
			envVar:         "USE_MANAGED_CERTS",
			expectedResult: true,
		},
	}
	for _, tc := range cases {
		if tc.envVar != "" {
			t.Setenv(tc.envVar, "true")
		}
		Expect(BoolFromEnvVar("USE_MANAGED_CERTS")).To(Equal(tc.expectedResult))
	}
}
