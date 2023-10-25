package common

const (
	koDataPathEnv         = "KO_DATA_PATH"
	ShipwrightImagePrefix = "IMAGE_SHIPWRIGHT_"

	TektonOpMinSupportedVersion = "v0.50.0"
	TektonOpMinSupportedMajor   = 0
	TektonOpMinSupportedMinor   = 50

	Retain int = iota
	Overwrite
)
