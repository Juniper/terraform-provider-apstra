#!/usr/bin/env bash

set -euo pipefail

CHANGIE_CFG='.changie.yaml'
BASE_BRANCH="main"

fail() {
    echo "ERROR: $*" >&2
    exit 1
}

check_commands() {
    command -v changie >/dev/null || fail "'changie' CLI tool is not installed"
    command -v gh >/dev/null || fail "'gh' CLI tool is not installed"
    command -v git >/dev/null || fail "'git' CLI tool is not installed"
    command -v yq >/dev/null || fail "'yq' CLI tool is not installed"

    gh auth status >/dev/null 2>&1 || {
        echo "ERROR: 'gh' CLI is not authenticated. Run: gh auth login"
        exit 1
    }
}

check_repo() {
    base_branch="$(git branch --show-current)"
    [[ "$base_branch" == "$BASE_BRANCH" ]] || fail "You must be on the '$BASE_BRANCH' branch to prepare a release"

    # Ensure the working tree is clean before making any changes
    [[ -z "$(git status --porcelain)" ]] || fail "Working tree is not clean cannot prepare a release"

    # Ensure we're in sync with origin
    git fetch --quiet origin "$BASE_BRANCH"
    local_head="$(git rev-parse HEAD)"
    remote_head="$(git rev-parse origin/$BASE_BRANCH)"
    [[ "$local_head" == "$remote_head" ]] || fail "Local $BASE_BRANCH is not in sync with origin/$BASE_BRANCH"
}

parse_config() {
    [[ -e "$CHANGIE_CFG" ]] || fail "Changie config $CHANGIE_CFG does not exist"
    [[ -f "$CHANGIE_CFG" ]] || fail "Changie config $CHANGIE_CFG is not a file"
    [[ -r "$CHANGIE_CFG" ]] || fail "Changie config $CHANGIE_CFG is not readable"

    # Read the changie configuration
    CHANGES_DIR="$(yq -r '.changesDir // ""' "$CHANGIE_CFG")"
    UNRELEASED_SUBDIR="$(yq -r '.unreleasedDir // ""' "$CHANGIE_CFG")"
    CHANGELOG="$(yq -r '.changelogPath // ""'  "$CHANGIE_CFG")"
    VERSION_EXT="$(yq -r '.versionExt // ""' "$CHANGIE_CFG")"

    [[ -n "$CHANGES_DIR" ]] || fail "changesDir is not configured in $CHANGIE_CFG"
    [[ -n "$UNRELEASED_SUBDIR" ]] || fail "unreleasedDir is not configured in $CHANGIE_CFG"
    [[ -n "$VERSION_EXT" ]] || fail "versionExt is not configured in $CHANGIE_CFG"
    [[ -n "$CHANGELOG" ]] || fail "changelogPath is not configured in $CHANGIE_CFG"
    [[ -e "$CHANGELOG" ]] || fail "change log file $CHANGELOG does not exist"
    [[ -f "$CHANGELOG" ]] || fail "change log file $CHANGELOG is not a file"
    [[ -r "$CHANGELOG" ]] || fail "change log file $CHANGELOG is not readable"
    [[ -w "$CHANGELOG" ]] || fail "change log file $CHANGELOG is not writable"

    UNRELEASED_DIR="$CHANGES_DIR/$UNRELEASED_SUBDIR"
    [[ -e "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR does not exist"
    [[ -d "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not a directory"
    [[ -r "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not readable"
    [[ -x "$UNRELEASED_DIR" ]] || fail "unreleased changelog fragment directory $UNRELEASED_DIR is not searchable"
}

check_new_semver_possible() {
    # Get all kind keys where auto == none. These kinds of changes do not bump any of the semantic
    # version chunks, and therefore cannot be relied upon to generate a new semver string.
    none_keys="$(yq -r '.kinds[] | select(.auto == "none") | .key' "$CHANGIE_CFG")"

    # Collect the filenames of change fragments we'll pull into the new release notes.
    shopt -s nullglob # we don't want the literal string if globbing can't expand filename(s)
    fragments=("${UNRELEASED_DIR}/"*.yaml)
    shopt -u nullglob # reset filename globbing to normal behavior
    (( ${#fragments[@]} > 0 )) || fail "No unreleased YAML fragments found in '$UNRELEASED_DIR'"

    # At least one of the fragments must be able to bump our semver string or we can't generate a release.
    fragments_bump_semver=0
    for fragment in "${fragments[@]}"; do
        kind="$(yq -r '.kind // ""' "$fragment")"
        [[ -n "$kind" ]] || fail "fragment $fragment: 'kind' field not found"
        if ! grep -qxF "$kind" <<< "$none_keys"; then
            fragments_bump_semver=1
            break
        fi
    done
    if ! (( fragments_bump_semver )); then
        fail "All change fragments are 'auto: none' kinds. Cannot generate a release without user-facing changes."
    fi
}

get_next_version() {
    next_version="$(changie next auto)" || fail "'changie next auto' failed"
    [[ -n "$next_version" ]] || fail "changie generated an empty next version string"
    printf '%s\n' "$next_version"
}

stage_changes() {
  changie batch auto
  changie merge
  git add -A -- "$UNRELEASED_DIR"            # stage removal of change fragment files
  git add "$CHANGES_DIR/${1}.${VERSION_EXT}" # stage addition of per-version changelog file
  git add "$CHANGELOG"                       # stage update to cumulative changelog file
  git diff --quiet || fail "Unstaged changes remain after staging; some changie-modified files may have been missed"
  [[ -z "$(git ls-files --others --exclude-standard)" ]]  || fail "Untracked files found after staging; some changie-created files may have been missed"
}

check_commands
check_repo
parse_config
check_new_semver_possible

next_version=$(get_next_version)
release_branch="release-${next_version}"
git checkout -b "$release_branch"

stage_changes "$next_version"

git commit -m "release $next_version"
git push --set-upstream origin "$release_branch"

gh pr create \
    --base "$BASE_BRANCH" \
    --head "$release_branch" \
    --title "Release $next_version" \
    --body "Automated release PR for release ${next_version} generated by $0." \
    --label release
