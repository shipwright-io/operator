#! /bin/sh

set -eu

KUBECTL_BIN=${KUBECTL_BIN:-kubectl}

echo "# Using KinD context..."
${KUBECTL_BIN} config use-context "kind-kind"
echo "# KinD nodes:"
${KUBECTL_BIN} get nodes

nodeStatus=$(${KUBECTL_BIN} get node kind-control-plane -o json | jq -r .'status.conditions[] | select(.type == "Ready") | .status')
if [ "${nodeStatus}" != "True" ]; then
    echo "# Node is not ready:"
    ${KUBECTL_BIN} describe node kind-control-plane
    echo "# Pods:"
    ${KUBECTL_BIN} get pod -A
    echo "# Events:"
    ${KUBECTL_BIN} get events -A

    exit 1
fi
