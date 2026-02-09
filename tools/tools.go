//go:build tools

package tools

import (
	// license compliance
	_ "github.com/google/go-licenses/v2"

	// staticcheck
	_ "honnef.co/go/tools/cmd/staticcheck"

	// gofumpt does strict formatting
	_ "mvdan.cc/gofumpt"
)
