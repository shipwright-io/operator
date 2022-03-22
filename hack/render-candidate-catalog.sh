#! /usr/bin/env bash

# Copyright The Shipwright Contributors
#
# SPDX-License-Identifier: Apache-2.0

# render-candidate-catalog.sh creates a file-based OLM catalog containing the following:
#
# 1. The candidate shipwright-operator image (from a local/CI build)
# 2. The existing Tekton and Shipwright operators that are released on OperatorHub.io
#
# The core of the catalog are JSON manifests contained in `test/catalog`.
# This script copies those manifests to _output/catalog, then does a bit of opm and sed
# manipulation to create catalog entries for the candidate operator deployment.

set -euo pipefail

echo "Rendering candidate operator catalog"

YQ_BIN=${YQ_BIN:-${PWD}/bin/yq}
OPM_BIN=${OPM_BIN:-${PWD}/bin/opm}
SED_BIN=${SED_BIN:-sed}
CSV_VERSION=${CSV_VERSION:-${VERSION}}
USE_HTTP=${USE_HTTP:-false}

channelSpec="_output/catalog/candidate/shipwright-operator-channel-candidate.json"
bundleSpec="_output/catalog/candidate/shipwright-operator-bundle.json"

# Initialize the new catalog
echo "Copying catalog manifests"
rm -f "_output/catalog.Dockerfile"
rm -rf "_output/catalog"
mkdir -p _output
cp -r test/catalog _output/catalog

echo "Rendering bundle image into the candidate catalog"
# Replace the placeholder bundle name for the "candidate" OLM subscription channel
${SED_BIN} -i -E 's|"name": "shipwright-operator-latest"|"name": "shipwright-operator.v'"${CSV_VERSION}"'"|g' "${channelSpec}"
# Render the OLM content from the existing bundle image (for local testing/CI)
${OPM_BIN} render "${BUNDLE_IMG}" --use-http="${USE_HTTP}"> ${bundleSpec}

echo "Generating catalog Dockerfile"
${OPM_BIN} generate dockerfile "_output/catalog"

echo "Done"
