package private

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type EphemeralApiToken struct {
	Token         string        `json:"token"`
	ExpiresAt     time.Time     `json:"expires_at"`
	WarnThreshold time.Duration `json:"warn_threshold"`
}

func (o *EphemeralApiToken) LoadApiData(_ context.Context, token string, expiresAt time.Time, warnThreshold time.Duration, diags *diag.Diagnostics) {
	o.Token = token
	o.ExpiresAt = expiresAt
	o.WarnThreshold = warnThreshold
}

func (o *EphemeralApiToken) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, d := ps.GetKey(ctx, "EphemeralApiToken")
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	err := json.Unmarshal(b, &o)
	if err != nil {
		diags.AddError("failed to unmarshal private state", err.Error())
		return
	}
}

func (o *EphemeralApiToken) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, "EphemeralApiToken", b)...)
}
