all: install gofmt gofumpt compliance docs govet unit-tests staticcheck check-repo-clean

check-repo-clean:
	git update-index --refresh && git diff-index --quiet HEAD --

compliance:
	bash scripts/compliance.sh --minimal

compliance-with-source:
	bash scripts/compliance.sh

compliance-check: compliance check-repo-clean

device-integration-tests:
	go test -tags device-integration -v ./...

docs:
	@sh -c "$(CURDIR)/scripts/tfplugindocs.sh"

docs-check:
	@sh -c "$(CURDIR)/scripts/tfplugindocs.sh --check"

gofmt:
	@sh -c "$(CURDIR)/scripts/gofmtcheck.sh"

gofumpt:
	@sh -c "$(CURDIR)/scripts/gofumptcheck.sh"

govet:
	go vet -v ./...

install:
	go install

integration-tests:
	go test -tags integration -v ./...

release:
	printenv GITHUB_TOKEN > /dev/null || (echo "GITHUB_TOKEN not found in environment"; false)
	(cd tools/goreleaser && GPG_FINGERPRINT=4EACB71B2FC20EC8499576BDCB9C922903A66F3F go tool goreleaser release --clean)

release-dry-run:
	(cd tools/goreleaser && GPG_FINGERPRINT=4EACB71B2FC20EC8499576BDCB9C922903A66F3F go tool goreleaser release --clean --skip-publish)

release-pr:
	@sh -c "$(CURDIR)/scripts/make_release_pr.sh"

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck ./...

unit-tests:
	go test -v ./...

.PHONY: \
  all \
  check-repo-clean \
  compliance \
  compliance-with-source \
  compliance-check \
  device-integration-tests \
  docs \
  docs-check \
  gofmt \
  gofumpt \
  govet \
  install \
  integration-tests \
  release \
  release-dry-run \
  release-pr \
  staticcheck \
  unit-tests
