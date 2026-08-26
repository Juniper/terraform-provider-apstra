#!/usr/bin/env bash

printenv

set -euo pipefail

BASE_REF="${1:?usage: $0 <base-ref> [normal|release]}"
MODE_INPUT="${2:-}"

echo "Dispatching Changie PR checks against ${BASE_REF}..."

fail() {
    echo "ERROR: $*" >&2
    exit 1
}

# Ensure BASE_REF resolves to a commit object
git rev-parse --verify --quiet "${BASE_REF}^{commit}" >/dev/null || fail "BASE_REF '${BASE_REF}' does not resolve to a valid commit"

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
normal_script="$script_dir/changie_normal_pr_check.sh"
release_script="$script_dir/changie_release_pr_check.sh"

[[ -x "$normal_script" ]] || fail "missing or non-executable script: $normal_script"
[[ -x "$release_script" ]] || fail "missing or non-executable script: $release_script"

pr_has_release_label() {
    command -v jq >/dev/null 2>&1 || fail "jq is required to read labels from GITHUB_EVENT_PATH"
    [[ -n "${GITHUB_EVENT_PATH:-}" ]] || fail "GITHUB_EVENT_PATH is not set; pass mode as second argument (normal|release)"
    [[ -r "$GITHUB_EVENT_PATH" ]] || fail "GITHUB_EVENT_PATH is not readable: $GITHUB_EVENT_PATH"

    # Use lowercase label names to make matching case-insensitive.
    jq -e '.pull_request.labels[]?.name | ascii_downcase == "release"' "$GITHUB_EVENT_PATH" >/dev/null 2>&1
}

resolve_mode() {
    local mode=""

    if [[ -n "$MODE_INPUT" ]]; then
        mode="$MODE_INPUT"
    elif [[ -n "${CHANGIE_PR_KIND:-}" ]]; then
        mode="$CHANGIE_PR_KIND"
    elif pr_has_release_label; then
        mode="release"
    else
        mode="normal"
    fi

    case "$mode" in
        normal|release)
            printf '%s\n' "$mode"
            ;;
        *)
            fail "invalid mode '$mode' (expected normal or release)"
            ;;
    esac
}

mode="$(resolve_mode)"
echo "Detected PR mode: $mode"

case "$mode" in
    normal)
        exec "$normal_script" "$BASE_REF"
        ;;
    release)
        exec "$release_script" "$BASE_REF"
        ;;
esac
