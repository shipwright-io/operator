#!/bin/bash

set -eu -o pipefail

# This script generates a pull request to add the release to operatorhub.io
# Prerequisites:
# - The user has cloned and forked the operator catalog repository
# - The machine running this script has crane installed: https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md

# Environment variables to tune the script's behavior
# - OPERATORHUB_DIR: The local path to the operator catalog repository.
# - VERSION: The version of the operator to release.
# - HUB_REPO: A regular expression to match the GitHub org/repository of the catalog. Note that this
#   is a regular expression, so special characters need to be escaped.

OPERATORHUB_DIR=${OPERATORHUB_DIR:-$HOME/go/src/github.com/k8s-operatorhub/community-operators}
VERSION=${VERSION:-latest}
HUB_REPO=${HUB_REPO:-"k8s-operatorhub\/community-operators"}
# Regular expression match [https://github.com/|git@github.com:]
hubRegEx="(https:\/\/github\.com\/|git@github\.com:)${HUB_REPO}"

echo "Preparing to release Shipwright Operator ${VERSION} to Operator catalog github.com/${HUB_REPO}"

if [[ ! -d "$OPERATORHUB_DIR" ]]; then
  echo "Please clone the operator catalog repository repository to $OPERATORHUB_DIR"
  exit 1
fi

pushd "$OPERATORHUB_DIR"

originURL=$(git remote get-url origin)
if [[ "$originURL" =~ ${hubRegEx} ]]; then
  echo "Please set the origin remote to your fork of the operator catalog repository"
  exit 1
fi

upstreamURL=$(git remote get-url upstream)
if [[ ! "$upstreamURL" =~ ${hubRegEx} ]]; then
  echo "Please set the upstream remote ${upstreamURL} to the operator catalog repository"
  exit 1
fi

git fetch
git switch main
git pull upstream main
git checkout -b "shipwright-${VERSION}"

mkdir -p "operators/shipwright-operator/${VERSION}"
pushd "operators/shipwright-operator/${VERSION}"

crane export "ghcr.io/shipwright-io/operator/operator-bundle:v${VERSION}" - | tar -xv

popd

# Commit and push changes to our GitHub fork
git add "operators/shipwright-operator/${VERSION}"
git commit -m "Update Shipwright Operator to ${VERSION}" -s
git push --set-upstream origin "shipwright-${VERSION}"

popd
