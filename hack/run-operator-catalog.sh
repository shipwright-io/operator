#!/usr/bin/env bash

# hack/run-operator-catalog.sh
#
# Run the operator from a catalog image.
# Required environment variables:
#
# - CATALOG_IMG: catalog image to deploy
# - CSV_VERSION: version tag of the cluster service version
#
# Optional environment variables
#
# - KUSTOMIZE_BIN: path to kustomize
# - KUBECTL_BIN: path to kubectl (or equivalent command line)
# - SED_BIN: path to GNU sed
# - CATALOG_NAMESPACE: Namespace to deploy the catalog. Defaults to shipwright-operator.
# - SUBSCRIPTION_NAMESPACE: Namespace to install the operator via an OLM subscription. Defaults to
#   shipwright-operator.
# - NAME_PREFIX: prefix to use for all resource names. Defaults to "shipwright-"
# - TEKTON_OPERATOR_VERSION: Tekton Operator version to install. Defaults to v0.77.0 (matches go.mod).
# - CERT_MANAGER_VERSION: cert-manager version to install. Defaults to v1.13.0.

set -eu -o pipefail

KUSTOMIZE_BIN=${KUSTOMIZE_BIN:-${PWD}/bin/kustomize}
KUBECTL_BIN=${KUBECTL_BIN:-kubectl}
SED_BIN=${SED_BIN:-sed}
CATALOG_NAMESPACE=${CATALOG_NAMESPACE:-shipwright-operator}
SUBSCRIPTION_NAMESPACE=${SUBSCRIPTION_NAMESPACE:-shipwright-operator}
NAME_PREFIX=${NAME_PREFIX:-shipwright-}

if [[ -z ${CATALOG_IMG} ]]; then
    echo "CATALOG_IMG environment variable must be set"
    exit 1
fi

if [[ -z ${CSV_VERSION} ]]; then
    echo "CSV_VERSION environment variable must be set"
    exit 1
fi

function add_kustomizations() {
    echo "Adding replacements not supported by kustomize"
    ${SED_BIN} -i -E "s|image: (.+)$|image: ${CATALOG_IMG}|g" config/catalog/catalog_source.yaml
    ${SED_BIN} -i -E "s|startingCSV: (.+)$|startingCSV: shipwright-operator.v${CSV_VERSION}|g" config/subscription/subscription.yaml
    ${SED_BIN} -i -E "s|sourceNamespace: (.+)$|sourceNamespace: ${CATALOG_NAMESPACE}|g" config/subscription/subscription.yaml
    ${SED_BIN} -i -E "s|source: (.+)$|source: ${NAME_PREFIX}operator|g" config/subscription/subscription.yaml

    echo "Applying catalog source and subscription from kustomize"

    pushd config/catalog
    ${KUSTOMIZE_BIN} edit set namespace "${CATALOG_NAMESPACE}"
    ${KUSTOMIZE_BIN} edit set nameprefix "${NAME_PREFIX}"
    popd

    pushd config/subscription
    ${KUSTOMIZE_BIN} edit set namespace "${SUBSCRIPTION_NAMESPACE}"
    ${KUSTOMIZE_BIN} edit set nameprefix "${NAME_PREFIX}"
    popd
}

function dump_state() {
    echo "Dumping OLM catalog sources"
    ${KUBECTL_BIN} get catalogsources -n "${CATALOG_NAMESPACE}" -o yaml
    echo "Dumping OLM subscriptions"
    ${KUBECTL_BIN} get subscriptions -n "${SUBSCRIPTION_NAMESPACE}" -o yaml
    echo "Dumping OLM installplans"
    ${KUBECTL_BIN} get installplans -n "${SUBSCRIPTION_NAMESPACE}" -o yaml
    echo "Dumping OLM CSVs"
    ${KUBECTL_BIN} get clusterserviceversions -n "${SUBSCRIPTION_NAMESPACE}" -o yaml
    echo "Dumping pods"
    ${KUBECTL_BIN} get pods -n "${SUBSCRIPTION_NAMESPACE}" -o yaml
    echo "${CATALOG_NAMESPACE} -ne ${SUBSCRIPTION_NAMESPACE}"
    if [ "${CATALOG_NAMESPACE}" != "${SUBSCRIPTION_NAMESPACE}" ]; then
        ${KUBECTL_BIN} get pods -n "${CATALOG_NAMESPACE}" -o yaml
    fi
}

function wait_for_pod() {
    label=$1
    namespace=$2
    timeout=$3
    ${KUBECTL_BIN} wait --for=condition=Ready pod -l "${label}" -n "${namespace}" --timeout "${timeout}"
}

function install_tekton() {
    echo "Installing Tekton Operator"
    # Install Tekton Operator using the official installation method (https://github.com/tektoncd/operator/releases)
    # Minimum supported version is v0.50.0 (see pkg/common/const.go)
    # Default version matches the dependency in go.mod (v0.77.0) for stability
    # Can be overridden via TEKTON_OPERATOR_VERSION env var
    TEKTON_OPERATOR_VERSION=${TEKTON_OPERATOR_VERSION:-v0.77.0}
    echo "Installing Tekton Operator version ${TEKTON_OPERATOR_VERSION}"
    if ! ${KUBECTL_BIN} apply -f "https://storage.googleapis.com/tekton-releases/operator/previous/${TEKTON_OPERATOR_VERSION}/release.yaml"; then
        echo "Warning: Failed to install Tekton Operator ${TEKTON_OPERATOR_VERSION}, falling back to latest"
        ${KUBECTL_BIN} apply -f "https://storage.googleapis.com/tekton-releases/operator/latest/release.yaml"
    fi
    
    echo "Waiting for Tekton Operator to be ready"
    # Wait for the operator namespace to exist first
    for i in {1..30}; do
        if ${KUBECTL_BIN} get namespace tekton-operator &>/dev/null; then
            break
        fi
        sleep 2
    done
    
    if ! wait_for_pod "app=tekton-operator" "tekton-operator" 5m; then
        echo "Failed to deploy Tekton Operator"
        ${KUBECTL_BIN} get pods -n tekton-operator
        exit 1
    fi
    
    echo "Creating TektonConfig instance"
    # Check if TektonConfig already exists, delete it if it does to avoid annotation warnings
    if ${KUBECTL_BIN} get tektonconfigs.operator.tekton.dev/config &>/dev/null; then
        echo "TektonConfig already exists, deleting it to recreate cleanly"
        ${KUBECTL_BIN} delete tektonconfigs.operator.tekton.dev/config --ignore-not-found=true
        # Wait a moment for deletion to complete
        sleep 2
    fi
    
    ${KUBECTL_BIN} apply -f - <<EOF
apiVersion: operator.tekton.dev/v1alpha1
kind: TektonConfig
metadata:
  name: config
spec:
  profile: lite
  targetNamespace: tekton-pipelines
EOF
    
    echo "Waiting for TektonConfig to be ready"
    if ! ${KUBECTL_BIN} wait --for=condition=Ready tektonconfigs.operator.tekton.dev/config --timeout=5m; then
        echo "Failed to deploy TektonConfig"
        ${KUBECTL_BIN} get tektonconfigs -o yaml
        ${KUBECTL_BIN} get tektonconfigs config -o yaml
        exit 1
    fi
}

function install_cert_manager() {
    echo "Installing cert-manager"
    # Install cert-manager using the official installation method (https://github.com/cert-manager/cert-manager/releases)
    # Default version is v1.13.0 (stable release)
    # Can be overridden via CERT_MANAGER_VERSION env var
    CERT_MANAGER_VERSION=${CERT_MANAGER_VERSION:-v1.13.0}
    echo "Installing cert-manager version ${CERT_MANAGER_VERSION}"
    ${KUBECTL_BIN} apply -f "https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml"
    
    echo "Waiting for cert-manager pods to be ready"
    if ! wait_for_pod "app.kubernetes.io/instance=cert-manager" "cert-manager" 5m; then
        echo "Failed to deploy cert-manager"
        ${KUBECTL_BIN} get pods -n cert-manager
        exit 1
    fi
    
    echo "Waiting for cert-manager webhook to be ready"
    # Wait for the webhook to be ready by checking if it can validate certificates
    for i in {1..30}; do
        if ${KUBECTL_BIN} get validatingwebhookconfigurations cert-manager-webhook &>/dev/null; then
            echo "cert-manager webhook is ready"
            return 0
        fi
        sleep 2
    done
    echo "Warning: cert-manager webhook may not be fully ready, continuing anyway"
}

add_kustomizations

echo "Deploying catalog source"
${KUSTOMIZE_BIN} build config/catalog | ${KUBECTL_BIN} apply -f -

echo "Waiting for the catalog source to be ready"
# Wait 15 seconds for the catalog source pod to be created first.
sleep 15
if ! wait_for_pod "olm.catalogSource=${NAME_PREFIX}operator" "${CATALOG_NAMESPACE}" 1m; then
    echo "Failed to deploy catalog source, dumping operator state"
    dump_state
    exit 1
fi

echo "Deploying subscription"
${KUSTOMIZE_BIN} build config/subscription | ${KUBECTL_BIN} apply -f -

echo "Waiting for the operator to be ready"
# Wait 60 seconds for the operator pod to be created
# This needs extra time due to OLM's reconciliation process
sleep 60
if ! wait_for_pod "app=shipwright-operator" "${SUBSCRIPTION_NAMESPACE}" 5m; then
    echo "Failed to deploy, dumping operator state"
    dump_state
    exit 1
fi

echo "Installing prerequisites"
install_cert_manager
install_tekton

echo "Deploying Shipwright build controller"

${KUBECTL_BIN} apply -f - <<EOF
kind: ShipwrightBuild
apiVersion: operator.shipwright.io/v1alpha1
metadata:
  name: shipwright
spec:
    targetNamespace: ${SUBSCRIPTION_NAMESPACE}
EOF

if ! ${KUBECTL_BIN} wait --for=condition=Ready=true shipwrightbuilds.operator.shipwright.io/shipwright --timeout=5m; then
    echo "Failed to deploy ShipwrightBuild - dumping ShipwrightBuild state"
    ${KUBECTL_BIN} get shipwrightbuilds -o yaml
    dump_state
    exit 1
fi

exit 0
