package private

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ResourceDatacenterSwitchingZone is stored in private state by
// resourceDatacenterSwitchingZone methods Create() and Update().
// It contains a boolean which indicates whether the user specified a value for
// the route_target attribute last time the resource was modified.
// The boolean is used when planning updates: If the config for route_target is
// null, but this record indicates one was previously set, we'll set the value
// to Unknown to clear it from the API and allow Apstra to assign a new value.
type ResourceDatacenterSwitchingZone struct {
	HasRouteTarget bool `json:"has_route_target"`
}

func (o *ResourceDatacenterSwitchingZone) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	loadPrivateState(ctx, ps, o, diags)
}

func (o ResourceDatacenterSwitchingZone) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	setPrivateState(ctx, ps, o, diags)
}
