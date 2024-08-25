all: compliance-check docs-check gofmt govet unit-tests integration-tests device-integration-tests

check-repo-clean:
	git update-index --refresh && git diff-index --quiet HEAD --

compliance:
    # golang.org/x/sys ignored to avoid producing different results on different platforms (x/sys vs. x/sys/unix, etc...)
    # The license details, if it were included in the Third_Party_Code directory, would look something like this, depending
    # on the build platform:

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

	go run github.com/chrismarget-j/go-licenses save   --ignore github.com/Juniper --ignore golang.org/x/sys --save_path Third_Party_Code --force ./... || exit 1 ;\
	go run github.com/chrismarget-j/go-licenses report --ignore github.com/Juniper --ignore golang.org/x/sys --template .notices.tpl ./... > Third_Party_Code/NOTICES.md || exit 1 ;\

compliance-check: compliance check-repo-clean

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

docs-check:
	@sh -c "$(CURDIR)/scripts/tfplugindocs.sh"

gofmt:
	@sh -c "$(CURDIR)/scripts/gofmtcheck.sh"

govet:
	go vet -v ./...

unit-tests:
	go test -v ./...

integration-tests:
	go test -tags integration -v ./...

device-integration-tests:
	go test -tags device-integration -v ./...

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck ./...

release:
	printenv GITHUB_TOKEN > /dev/null || (echo "GITHUB_TOKEN not found in environment"; false)
	GPG_FINGERPRINT=4EACB71B2FC20EC8499576BDCB9C922903A66F3F go run github.com/goreleaser/goreleaser@v1.26.2 release --clean

gofumpt:
	@sh -c "$(CURDIR)/scripts/gofumptcheck.sh"

.PHONY: all compliance compliance-check docs docs-check gofmt govet unit-tests integration-tests device-integration-tests staticcheck
