#!/usr/bin/env bash

set -euo pipefail

KO_BIN=${KO_BIN:-ko}
CONTAINER_ENGINE=${CONTAINER_ENGINE:-docker}
USERNAME=${REGISTRY_USERNAME:-""}
PASSWORD=${REGISTRY_PASSWORD:-""}

function login() {
    echo "Logging into container registry $IMAGE_HOST"
    echo "$PASSWORD" | ${KO_BIN} login -u "$USERNAME" --password-stdin "$IMAGE_HOST"
    echo "$PASSWORD" | ${CONTAINER_ENGINE} login -u "$USERNAME" --password-stdin "$IMAGE_HOST"
}

if [ -z "${IMAGE_HOST}" ]; then
    echo "Error: image host is not set"
    exit 1
fi

if [ -n "${USERNAME}" ] && [ -n "${PASSWORD}" ]; then
    login
else
    echo "Skipping registry login - build will rely on existing credentials"
fi

echo "Building and pushing the operator and bundle"
make bundle-push CONTAINER_ENGINE="${CONTAINER_ENGINE}" IMAGE_PUSH=true
