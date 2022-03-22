# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.8.0

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the container image host, namespace, and part of the image name for remote
# images. This variable is used to construct full image tags for bundle and catalog images.
# IMAGE_HOST defines the host registry, defaults to GitHub's container registry (ghcr.io)
# IMAGE_NAMEPSACE defines the location where images are organized for a user - this can sometimes
# be called an "organization."
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# ghcr.io/shipwright-io/operator/operator-bundle:$VERSION and
# ghcr.io/shipwright-io/operator/operator-catalog:$VERSION.
IMAGE_HOST ?= ghcr.io
IMAGE_NAMESPACE ?= shipwright-io/operator
IMAGE_REPO ?= $(IMAGE_HOST)/$(IMAGE_NAMESPACE)
IMAGE_TAG_BASE ?= $(IMAGE_REPO)/operator

# TAG allows the tag for the operator image to be changed. Defaults to the VERSION
TAG ?= $(VERSION)

# BUNDLE_IMG defines the image:tag used for the bundle. 
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):$(TAG)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

CONTAINER_ENGINE ?= docker
IMAGE_PUSH ?= true

KUBECTL_BIN ?= kubectl
SED_BIN ?= sed

# operating-system type and architecture based on golang
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: verify-manifests 
verify-manifests: manifests ## Verify manifests were generated and committed to git
	hack/check-git-status.sh manifests

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: verify-generate
verify-generate: generate ## Verify code was generated and git status is clean
	hack/check-git-status.sh generate

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: verify-fmt
verify-fmt: fmt ## Verify formatting and ensure git status is clean
	hack/check-git-status.sh fmt

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

BINDATA = $(shell pwd)/cmd/operator/kodata
.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KO_DATA_PATH=${BINDATA} KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out -p 1 -failfast -test.v -test.failfast

##@ Build

.PHONY: build
build: generate fmt vet ## Build operator binary.
	go build -o bin/operator ./cmd/operator

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/operator

# ko-publish replaces the "docker-build" and "docker-push" targets for operator-sdk projects.
.PHONY: ko-publish
ko-publish: ko ## Build and push the image with ko.
	KO_DOCKER_REPO=${IMAGE_REPO} $(KO) publish --base-import-paths --push=${IMAGE_PUSH} -t ${TAG} ./cmd/operator

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL_BIN) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL_BIN) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL_BIN) apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL_BIN) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: clean
clean: ## Cleans out all downloaded dependencies for development and testing
	rm -rf bin
	rm -rf testbin
	rm -rf _output

.PHONY: bin-dir
bin-dir: ## Creates a local "bin" directory for helper applications.
	@mkdir ./bin || true

KO = $(shell pwd)/bin/ko
.PHONY: ko
ko: bin-dir ## Installs ko locally if necessary.
	OS=${OS} ARCH=${ARCH} hack/install-ko.sh $(KO)

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)


OPERATOR_SDK = $(shell pwd)/bin/operator-sdk
.PHONY: operator-sdk
operator-sdk: bin-dir ## Installs operator-sdk
	OS=${OS} ARCH=${ARCH} hack/install-operator-sdk.sh $(OPERATOR_SDK)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests --interactive=false -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: verify-bundle
verify-bundle: bundle ## Verify bundle manifests were generated and committed to git
	hack/check-git-status.sh bundle

.PHONY: bundle-build
bundle-build: bundle ## Build the bundle image.
	$(CONTAINER_ENGINE) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: bundle-build ## Push the bundle image to the registry
	$(CONTAINER_ENGINE) push $(BUNDLE_IMG)

.PHONY: release
release: ko ## Build and push the full release
	CONTAINER_ENGINE="$(CONTAINER_ENGINE)" KO_BIN="$(KO)" IMAGE_HOST=${IMAGE_HOST} IMAGE_NAMESPACE=${IMAGE_NAMESPACE} TAG=${TAG} hack/release.sh

.PHONY: install-olm
install-olm: operator-sdk ## Install OLM on the current cluster
	$(OPERATOR_SDK) olm install

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.21.0/$(OS)-$(ARCH)-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# Use HTTP for opm registry operations
OPM_USE_HTTP ?= false

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding a bundle image to a candidate file-based OLM catalog.
.PHONY: catalog-build
catalog-build: opm ## Build a file-based OLM catalog image containing a candidate operator bundle.
	BUNDLE_IMG=$(BUNDLE_IMG) OPM_BIN=$(OPM) SED_BIN=$(SED_BIN) CSV_VERSION=$(VERSION) USE_HTTP=$(OPM_USE_HTTP) hack/render-candidate-catalog.sh
	$(CONTAINER_ENGINE) build -f _output/catalog.Dockerfile -t $(CATALOG_IMG) _output

.PHONY: catalog-push
catalog-push: catalog-build ## Build and push an OLM catalog image with a candidate operator bundle.
	$(CONTAINER_ENGINE) push $(CATALOG_IMG)

##@ CI Testing

CATALOG_NAMESPACE ?= shipwright-operator

.PHONY: catalog-run
catalog-run: ## Run the operator from a catalog image, using an OLM subscription
	CATALOG_IMG=$(CATALOG_IMG) CSV_VERSION=$(VERSION) KUBECTL_BIN=$(KUBECTL_BIN) NAMESPACE=$(CATALOG_NAMESPACE) SED_BIN=$(SED_BIN) hack/run-operator-catalog.sh

.PHONY: verify-kind
verify-kind:
	KUBECTL_BIN=$(KUBECTL_BIN) test/kind/verify-kind.sh

.PHONY: deploy-kind-registry
deploy-kind-registry:
	CONTAINER_ENGINE=$(CONTAINER_ENGINE) KUBECTL_BIN=$(KUBECTL_BIN) test/kind/deploy-registry.sh

.PHONY: deploy-kind-registry-post
deploy-kind-registry-post:
	CONTAINER_ENGINE=$(CONTAINER_ENGINE) KUBECTL_BIN=$(KUBECTL_BIN) test/kind/deploy-registry-post.sh