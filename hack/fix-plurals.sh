#! /bin/bash

# Script to work around https://github.com/operator-framework/operator-sdk/issues/4453

set -e

sed -i 's/shipwrightbuildren/shipwrightbuilds/' config/crd/bases/operator.shipwright.io_shipwrightbuildren.yaml
mv -f config/crd/bases/operator.shipwright.io_shipwrightbuildren.yaml config/crd/bases/operator.shipwright.io_shipwrightbuilds.yaml
