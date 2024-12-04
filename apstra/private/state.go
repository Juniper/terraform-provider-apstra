package private

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// State is intended as a stand-in for ProviderData from the not-import-able
// github.com/hashicorp/terraform-plugin-framework/internal/privatestate package.
type State interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}
