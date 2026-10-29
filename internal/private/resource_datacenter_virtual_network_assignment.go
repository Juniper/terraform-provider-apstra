package private

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ResourceDatacenterVirtualNetworkAssignment is stored in private state by Create() and Update() methods of the
// ResourceDatacenterVirtualNetworkAssignment resource. It is used to track whether which Optional but non-Computed
// attributes were set by the user. During Read() these values will be set to null if the user left them null in
// the configuration. This is done for consistency during various lifecycle stages of the resource. During Create(),
// we *might* be able to determine whether Apstra has assigned a value for these attributes, but it's a waiting game.
// The resource pool might not be assigned yet. So either we wait an unreasonable amount of time for an update which
// may or may not happen, or we set the values to null. We choose the latter. During Read(), we use this private
// state to restore those values (matching the user's configuration) so that the behavior isn't surprising (sometimes
// null and sometimes with a value assigned).
type ResourceDatacenterVirtualNetworkAssignment struct {
	IPv4ModeNull    bool `json:"ipv4_mode_null"`
	IPv4AddressNull bool `json:"ipv4_address_null"`
	IPv6ModeNull    bool `json:"ipv6_mode_null"`
	IPv6AddressNull bool `json:"ipv6_address_null"`
}

func (o *ResourceDatacenterVirtualNetworkAssignment) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
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

func (o ResourceDatacenterVirtualNetworkAssignment) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, fmt.Sprintf("%T", o), b)...)
}
