#! /bin/bash

set -x
set -e

fixCommand="$*"

if [[ -n $(git status --porcelain) ]]; then
    echo "Verify check failed - run \`make ${fixCommand}\` and commit the resulting changes to git."
    echo "Differences:"
    git diff
    exit 1
fi