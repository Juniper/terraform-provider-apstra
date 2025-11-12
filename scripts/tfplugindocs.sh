#!/usr/bin/env bash

# Default values
check=false

# Loop over all arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --check)
            check=true
            ;;
        *)
            echo "Unknown argument: $1" >&2
            exit 1
            ;;
    esac
    shift
done

(cd tools/tfplugindocs && go tool tfplugindocs generate --provider-dir ../.. || exit 1)

if [ "$check" == "false" ]; then
  exit 0
fi

grep -E '^subcategory: ""$' docs/data-sources/*        && echo "missing subcategory" && exit 1
grep -E '^subcategory: ""$' docs/resources/*           && echo "missing subcategory" && exit 1
grep -E '^subcategory: ""$' docs/ephemeral-resources/* && echo "missing subcategory" && exit 1

git update-index --refresh || exit 1

git diff-index --quiet HEAD -- || exit 1

