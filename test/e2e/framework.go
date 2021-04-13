package e2e

import (
	"context"
	"encoding/json"

	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func waitForOperator(ctx context.Context, client client.Client) error {
	var err error
	Eventually(func() bool {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "operator-controller-manager",
				Namespace: "shipwright-operator",
			},
		}
		err = client.Get(ctx,
			types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name},
			deployment)
		if errors.IsNotFound(err) {
			return false
		}
		if err != nil {
			// Break out early, return an error
			return true
		}
		return deployment.Status.AvailableReplicas > 0 && deployment.Status.UpdatedReplicas > 0
	}).Should(BeTrue())

	return err
}

func getOperatorDeployment(ctx context.Context, k8sClient client.Client) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "operator-controller-manager",
			Namespace: "shipwright-operator",
		},
	}
	err := k8sClient.Get(ctx,
		types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name},
		deployment)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	return deployment, err
}

func getOperatorDeploymentJSON(ctx context.Context, k8sClient client.Client) (string, error) {
	deployment, err := getOperatorDeployment(ctx, k8sClient)
	if err != nil {
		return "", err
	}
	if deployment == nil {
		return "", nil
	}
	prettyJSON, err := json.MarshalIndent(deployment, "", "  ")
	if err != nil {
		return "", err
	}
	return string(prettyJSON), nil
}

func getOperatorPods(ctx context.Context, k8sClient client.Client) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	err := k8sClient.List(ctx, pods, client.InNamespace("shipwright-operator"))
	if errors.IsNotFound(err) {
		return nil, nil
	}
	return pods, err
}

func getOperatorPodsJSON(ctx context.Context, k8sClient client.Client) (string, error) {
	pods, err := getOperatorPods(ctx, k8sClient)
	if err != nil {
		return "", err
	}
	if pods == nil {
		return "", nil
	}
	prettyJSON, err := json.MarshalIndent(pods, "", "  ")
	if err != nil {
		return "", err
	}
	return string(prettyJSON), nil
}
