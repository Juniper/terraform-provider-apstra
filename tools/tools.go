//go:build tools

package tools

import (
	// staticcheck
	_ "honnef.co/go/tools/cmd/staticcheck"

	// gofumpt does strict formatting
	_ "mvdan.cc/gofumpt"
)
