//go:build tools

package tools

import (
	// document generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"

	// license compliance
	_ "github.com/chrismarget-j/go-licenses"

	// staticcheck
	_ "honnef.co/go/tools/cmd/staticcheck"
)
