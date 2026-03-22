package triggers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	"github.com/shipwright-io/operator/pkg/common"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
)

const buildsCRD = "builds.shipwright.io"

// ReconcileTriggers reconciles the desired Triggers deployment on the cluster.
// Returns `true` if triggers were not installed and a requeue is required.
func ReconcileTriggers(ctx context.Context, crdClient crdclientv1.ApiextensionsV1Interface, log logr.Logger, manifest manifestival.Manifest) (bool, error) {
	crdExists, err := common.CRDExist(ctx, crdClient, buildsCRD)
	if err != nil {
		return true, err
	}
	// If the CRD for Shipwright's builds is not installed yet, the reconciler should requeue.
	if !crdExists {
		return true, nil
	}
	// Apply the provided manifest containing the triggers resources
	err = manifest.Apply()
	if err != nil {
		return true, err
	}
	return false, nil
}
