package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/shipwright-io/operator/pkg/common"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileBuildStrategy(ctx context.Context, crdClient crdclientv1.ApiextensionsV1Interface, c client.Client, logger logr.Logger, targetNamespace string) (bool, error) {
	_, err := IsBuildStrategyCRDsInstalled(ctx, crdClient)
	if err != nil {
		return true, err
	}

	err = common.WaitForPodRunning(ctx, c, client.MatchingLabels{"name": "shp-build-webhook"}, logger, targetNamespace)
	if err != nil {
		return true, err
	}

	dataPaths, err := GetStrategiesPaths()
	if err != nil {
		return false, fmt.Errorf("Error retreiving data paths for buildstrategy")
	}

	manifest, err := common.SetupManifestival(c, dataPaths, logger)
	if err != nil {
		return false, fmt.Errorf("Error creating inital strategy manifest")
	}

	if err = manifest.Apply(); err != nil {
		return true, err
	}

	return false, nil
}

func IsBuildStrategyCRDsInstalled(ctx context.Context, crdClient crdclientv1.ApiextensionsV1Interface) (bool, error) {
	ok, err := common.CRDExist(ctx, crdClient, "clusterbuildstrategies.shipwright.io")
	if err != nil {
		return true, err
	}
	if ok {
		_, err = common.CRDExist(ctx, crdClient, "buildstrategies.shipwright.io")
		if err != nil {
			return true, err
		}
	}

	return false, nil
}

func GetStrategiesPaths() ([]string, error) {
	dataPath, err := common.KoDataPath()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filepath.Join(dataPath, common.BuildStrategyDataDir))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list, err := file.Readdirnames(0) // 0 to read all files and folders
	if err != nil {
		return nil, err
	}
	for i, v := range list {
		list[i] = fmt.Sprintf("%s/%s", common.BuildStrategyDataDir, v)
	}
	return list, nil
}
