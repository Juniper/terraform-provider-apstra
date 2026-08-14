#!/usr/bin/env bash

set -euo pipefail

BASE_REF="${1:?usage: $0 <base-ref>}"
echo "Checking release PR against ${BASE_REF}..."

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
VERSION_EXT="$(yq -r '.versionExt // ""' "$CHANGIE_CFG")"
CHANGELOG="$(yq -r '.changelogPath // ""'  "$CHANGIE_CFG")"

[[ -n "$CHANGES_DIR" ]] || fail "changesDir is not configured in $CHANGIE_CFG"
[[ -n "$UNRELEASED_SUBDIR" ]] || fail "unreleasedDir is not configured in $CHANGIE_CFG"
[[ -n "$VERSION_EXT" ]] || fail "versionExt is not configured in $CHANGIE_CFG"
[[ -n "$CHANGELOG" ]] || fail "changelogPath is not configured in $CHANGIE_CFG"
[[ -e "$CHANGELOG" ]] || fail "change log file $CHANGELOG does not exist"
[[ -f "$CHANGELOG" ]] || fail "change log file $CHANGELOG is not a file"
[[ -r "$CHANGELOG" ]] || fail "change log file $CHANGELOG is not readable"

UNRELEASED_DIR="$CHANGES_DIR/$UNRELEASED_SUBDIR"
[[ -e "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR does not exist"
[[ -d "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not a directory"
[[ -r "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not readable"
[[ -x "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not searchable"

# Calculate the version that Changie expects to release from the base revision.
tmp_worktree="$(mktemp -d)"
cleanup() {
    git worktree remove --force "$tmp_worktree" >/dev/null 2>&1 || true
}
trap cleanup EXIT

git worktree add --detach "$tmp_worktree" "$BASE_REF" >/dev/null

expected_version="$(
    cd "$tmp_worktree"
    changie next auto
)"

echo "Expected release version: $expected_version"

# The release must consume all outstanding fragments from unreleasedDir.
while IFS= read -r -d '' path; do
    # .gitkeep is allowed to remain so that the directory persists in Git.
    [ "$path" = "$UNRELEASED_DIR/.gitkeep" ] && continue

    fail "release PR leaves an unreleased change fragment: $path"
done < <(find "$UNRELEASED_DIR" -type f -print0)

expected_version_path="$CHANGES_DIR/$expected_version.$VERSION_EXT"

# Inspect the PR diff.
expected_version_found=0
changelog_modified=0

# Paired reads of status and path should be safe with --no-renames (always two fields per record)
while IFS= read -r -d '' status && IFS= read -r -d '' path; do
    case "$status:$path" in
        # The expected version release file should be [A]dded. Note it.
        A:"$expected_version_path")
            expected_version_found=1
            ;;

        # Don't delete $UNRELEASED_DIR/.gitkeep.
        D:"$UNRELEASED_DIR/.gitkeep")
            fail "release PR deletes $UNRELEASED_DIR/.gitkeep"
            ;;

        # Changie consumes these fragments during the release.
        D:"$UNRELEASED_DIR/"*)
            ;;

        # No other changie files should be touched.
        *:"$CHANGES_DIR/"*)
            fail "release PR modifies an unexpected Changie file: $path"
            ;;

        # The changelog must be modified. Note it.
        M:"$CHANGELOG")
            changelog_modified=1
            ;;

        # No other changes permitted.
        *)
            fail "release PR modifies an unrelated file: $path"
            ;;
    esac
done < <(git diff -z --name-status --no-renames "$BASE_REF"...HEAD)


if ! (( expected_version_found )); then
    fail "release PR must add $expected_version_path"
fi

if ! (( changelog_modified )); then
    fail "release PR does not modify $CHANGELOG"
fi

echo "Release PR changie checks passed."
