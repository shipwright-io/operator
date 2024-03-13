#!/usr/bin/env bash
#
# Installs ko in the informed location, by downloading and extracting tarball in place.
#
# $ install-ko.sh "bin/ko"
#

set -e

DEST="${1:-bin/ko}"
KO_VERSION="${KO_VERSION:-0.15.2}"

OS="${OS:-linux}"
ARCH="${ARCH:-amd64}"

# making amendments to the way ko releases are named, more architectures may be added later on
if [ $ARCH == "amd64" ] ; then
    ARCH="x86_64"
fi
# operational system name is capitalized
OS="${OS^}"

KO_URL_HOST="${KO_HOST:-github.com}"
KO_URL_PATH="${KO_URL_PATH:-google/ko/releases/download}"
KO_URL="https://${KO_URL_HOST}/${KO_URL_PATH}/v${KO_VERSION}/ko_${KO_VERSION}_${OS}_${ARCH}.tar.gz"

if [ -x ${DEST} ] ; then
    echo "# ko is already installed at '${DEST}'"
    exit 0
fi

BASE_DIR="$(dirname ${DEST})"

curl --fail --silent --location "${KO_URL}" \
    |tar xzf - -C ${BASE_DIR} ko
chmod +x "${DEST}"
