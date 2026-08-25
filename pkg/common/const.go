package common

const (
	koDataPathEnv         = "KO_DATA_PATH"
	ShipwrightImagePrefix = "IMAGE_SHIPWRIGHT_"

	// OperatorNamespaceEnvVar holds the operator pod's namespace.
	OperatorNamespaceEnvVar = "POD_NAMESPACE"
	// DefaultNamespace is the fallback namespace when POD_NAMESPACE is unset.
	DefaultNamespace = "shipwright-build"

	TektonOpMinSupportedVersion = "v0.50.0"
	TektonOpMinSupportedMajor   = 0
	TektonOpMinSupportedMinor   = 50

	Retain int = iota
	Overwrite
)
