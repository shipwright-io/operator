#! /bin/bash

set -e
set -o pipefail

ENVTEST_ASSETS_DIR=$(pwd)/testbin
mkdir -p "${ENVTEST_ASSETS_DIR}"
test -f "${ENVTEST_ASSETS_DIR}/setup-envtest.sh" || curl -sSLo "${ENVTEST_ASSETS_DIR}/setup-envtest.sh" https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh

source "${ENVTEST_ASSETS_DIR}/setup-envtest.sh"

fetch_envtest_tools "${ENVTEST_ASSETS_DIR}"
setup_envtest_env "${ENVTEST_ASSETS_DIR}"
# Run tests sequentially - the controller integration tests cannot be run concurrently
go test ./... -coverprofile cover.out -p 1 -failfast -test.v -test.failfast
