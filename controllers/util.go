package controllers

import (
	"fmt"
	"os"
	"strings"

	mf "github.com/manifestival/manifestival"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// koDataPathEnv ko data-path environment variable.
const (
	koDataPathEnv         = "KO_DATA_PATH"
	InstallTriggers       = "INSTALL_TRIGGERS"
	ShipwrightImagePrefix = "IMAGE_SHIPWRIGHT_"
)

// koDataPath retrieve the data path environment variable, returning error when not found.
func koDataPath() (string, error) {
	dataPath, exists := os.LookupEnv(koDataPathEnv)
	if !exists {
		return "", fmt.Errorf("'%s' is not set", koDataPathEnv)
	}
	return dataPath, nil
}

// contains returns true if the string if found in the slice.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// imagesFromEnv will provide map of key value.
func imagesFromEnv(prefix string) map[string]string {
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
func toLowerCaseKeys(keyValues map[string]string) map[string]string {
	newMap := map[string]string{}

	for k, v := range keyValues {
		key := strings.ToLower(k)
		newMap[key] = v
	}

	return newMap
}

// deploymentImages replaces container and env vars images.
func deploymentImages(images map[string]string) mf.Transformer {
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

func withTriggers() bool {
	return os.Getenv(InstallTriggers) == "true"
}
