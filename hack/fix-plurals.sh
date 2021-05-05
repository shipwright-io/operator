#!/bin/bash
#
# Script to work around https://github.com/operator-framework/operator-sdk/issues/4453
#

set -e

# allowing sed binary customization, so macos users can use an alternative gnu/sed, usually named "gsed"
SED_BIN="${SED_BIN:-sed}"

${SED_BIN} -i 's/shipwrightbuildren/shipwrightbuilds/' config/crd/bases/operator.shipwright.io_shipwrightbuildren.yaml
mv -f config/crd/bases/operator.shipwright.io_shipwrightbuildren.yaml config/crd/bases/operator.shipwright.io_shipwrightbuilds.yaml
