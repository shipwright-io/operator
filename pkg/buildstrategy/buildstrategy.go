package buildstrategy

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	"github.com/shipwright-io/operator/pkg/common"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
)

const clusterBuildStrategiesCRD = "clusterbuildstrategies.shipwright.io"

// ReconcileBuildStrategies reconciles the desired ClusterBuildStrategies to install on the cluster.
// Returns `true` if the build strategies were not installed and a requeue is required.
func ReconcileBuildStrategies(ctx context.Context, crdClient crdclientv1.ApiextensionsV1Interface, log logr.Logger, manifest manifestival.Manifest) (bool, error) {
	crdExists, err := common.CRDExist(ctx, crdClient, clusterBuildStrategiesCRD)
	if err != nil {
		return true, err
	}
	// If the CRD for Shipwright's cluster build strategies were not installed yet, the reconciler
	// should requeue.
	if !crdExists {
		return true, nil
	}
	// Apply the provided manifest containing the build strategies
	err = manifest.Apply()
	if err != nil {
		return true, err
	}
	return false, nil
}
