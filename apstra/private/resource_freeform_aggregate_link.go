package private

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ResourceFreeformAggregateLink is stored in private state by the
// resourceFreeformAggregateLink Read() function.
// It contains a map[string]string which links system IDs (key) to
// aggregate interface ID. These are the IDs found in FreeformAggregateLinkEndpoint
// types.
// We collect it during Read() so that Update() can include the interface
// ID when it invokes the PATCH API.
// Update() does not use copies of this data from state because the endpoints
// are a list: Addition or removal of an endpoint from the list could cause
// the per-system IDs to be off by one.
type ResourceFreeformAggregateLink struct {
	SysIDToLAGID map[string]string
}

func (o *ResourceFreeformAggregateLink) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, d := ps.GetKey(ctx, fmt.Sprintf("%T", *o))
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	if len(b) == 0 {
		return
	}

	err := json.Unmarshal(b, &o)
	if err != nil {
		diags.AddError("failed to unmarshal private state", err.Error())
		return
	}
}

func (o ResourceFreeformAggregateLink) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, fmt.Sprintf("%T", o), b)...)
}
