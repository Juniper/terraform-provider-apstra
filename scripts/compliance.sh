#!/usr/bin/env bash
set -euo pipefail

TPC=Third_Party_Code

IGNORE=()
IGNORE+=(--ignore)
IGNORE+=(github.com/Juniper) # don't bother with Juniper licenses
IGNORE+=(--ignore)
IGNORE+=(golang.org/x/sys) # explained below

# golang.org/x/sys is ignored to avoid producing different results on different platforms (x/sys vs. x/sys/unix, etc...)
# The license details for this package, if they were included in the Third_Party_Code directory, would look something
# like this, depending on the build platform:

    ### golang.org/x/sys/unix
    #
    #* Name: golang.org/x/sys/unix
    #* Version: v0.17.0
    #* License: [BSD-3-Clause](https://cs.opensource.google/go/x/sys/+/v0.17.0:LICENSE)
    #
    #
    #Copyright (c) 2009 The Go Authors. All rights reserved.
    #
    #Redistribution and use in source and binary forms, with or without
    #modification, are permitted provided that the following conditions are
    #met:
    #
    #   * Redistributions of source code must retain the above copyright
    #notice, this list of conditions and the following disclaimer.
    #   * Redistributions in binary form must reproduce the above
    #copyright notice, this list of conditions and the following disclaimer
    #in the documentation and/or other materials provided with the
    #distribution.
    #   * Neither the name of Google Inc. nor the names of its
    #contributors may be used to endorse or promote products derived from
    #this software without specific prior written permission.
    #
    #THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
    #"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
    #LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
    #A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
    #OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
    #SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
    #LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
    #DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
    #THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
    #(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
    #OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.


go run github.com/google/go-licenses/v2 save   ${IGNORE[@]} --save_path "${TPC}" --force ./...
go run github.com/google/go-licenses/v2 report ${IGNORE[@]} --template .notices.tpl ./... > "${TPC}/NOTICES.md"

# The `save` command above collects only license and notice files from packages with licenses identified as
# `RestrictionsShareLicense` and collects the entire source tree when the license is identified as
# `RestrictionsShareCode`.
#
# It's true that some licenses require us to "make available" the upstream source code, but I'm not sure
# that doing so as *part of this repository* is appropriate.
# 1. The go package system makes it perfectly clear what we're using and where we got it.
# 2. If somebody wants to really push the issue, we'll find a way to deliver the source independent of this repository.
#
# The line below deletes "saved" files other than those beginning with "LICENSE" and "NOTICE"
find "$TPC" -type f ! -name 'LICENSE*' ! -name 'NOTICE*' -print0 | xargs -0 rm --

# We now likely have some empty directories. Get rid of 'em.
find "$TPC" -depth -type d -empty -exec rmdir -- "{}" \;
