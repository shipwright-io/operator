package test

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	o "github.com/onsi/gomega"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/operator/pkg/common"
)

// timeout amount of time to wait for Eventually methods
var timeout = 30 * time.Second

// EventuallyExists checks if an object with the given namespace+name and type eventually exists.
func EventuallyExists(ctx context.Context, k8sClient client.Client, obj client.Object) {
	o.Eventually(func() bool {
		key := types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		}
		err := k8sClient.Get(ctx, key, obj)
		if errors.IsNotFound(err) {
			return false
		}
		if meta.IsNoMatchError(err) {
			// For CRDs created by the operator, we may need to wait.
			return false
		}
		o.Expect(err).NotTo(o.HaveOccurred())
		return true
	}, timeout).Should(o.BeTrue(), "waiting for object %s/%s to exist", obj.GetNamespace(), obj.GetName())
}

// EventuallyContainFinalizer retrieves and inspect the object to assert if the informed finalizer
// string is in the object.
func EventuallyContainFinalizer(
	ctx context.Context,
	k8sClient client.Client,
	obj client.Object,
	finalizer string,
) {
	o.Eventually(func() bool {
		key := types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		}
		if err := k8sClient.Get(ctx, key, obj); err != nil {
			return false
		}
		for _, s := range obj.GetFinalizers() {
			if s == finalizer {
				return true
			}
		}
		return false
	}, timeout).Should(o.BeTrue())
}

// CRDEventuallyExists checks if a custom resource definition with the given name eventually exists.
func CRDEventuallyExists(ctx context.Context, k8sClient client.Client, crdName string) {
	crd := &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
	}
	EventuallyExists(ctx, k8sClient, crd)
}

// EventuallyRemoved checks if an object is eventually deleted
func EventuallyRemoved(ctx context.Context, k8sClient client.Client, obj client.Object) {
	o.Eventually(func() bool {
		key := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
		err := k8sClient.Get(ctx, key, obj)
		return errors.IsNotFound(err)
	}, timeout).Should(o.BeTrue())
}

// CRDEventuallyRemoved checks if a custom resource definition has been eventually removed
func CRDEventuallyRemoved(ctx context.Context, k8sClient client.Client, crdName string) {
	crd := &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
	}
	EventuallyRemoved(ctx, k8sClient, crd)
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
	err = filepath.WalkDir(strategyPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		clusterBuildStrategy := &v1alpha1.ClusterBuildStrategy{}
		decodeErr := decodeYaml(path, clusterBuildStrategy)
		if decodeErr != nil {
			return decodeErr
		}
		sampleNames = append(sampleNames, clusterBuildStrategy.Name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sampleNames, nil
}

func decodeYaml(path string, obj *v1alpha1.ClusterBuildStrategy) error {
	yaml, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(yaml), 16)
	return decoder.Decode(obj)
}
