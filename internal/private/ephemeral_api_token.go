package private

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type EphemeralApiToken struct {
	Token         string        `json:"token"`
	ExpiresAt     time.Time     `json:"expires_at"`
	WarnThreshold time.Duration `json:"warn_threshold"`
	DoNotLogOut   bool          `json:"do_not_log_out"`
}

func (o *EphemeralApiToken) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	// This type uses a custom key in private state, but beginning in provider version 0.x.y the data will
	// be stored twice: once under the custom key and once under a key derived from the type name.
	// The custom key should be deprecated in a 2.x.y release or later.
	loadPrivateState(ctx, ps, o, diags, "EphemeralApiToken")
}

func (o EphemeralApiToken) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	// This type uses a custom key in private state, but beginning in provider version 0.x.y the data will
	// be stored twice: once under the custom key and once under a key derived from the type name.
	// The custom key should be deprecated in a 2.x.y release or later.
	setPrivateState(ctx, ps, o, diags, "EphemeralApiToken")
}
