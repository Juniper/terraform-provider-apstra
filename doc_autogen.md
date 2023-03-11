### Steps to convert Data Source or Resource documentation from the hand-built README.md to using `tfplugindocs`

1. The details for a given `resource` or `data source` may have been previously documented in [README.md](README.md).
1. Add documentation into the `Schema()` function as `MarkdownDescription` elements within the
`apstra/data_source_<name>.go` or `apstra/resource_<name>.go` files at various levels:
   - at the level of each Attribute (or nested Attribute) element
   - on each individual attribute
1. Migrate the example block(s) of terraform config which may have been documented in [README.md](README.md) into
`examples/<type>/<name>/example.tf`
1. Delete that `data-source` or `resource` documentation from README.md.
1. Execute `go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs` from the repo root to rebuild the docs.
1. Alternate: install tfplugindocs: `go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest`
