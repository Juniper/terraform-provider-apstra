### Steps to convert a Data Source documentation from the hand-built README.md to using `tfplugindocs`

1. find the details for a given resource in README.md or https://github.com/chrismarget-j/terraform-provider-apstra/blob/main/README.md
1. migrate information from the README.md into the `GetSchema()` function as `MarkdownDescription` elements within the
`apstra/data_source_<name>.go` files various levels:
   - at the level of each Attributes (or nested Attributes) element
   - on each individual attribute
1. migrate the example block(s) of terraform config from the README.md into `examples/data-sources/<name>/<filename>`
1. create `templates/data-sources/<name>.md.tmpl` and cite the files above in here.
1. delete that data-source's documentation from README.md
1. execute `go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs` from the repo root to build the docs
