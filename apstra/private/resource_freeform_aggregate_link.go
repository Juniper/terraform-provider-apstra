package private

import (
	"context"

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
	loadPrivateState(ctx, ps, o, diags)
}

func (o ResourceFreeformAggregateLink) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	setPrivateState(ctx, ps, o, diags)
}
