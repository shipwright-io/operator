package common

import (
	"time"
)

const (
	koDataPathEnv         = "KO_DATA_PATH"
	ShipwrightImagePrefix = "IMAGE_SHIPWRIGHT_"

	TektonOpMinSupportedVersion = "v0.50.0"
	TektonOpMinSupportedMajor   = 0
	TektonOpMinSupportedMinor   = 50

	Retain int = iota
	Overwrite

	CertificateDataDir   = "certificates"
	BuildDataDir         = "build"
	BuildStrategyDataDir = "buildstrategy"

	APICallRetryInterval = 500 * time.Millisecond
	Timeout              = 300 * time.Second
)
