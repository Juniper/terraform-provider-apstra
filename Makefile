all: compliance-check docs-check gofmt govet unit-tests integration-tests device-integration-tests

check-repo-clean:
	git update-index --refresh && git diff-index --quiet HEAD --

compliance:
	go run github.com/chrismarget-j/go-licenses save --ignore terraform-provider-apstra --ignore github.com/Juniper/apstra-go-sdk --save_path=./Third_Party_Code --force ./... || exit 1 ;\
	go run github.com/chrismarget-j/go-licenses report ./... --ignore terraform-provider-apstra --ignore github.com/Juniper/apstra-go-sdk/apstra --template .notices.tpl > Third_Party_Code/NOTICES.md || exit 1 ;\

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

.PHONY: all compliance compliance-check docs docs-check gofmt govet unit-tests integration-tests device-integration-tests staticcheck
