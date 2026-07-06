module github.com/shipwright-io/operator

go 1.26.3

require (
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/go-logr/logr v1.4.3
	github.com/manifestival/controller-runtime-client v0.4.0
	github.com/manifestival/manifestival v0.7.2
	github.com/onsi/ginkgo/v2 v2.31.0
	github.com/onsi/gomega v1.42.0
	github.com/shipwright-io/build v0.20.0
	github.com/tektoncd/operator v0.77.0
	go.yaml.in/yaml/v3 v3.0.4
	k8s.io/api v0.36.1
	k8s.io/apiextensions-apiserver v0.36.1
	k8s.io/apimachinery v0.36.1
	// go mod tidy forces this to v1.5.2
	k8s.io/client-go v1.5.2
	knative.dev/pkg v0.0.0-20260318013857-98d5a706d4fd
	sigs.k8s.io/controller-runtime v0.24.1
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/jsonreference v0.21.4 // indirect
	github.com/go-openapi/swag v0.25.4 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.4 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/fileutils v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/mangling v0.25.4 // indirect
	github.com/go-openapi/swag/netutils v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/openshift-pipelines/pipelines-as-code v0.36.0 // indirect
	github.com/openshift-pipelines/tektoncd-pruner v0.0.0-20250711075231-9c8624123820 // indirect
	github.com/openshift/api v0.0.0-20260302174620-dcac36b908db // indirect
	github.com/openshift/apiserver-library-go v0.0.0-20230816171015-6bfafa975bfb // indirect
	github.com/openshift/client-go v0.0.0-20251015124057-db0dee36e235 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/tektoncd/pipeline v1.12.0 // indirect
	github.com/tektoncd/triggers v0.32.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.65.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/term v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260401024825-9d38bb4040a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
	google.golang.org/grpc v1.80.0 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	k8s.io/utils v0.0.0-20260210185600-b8788abfbbc2 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

// Go modules at times does not effectively resolve transitive dependencies that share
// k8s.io/* packages. These should be kept internally consistent with a "common" k8s version - this
// is more of an art than a science, and can change with successive dependency updates.
//
// Each replace here should provide a comment explaining why it is warranted, and its necessity
// should be tested with successive dependency updates.

// k8s.io/client-go has a "latest" semantic version of v1.5.2, which predates the project
// go module version standardization (v0.y.z, with y and z representing the k8s 1.y.z minor/patch
// versions). `go mod tidy` will often overwrite the desired client-go version to v1.5.2, so we
// pin the version here.
replace k8s.io/client-go => k8s.io/client-go v0.36.1

// openshift packages pulled in by tektoncd/operator use structured-merge-diff/v4, which is
// incompatible with k8s 1.34+ (structured-merge-diff/v6). Pin to newer versions that use v6.
// The openshift @latest targets k8s 1.36, so we also pin k8s.io/api and k8s.io/apimachinery
// to prevent MVS from upgrading them beyond 1.36.
//
// knative.dev/pkg @latest removed metrics and tracing/config packages that tektoncd/triggers
// still imports transitively. Pin to the version compatible with our dependency set.
replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20260302174620-dcac36b908db
	github.com/openshift/apiserver-library-go => github.com/openshift/apiserver-library-go v0.0.0-20260123124658-a67c1b7813a4
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20260302182750-20813ce71ca6
	k8s.io/api => k8s.io/api v0.36.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.36.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.36.1
	knative.dev/pkg => knative.dev/pkg v0.0.0-20251224022817-f2b9f9c3f2ca
)
