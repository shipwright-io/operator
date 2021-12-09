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
