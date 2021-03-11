#! /bin/bash

set -e

dest=${1:-bin/operator-sdk}
curl -sL -o "${dest}" https://github.com/operator-framework/operator-sdk/releases/download/v1.4.2/operator-sdk_linux_amd64
chmod +x "${dest}"
