#!/usr/bin/env bash

set -euo pipefail

BASE_REF="${1:?usage: $0 <base-ref>}"
echo "Checking normal PR against ${BASE_REF}..."

fail() {
    echo "ERROR: $*" >&2
    exit 1
}

# Ensure BASE_REF resolves to a commit object
git rev-parse --verify --quiet "${BASE_REF}^{commit}" >/dev/null || fail "BASE_REF '${BASE_REF}' does not resolve to a valid commit"

CHANGIE_CFG='.changie.yaml'
[[ -e "$CHANGIE_CFG" ]] || fail "Changie config $CHANGIE_CFG does not exist"
[[ -f "$CHANGIE_CFG" ]] || fail "Changie config $CHANGIE_CFG is not a file"
[[ -r "$CHANGIE_CFG" ]] || fail "Changie config $CHANGIE_CFG is not readable"

# Read the changie configuration
CHANGES_DIR="$(yq -r '.changesDir // ""' "$CHANGIE_CFG")"
UNRELEASED_SUBDIR="$(yq -r '.unreleasedDir // ""' "$CHANGIE_CFG")"
CHANGELOG="$(yq -r '.changelogPath // ""'  "$CHANGIE_CFG")"

[[ -n "$CHANGES_DIR" ]] || fail "changesDir is not configured in $CHANGIE_CFG"
[[ -n "$UNRELEASED_SUBDIR" ]] || fail "unreleasedDir is not configured in $CHANGIE_CFG"
[[ -n "$CHANGELOG" ]] || fail "changelogPath is not configured in $CHANGIE_CFG"
[[ -e "$CHANGELOG" ]] || fail "change log file $CHANGELOG does not exist"
[[ -f "$CHANGELOG" ]] || fail "change log file $CHANGELOG is not a file"
[[ -r "$CHANGELOG" ]] || fail "change log file $CHANGELOG is not readable"

UNRELEASED_DIR="$CHANGES_DIR/$UNRELEASED_SUBDIR"
[[ -e "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR does not exist"
[[ -d "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not a directory"
[[ -r "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not readable"
[[ -x "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not searchable"

new_fragment_path=""
new_fragment_count=0

# Paired reads of status and path should be safe with --no-renames (always two fields per record)
while IFS= read -r -d '' status && IFS= read -r -d '' path; do
    case "$status:$path" in
        # Don't delete $UNRELEASED_DIR/.gitkeep.
        D:"$UNRELEASED_DIR/.gitkeep")
            fail "normal PR deletes $UNRELEASED_DIR/.gitkeep"
            ;;

        # This thing isn't expected to change, but we don't want to accidentally count it as a change fragment.
        *:"$UNRELEASED_DIR/.gitkeep")
            ;;

        # Count the change fragments added via this PR. For our purposes, everything in here is a change fragment.
        A:"$UNRELEASED_DIR/"*)
            new_fragment_path="$path"
            new_fragment_count=$((new_fragment_count + 1)) # Cannot use ((new_fragment_count++)) with set -e
            ;;

        # No other changes are expected in the unreleased fragments directory.
        *:"$UNRELEASED_DIR/"*)
            fail "normal PR modifies an existing unreleased fragment: $path"
            ;;

        # Nothing should be changing in the Changie directory.
        *:"$CHANGES_DIR/"*)
            fail "normal PR modifies a Changie file: $path"
            ;;

        # Changelog shouldn't be changing in a non-release PR.
        *:"$CHANGELOG")
            fail "normal PR modifies $CHANGELOG"
            ;;
    esac
done < <(git diff -z --name-status --no-renames "$BASE_REF"...HEAD)

if [ "$new_fragment_count" -ne 1 ]; then
    fail "normal PR must add exactly one change fragment under $UNRELEASED_DIR/ (found $new_fragment_count)"
fi

echo "Validating $new_fragment_path with changie..."

# Validate all outstanding fragments with Changie. We use 'patch' rather than
# 'auto' because fragments whose kinds have `auto: none` cannot independently
# drive determination of a release version.
changie batch patch --dry-run >/dev/null

echo "Normal PR changie checks passed."
