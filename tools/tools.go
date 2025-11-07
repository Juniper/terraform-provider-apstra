//go:build tools

package tools

import (
	// document generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"

	// license compliance
	_ "github.com/google/go-licenses/v2"

	// staticcheck
	_ "honnef.co/go/tools/cmd/staticcheck"

	// release
	_ "github.com/goreleaser/goreleaser/v2"

	// gofumpt does strict formatting
	_ "mvdan.cc/gofumpt"
)
