package test

import (
	"context"
	"time"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	o "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		o.Expect(err).NotTo(o.HaveOccurred())
		return true
	}, timeout).Should(o.BeTrue())
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
