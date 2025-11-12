all: compliance-check docs-check gofmt govet unit-tests integration-tests device-integration-tests

check-repo-clean:
	git update-index --refresh && git diff-index --quiet HEAD --

compliance:
	bash scripts/compliance.sh --minimal

compliance-with-source:
	bash scripts/compliance.sh

compliance-check: compliance check-repo-clean

docs:
	@sh -c "$(CURDIR)/scripts/tfplugindocs.sh"

docs-check:
	@sh -c "$(CURDIR)/scripts/tfplugindocs.sh --check"

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
	(cd tools/goreleaser && GPG_FINGERPRINT=4EACB71B2FC20EC8499576BDCB9C922903A66F3F go tool goreleaser release --clean)

release-dry-run:
	(cd tools/goreleaser && GPG_FINGERPRINT=4EACB71B2FC20EC8499576BDCB9C922903A66F3F go tool goreleaser release --clean --skip-publish)

gofumpt:
	@sh -c "$(CURDIR)/scripts/gofumptcheck.sh"

.PHONY: all compliance compliance-check docs docs-check gofmt govet unit-tests integration-tests device-integration-tests staticcheck
