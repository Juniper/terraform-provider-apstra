package private

import (
	"context"

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
	VLANNull        bool `json:"vlan_null"`
}

func (o *ResourceDatacenterVirtualNetworkAssignment) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	loadPrivateState(ctx, ps, o, diags)
}

func (o ResourceDatacenterVirtualNetworkAssignment) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	setPrivateState(ctx, ps, o, diags)
}
