# VERSION defines the project version for the bundle. 
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.2.0

# CHANNELS define the bundle channels used in the bundle. 
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
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

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTAINER_ENGINE ?= docker
IMAGE_REPO ?= quay.io/shipwright
TAG ?= $(VERSION)
IMAGE_PUSH ?= true

IMAGE_TAG_BASE ?= $(IMAGE_REPO)/operator

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):$(TAG)

# BUNDLE_IMG defines the image:tag used for the bundle. 
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:$(TAG)

# operating-system type and architecture based on golang
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

KUBECTL_BIN ?= kubectl
SED_BIN ?= sed

all: operator

build: operator

clean:
	rm -rf bin
	rm -rf testbin

# Run tests
BINDATA = $(shell pwd)/cmd/operator/kodata
test: generate fmt vet manifests
	KO_DATA_PATH=${BINDATA} hack/test-with-envtest.sh

# Build manager binary
operator: generate fmt vet
	go build -o bin/operator ./cmd/operator

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./cmd/operator

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | $(KUBECTL_BIN) apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | $(KUBECTL_BIN) delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller="$(IMG)"
	$(KUSTOMIZE) build config/default | $(KUBECTL_BIN) apply -f -

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | $(KUBECTL_BIN) delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Verify manifests were generated and committed to git
verify-manifests: manifests
	hack/check-git-status.sh manifests

# Run go fmt against code
fmt:
	go fmt ./...

# Verify formatting and ensure git status is clean
verify-fmt: fmt
	hack/check-git-status.sh fmt

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Verify code was generated and git status is clean
verify-generate: generate
	hack/check-git-status.sh generate

# Creates a local "bin" directory for helper applications.
bin-dir:
	@mkdir ./bin || true

# Installs ko on the specified location
KO = $(shell pwd)/bin/ko
ko: bin-dir
	OS=${OS} ARCH=${ARCH} hack/install-ko.sh $(KO)

# Build and push the image with ko
ko-publish: ko
	KO_DOCKER_REPO=${IMAGE_REPO} $(KO) publish --base-import-paths --push=${IMAGE_PUSH} -t ${TAG} ./cmd/operator

# Download controller-gen locally if necessary
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1)

# Download kustomize locally if necessary
KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# Installs operator-sdk on specified location
OPERATOR_SDK = $(shell pwd)/bin/operator-sdk
operator-sdk: bin-dir
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

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize operator-sdk
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller="${IMG}"
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle


# Verify bundle manifests were generated and committed to git
verify-bundle: bundle
	hack/check-git-status.sh bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build: bundle
	$(CONTAINER_ENGINE) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

# Push the bundle image to the registry
.PHONY: bundle-push
bundle-push: bundle-build
	$(CONTAINER_ENGINE) push $(BUNDLE_IMG)

# Install OLM on the current cluster
.PHONY: install-olm
install-olm: operator-sdk
	$(OPERATOR_SDK) olm install

.PHONY: opm
OPM = ./bin/opm
opm:
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$(OS)-$(ARCH)-opm ;\
	chmod +x $(OPM) ;\
	}
else 
OPM = $(shell which opm)
endif
endif

BUNDLE_IMGS ?= $(BUNDLE_IMG) 
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:$(VERSION)

#
# ifneq ($(origin CATALOG_BASE_IMG), undefined) 
# FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
# endif
# $(OPM) index add --container-tool $(CONTAINER_ENGINE) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

CATALOG_INDEX_IMG ?= quay.io/operatorhubio/catalog:latest

# Build a catalog image with the operator bundle included
.PHONY: catalog-build
catalog-build: opm
	$(OPM) index add --container-tool $(CONTAINER_ENGINE) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) --from-index=$(CATALOG_INDEX_IMG)
	
# Build and push a catalog image with the operator bundle to a container registry
.PHONY: catalog-push
catalog-push: catalog-build
	$(CONTAINER_ENGINE) push $(CATALOG_IMG)


CATALOG_NAMESPACE ?= shipwright-operator

# Run the operator from a catalog image, using an OLM subscription
.PHONY: catalog-run
catalog-run:
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