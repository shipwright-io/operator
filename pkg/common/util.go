package common

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"path/filepath"

	"github.com/go-logr/logr"
	mfc "github.com/manifestival/controller-runtime-client"
	"github.com/manifestival/manifestival"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetupManifestival instantiates a Manifestival instance for the provided file or directory
func SetupManifestival(client client.Client, fileOrDir string, recurse bool, logger logr.Logger) (manifestival.Manifest, error) {
	mfclient := mfc.NewClient(client)

	dataPath, err := KoDataPath()
	if err != nil {
		return manifestival.Manifest{}, err
	}
	manifest := filepath.Join(dataPath, fileOrDir)
	var src manifestival.Source
	if recurse {
		src = manifestival.Recursive(manifest)
	} else {
		src = manifestival.Path(manifest)
	}
	return manifestival.ManifestFrom(src, manifestival.UseClient(mfclient), manifestival.UseLogger(logger))
}

// KoDataPath retrieve the data path environment variable, returning error when not found.
func KoDataPath() (string, error) {
	dataPath, exists := os.LookupEnv(koDataPathEnv)
	if !exists {
		return "", fmt.Errorf("'%s' is not set", koDataPathEnv)
	}
	return dataPath, nil
}

// contains returns true if the string if found in the slice.
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// imagesFromEnv will provide map of key value.
func ImagesFromEnv(prefix string) map[string]string {
	images := map[string]string{}
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		keyValue := strings.Split(env, "=")
		name := strings.TrimPrefix(keyValue[0], prefix)
		url := keyValue[1]
		images[name] = url
	}

	return images
}

// toLowerCaseKeys converts key value to lower cases.
func ToLowerCaseKeys(keyValues map[string]string) map[string]string {
	newMap := map[string]string{}

	for k, v := range keyValues {
		key := strings.ToLower(k)
		newMap[key] = v
	}

	return newMap
}

// deploymentImages replaces container and env vars images.
func DeploymentImages(images map[string]string) manifestival.Transformer {
	return func(u *unstructured.Unstructured) error {
		if u.GetKind() != "Deployment" {
			return nil
		}

		d := &appsv1.Deployment{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, d)
		if err != nil {
			return err
		}

		containers := d.Spec.Template.Spec.Containers
		replaceContainerImages(containers, images)
		unstrObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(d)
		if err != nil {
			return err
		}
		u.SetUnstructuredContent(unstrObj)

		return nil
	}
}

func formKey(prefix, arg string) string {
	argument := strings.ToLower(arg)
	if prefix != "" {
		argument = prefix + argument
	}
	return strings.ReplaceAll(argument, "-", "_")
}

func replaceContainerImages(containers []corev1.Container, images map[string]string) {
	for i, container := range containers {
		name := formKey("", container.Name)
		if url, exist := images[name]; exist {
			containers[i].Image = url
		}

		replaceContainersEnvImage(container, images)
	}
}

func replaceContainersEnvImage(container corev1.Container, images map[string]string) {
	for index, env := range container.Env {
		if url, exist := images[formKey("", env.Name)]; exist {
			container.Env[index].Value = url
		}
	}
}

func CRDExist(ctx context.Context, client crdclientv1.ApiextensionsV1Interface, crdName string) (bool, error) {
	_, err := client.CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get customresourcedefinition %s: %v", crdName, err)
	}
	return true, nil
}

func BoolFromEnvVar(envVar string) bool {
	if v, ok := os.LookupEnv(envVar); ok {
		if vv, err := strconv.ParseBool(v); err == nil {
			return vv
		}
	}
	return false
}

// InjectAnnotations adds annotation key:value to a resource annotations
// overwritePolicy (Retain/Overwrite) decides whehther to overwrite an already existing annotation
// []kinds specify the Kinds on which the label should be applied
// if len(kinds) = 0, label will be apllied to all/any resources irrespective of its Kind
func InjectAnnotations(key, value string, overwritePolicy int, kinds ...string) manifestival.Transformer {
	return func(u *unstructured.Unstructured) error {
		kind := u.GetKind()
		if len(kinds) != 0 && !itemInSlice(kind, kinds) {
			return nil
		}
		annotations, found, err := unstructured.NestedStringMap(u.Object, "metadata", "annotations")
		if err != nil {
			return fmt.Errorf("could not find annotation set, %q", err)
		}
		if overwritePolicy == Retain && found {
			if _, ok := annotations[key]; ok {
				return nil
			}
		}
		if !found {
			annotations = map[string]string{}
		}
		annotations[key] = value
		err = unstructured.SetNestedStringMap(u.Object, annotations, "metadata", "annotations")
		if err != nil {
			return fmt.Errorf("error updating annotations for %s:%s, %s", kind, u.GetName(), err)
		}
		return nil
	}
}

func itemInSlice(item string, items []string) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

func IsOpenShiftPlatform() bool {
	return os.Getenv("PLATFORM") == "openshift"
}
