#!/usr/bin/env bash
set -euo pipefail

TPC=Third_Party_Code

# Default values
minimal_mode=false

# Loop over all arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --minimal)
            minimal_mode=true
            ;;
        *)
            echo "Unknown argument: $1" >&2
            exit 1
            ;;
    esac
    shift
done

IGNORE=()
IGNORE+=(--ignore)
IGNORE+=(github.com/Juniper) # don't bother with Juniper licenses

go tool go-licenses save   ${IGNORE[@]} --save_path "${TPC}" --force ./...
go tool go-licenses report ${IGNORE[@]} --template .notices.tpl ./... > "${TPC}/NOTICES.md"

# The `save` command above collects only license and notice files from packages with licenses identified as
# `RestrictionsShareLicense` and collects the entire source tree when the license is identified as
# `RestrictionsShareCode`.
#
# It's true that some licenses require us to "make available" the upstream source code, but I'm not sure
# that doing so as *part of this repository* is appropriate.
# 1. The go package system makes it perfectly clear what we're using and where we got it.
# 2. We can deliver those libraries as part of our release .zip files
if [[ "$minimal_mode" == true ]]; then
    echo "Removing third party source files."
    # The line below deletes "saved" files other than those beginning with "LICENSE" and "NOTICE"
    find "$TPC" -type f ! -name 'LICENSE*' ! -name 'NOTICE*' -print0 | xargs -0 rm --

    # We now likely have some empty directories. Get rid of 'em.
    find "$TPC" -depth -type d -empty -exec rmdir -- "{}" \;
fi
