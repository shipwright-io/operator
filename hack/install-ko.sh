#! /bin/bash

set -e

dest=${1:-bin}
mkdir -p "${dest}"
curl -fsL https://github.com/google/ko/releases/download/v0.8.1/ko_0.8.1_Linux_x86_64.tar.gz | tar xzf - -C "${dest}" ko
chmod +x "${dest}/ko"
