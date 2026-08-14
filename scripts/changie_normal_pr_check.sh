#!/usr/bin/env bash

set -euo pipefail

BASE_REF="${1:?usage: $0 <base-ref>}"
echo "Checking normal PR against ${BASE_REF}..."

CHANGIE_CFG='.changie.yaml'

fail() {
    echo "ERROR: $*" >&2
    exit 1
}


CHANGES_DIR="$(yq '.changesDir' < "$CHANGIE_CFG")"
[ -n "$CHANGES_DIR" ] || fail "changesDir is not configured in $CHANGIE_CFG"

UNRELEASED_DIR="$CHANGES_DIR/$(yq '.unreleasedDir' < "$CHANGIE_CFG")"
[ -n "$UNRELEASED_DIR" ] || fail "unreleasedDir is not configured in $CHANGIE_CFG"

CHANGELOG="$(yq '.changelogPath' < "$CHANGIE_CFG")"
[ -n "$CHANGELOG" ] || fail "changelogPath is not configured in $CHANGIE_CFG"

new_fragment_path=""
new_fragment_count=0

while IFS=$'\t' read -r status path; do
    case "$status:$path" in
        A:"$UNRELEASED_DIR/"*)
           new_fragment_path="$path"
           new_fragment_count=$((new_fragment_count + 1)) # Cannot use ((new_fragment_count++)) with set -e
           ;;

        *:"$UNRELEASED_DIR/"*)
           fail "normal PR modifies an existing unreleased fragment: $path"
           ;;

        *:"$CHANGES_DIR/"*)
           fail "normal PR modifies a Changie file: $path"
           ;;

        *:"$CHANGELOG")
            fail "normal PR modifies $CHANGELOG"
            ;;
    esac
done < <(git diff --name-status "$BASE_REF"...HEAD)

if [ "$new_fragment_count" -ne 1 ]; then
    fail "normal PR must add exactly one change fragment under $UNRELEASED_DIR/ (found $new_fragment_count)"
fi

echo "Validating $new_fragment_path with changie..."

# This check ensures syntax correcness of outstanding change fragments. We only care that changie
# *runs*, not that it produces any specific output. We use 'patch' (increment least significant
# semver chunk) here rather than 'auto' because a PR containing only fragments of the INTERNAL or
# NOTE kind of change fragments (those with `auto: none`) cannot generate a release. Changie won't
# be able to calculate a new version and would fail in a way that's not useful to our syntax chack.
changie batch patch --dry-run >/dev/null

echo "Normal PR changie checks passed."
