package tekton

import (
	"context"
	"fmt"

	"github.com/shipwright-io/operator/pkg/common"
	tektonoperatorv1alpha1 "github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	tektonoperatorclientv1alpha1 "github.com/tektoncd/operator/pkg/client/clientset/versioned/typed/operator/v1alpha1"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
)

// ReconcileTekton ensures that Tekton Pipelines has been installed.
// If Tekton Pipelines has not been installed, ReconcileTekton will create a TektonConfig object
// so that the Tekton Operator deploys Tekton Pipelines.
func ReconcileTekton(ctx context.Context,
	crdClient crdclientv1.ApiextensionsV1Interface,
	tektonOperatorClient tektonoperatorclientv1alpha1.OperatorV1alpha1Interface) (*tektonoperatorv1alpha1.TektonConfig, bool, error) {
	pipelinesInstalled, err := IsTektonPipelinesInstalled(ctx, crdClient)
	if err != nil {
		return nil, true, err
	}
	if pipelinesInstalled {
		return nil, false, nil
	}
	tektonOperatorInstalled, err := IsTektonOperatorInstalled(ctx, crdClient)
	if err != nil {
		return nil, true, err
	}
	if !tektonOperatorInstalled {
		return nil, false, fmt.Errorf("tekton operator not installed")
	}
	tektonVersion, err := GetTektonOperatorVersion(ctx, crdClient)
	if err != nil {
		return nil, true, fmt.Errorf("failed to determine Tekton Operator version: %v", err)
	}
	if tektonVersion.Major() < common.TektonOpMinSupportedMajor+1 && tektonVersion.Minor() < common.TektonOpMinSupportedMinor {
		return nil, true, fmt.Errorf("insufficient Tekton Operator version - must be greater than %s", common.TektonOpMinSupportedVersion)
	}
	tektonConfigPresent, err := IsTektonConfigPresent(ctx, tektonOperatorClient)
	if err != nil {
		return nil, true, err
	}
	if tektonConfigPresent {
		return nil, false, nil
	}
	// the tekton operator 'lite' profile is all Shipwright currently needs, so configure that up;
	// when Shipwright starts leveraging triggers, we will want to bump up to a 'base' or higher
	tektonConfig, err := CreateTektonConfigWithProfileAndTargetNamespace(ctx,
		tektonOperatorClient, "lite", "tekton-pipelines")
	if err != nil {
		return tektonConfig, true, err
	}
	return tektonConfig, false, nil
}

// IsTektonPipelinesInstalled checks if Tekton has been installed on the cluster.
func IsTektonPipelinesInstalled(ctx context.Context, client crdclientv1.ApiextensionsV1Interface) (bool, error) {
	return common.CRDExist(ctx, client, "taskruns.tekton.dev")
}

// IsTektonOperatorInstalled checks if the Tekton Operator has been installed on the cluster.
func IsTektonOperatorInstalled(ctx context.Context, client crdclientv1.ApiextensionsV1Interface) (bool, error) {
	return common.CRDExist(ctx, client, "tektonconfigs.operator.tekton.dev")
}

// GetTektonOperatorVersion gets the semantic version of the installed Tekton operator.
// If the Tekton operator is not installed, or its semantic version could not be determined, this returns an error.
func GetTektonOperatorVersion(ctx context.Context, client crdclientv1.ApiextensionsV1Interface) (*version.Version, error) {
	tektonOpCRD, err := client.CustomResourceDefinitions().Get(ctx, "tektonconfigs.operator.tekton.dev", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if tektonOpCRD.Labels == nil {
		return nil, fmt.Errorf("the CRD TektonConfig does not have the label operator.tekton.dev/release to get its version")
	}
	value, exists := tektonOpCRD.Labels["operator.tekton.dev/release"]
	if !exists {
		return nil, fmt.Errorf("the CRD TektonConfig does not have the label operator.tekton.dev/release to get its version")
	}
	version, err := version.ParseSemantic(value)
	if err != nil {
		return nil, err
	}
	return version, nil
}

// IsTektonConfigPresent checks if at least one TektonConfig instance is present.
func IsTektonConfigPresent(ctx context.Context, client tektonoperatorclientv1alpha1.OperatorV1alpha1Interface) (bool, error) {
	list, err := client.TektonConfigs().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	if list == nil {
		return false, nil
	}
	return len(list.Items) > 0, err
}

// CreateTektonConfigWithProfileAndTargetNamespace creates a TektonConfig object with the given
// profile and target namespace for Tekton components.
func CreateTektonConfigWithProfileAndTargetNamespace(ctx context.Context, client tektonoperatorclientv1alpha1.OperatorV1alpha1Interface, profile string, targetNamepsace string) (*tektonoperatorv1alpha1.TektonConfig, error) {
	tektonConfig := &tektonoperatorv1alpha1.TektonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "config",
		},
		Spec: tektonoperatorv1alpha1.TektonConfigSpec{
			Profile: profile,
			CommonSpec: tektonoperatorv1alpha1.CommonSpec{
				TargetNamespace: targetNamepsace,
			},
		},
	}
	return client.TektonConfigs().Create(ctx, tektonConfig, metav1.CreateOptions{})
}
