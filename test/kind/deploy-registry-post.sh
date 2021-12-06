#!/bin/sh

set -eu

DOCKER="${CONTAINER_ENGINE:-docker}"
KUBECTL_BIN="${KUBECTL_BIN:-kubectl}"

echo "Running KinD registry post-install actions"

reg_name='kind-registry'

echo "Connecting registry ${reg_name} to kind network"
# connect the registry to the cluster network
# (the network may already be connected)
${DOCKER} network connect "kind" "${reg_name}" || true

echo "Registry connected to kind network"

echo "Publishing local container registry on the KinD cluster"

${KUBECTL_BIN} apply -f test/kind/local-registry-cm.yaml

echo "Done"
