package private

import (
	"context"

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
	loadPrivateState(ctx, ps, o, diags)
}

func (o ResourceDatacenterRoutingZoneLoopbackAddresses) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	setPrivateState(ctx, ps, o, diags)
}
