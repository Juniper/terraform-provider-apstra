package private

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ResourceDatacenterRoutingZoneLoopbackAddresses is stored in private state by
// resourceDatacenterRoutingZoneLoopbackAddresses methods Create() and Update().
// It contains a record of switch node IDs which previously had IPv4 and IPv6
// loopback addresses configured.
// This record is consulted in the following methods:
// - Read() - we don't read all loopbacks, only ones previously configured.
// - Update() - previous assignments which do not appear in the plan are cleared.
// - Delete() - all previous assignments are cleared.
type ResourceDatacenterRoutingZoneLoopbackAddresses map[string]struct {
	HasIpv4 bool `json:"has_ipv4"`
	HasIpv6 bool `json:"has_ipv6"`
}

func (o *ResourceDatacenterRoutingZoneLoopbackAddresses) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, d := ps.GetKey(ctx, fmt.Sprintf("%T", *o))
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

func (o *ResourceDatacenterRoutingZoneLoopbackAddresses) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, fmt.Sprintf("%T", *o), b)...)
}
