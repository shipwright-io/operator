#!/bin/sh

set -eu

DOCKER="${CONTAINER_ENGINE:-DOCKER}"

# create registry container unless it already exists
REG_NAME='kind-registry'
REG_PORT='5000'
echo "Deploying container registry ${REG_NAME}:${REG_PORT}"
running="$(${DOCKER} inspect -f '{{.State.Running}}' "${REG_NAME}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  ${DOCKER} run \
    -d --restart=always -p "127.0.0.1:${REG_PORT}:5000" --name "${REG_NAME}" \
    registry:2
fi

echo "Container registry ${REG_NAME} deployed"
