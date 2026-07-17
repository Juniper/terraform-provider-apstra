package private

import (
	"context"
	"encoding/json"
	"fmt"

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

func (o ResourceDatacenterSwitchingZone) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, fmt.Sprintf("%T", o), b)...)
}
