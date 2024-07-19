package utils

import (
	"fmt"
	"path/filepath"

	"github.com/manifestival/manifestival"
	"github.com/shipwright-io/operator/internal/pkg/common"
)

// ParseBuildStrategyNames returns a list of object names from the embedded build strategy
// manifests.
func ParseBuildStrategyNames() ([]string, error) {
	koDataPath, err := common.KoDataPath()
	if err != nil {
		return nil, err
	}
	strategyPath := filepath.Join(koDataPath, "samples", "buildstrategy")
	sampleNames := []string{}
	manifest, err := manifestival.ManifestFrom(manifestival.Recursive(strategyPath))
	if err != nil {
		return sampleNames, err
	}
	for _, obj := range manifest.Resources() {
		if obj.GetKind() == "ClusterBuildStrategy" {
			sampleNames = append(sampleNames, obj.GetName())
		}

	}
	if len(sampleNames) == 0 {
		return sampleNames, fmt.Errorf("no ClusterBuildStrategy objects found in %s", strategyPath)
	}
	return sampleNames, nil
}
