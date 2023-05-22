all: gofmt govet unit-tests integration-tests device-integration-tests

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

.PHONY: all gofmt govet unit-tests integration-tests device-integration-tests
