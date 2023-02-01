package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type mlagInfo struct {
	MlagKeepaliveVLan       types.Int64  `tfsdk:"mlag_keepalive_vlan"`
	PeerLinkCount           types.Int64  `tfsdk:"peer_link_count"`
	PeerLinkSpeed           types.String `tfsdk:"peer_link_speed"`
	PeerLinkPortChannelId   types.Int64  `tfsdk:"peer_link_port_channel_id"`
	L3PeerLinkCount         types.Int64  `tfsdk:"l3_peer_link_count"`
	L3PeerLinkSpeed         types.String `tfsdk:"l3_peer_link_speed"`
	L3PeerLinkPortChannelId types.Int64  `tfsdk:"l3_peer_link_port_channel_id"`
}

func (o mlagInfo) schema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Details settings when the Leaf Switch is an MLAG-capable pair.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"mlag_keepalive_vlan": schema.Int64Attribute{
				MarkdownDescription: "MLAG keepalive VLAN ID.",
				Computed:            true,
			},
			"peer_link_count": schema.Int64Attribute{
				MarkdownDescription: "Number of links between MLAG devices.",
				Computed:            true,
			},
			"peer_link_speed": schema.StringAttribute{
				MarkdownDescription: "Speed of links between MLAG devices.",
				Computed:            true,
			},
			"peer_link_port_channel_id": schema.Int64Attribute{
				MarkdownDescription: "Peer link port-channel ID.",
				Computed:            true,
			},
			"l3_peer_link_count": schema.Int64Attribute{
				MarkdownDescription: "Number of L3 links between MLAG devices.",
				Computed:            true,
			},
			"l3_peer_link_speed": schema.StringAttribute{
				MarkdownDescription: "Speed of l3 links between MLAG devices.",
				Computed:            true,
			},
			"l3_peer_link_port_channel_id": schema.Int64Attribute{
				MarkdownDescription: "L3 peer link port-channel ID.",
				Computed:            true,
			},
		},
	}
}

func (o mlagInfo) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mlag_keepalive_vlan":          types.Int64Type,
		"peer_link_count":              types.Int64Type,
		"peer_link_speed":              types.StringType,
		"peer_link_port_channel_id":    types.Int64Type,
		"l3_peer_link_count":           types.Int64Type,
		"l3_peer_link_speed":           types.StringType,
		"l3_peer_link_port_channel_id": types.Int64Type}
}

func (o mlagInfo) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *mlagInfo) loadApiResponse(_ context.Context, in *goapstra.LeafMlagInfo, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load mlagInfo from nil pointer")
	}

	var l3PeerLinkPortChannelId, l3PeerLinkCount types.Int64
	var l3PeerLinkSpeed types.String
	if in.LeafLeafL3LinkCount > 0 {
		l3PeerLinkPortChannelId = types.Int64Value(int64(in.LeafLeafL3LinkPortChannelId))
		l3PeerLinkCount = types.Int64Value(int64(in.LeafLeafL3LinkCount))
		l3PeerLinkSpeed = types.StringValue(string(in.LeafLeafL3LinkSpeed))
	} else {
		l3PeerLinkPortChannelId = types.Int64Null()
		l3PeerLinkCount = types.Int64Null()
		l3PeerLinkSpeed = types.StringNull()
	}

	o.MlagKeepaliveVLan = types.Int64Value(int64(in.MlagVlanId))
	o.PeerLinkCount = types.Int64Value(int64(in.LeafLeafLinkCount))
	o.PeerLinkSpeed = types.StringValue(string(in.LeafLeafLinkSpeed))
	o.PeerLinkPortChannelId = types.Int64Value(int64(in.LeafLeafLinkPortChannelId))
	o.L3PeerLinkCount = l3PeerLinkCount
	o.L3PeerLinkSpeed = l3PeerLinkSpeed
	o.L3PeerLinkPortChannelId = l3PeerLinkPortChannelId
}

//func (o *mlagInfo) request() *goapstra.LeafMlagInfo {
//	if o == nil {
//		return nil
//	}
//
//	var leafLeafL3LinkCount int
//	if o.L3PeerLinkCount != nil {
//		leafLeafL3LinkCount = int(*o.L3PeerLinkCount)
//	}
//
//	var leafLeafL3LinkPortChannelId int
//	if o.L3PeerLinkPortChannelId != nil {
//		leafLeafL3LinkPortChannelId = int(*o.L3PeerLinkPortChannelId)
//	}
//
//	var leafLeafLinkPortChannelId int
//	if o.PeerLinkPortChannelId != nil {
//		leafLeafLinkPortChannelId = int(*o.PeerLinkPortChannelId)
//	}
//
//	var leafLeafL3LinkSpeed goapstra.LogicalDevicePortSpeed
//	if o.L3PeerLinkSpeed != nil {
//		leafLeafL3LinkSpeed = goapstra.LogicalDevicePortSpeed(*o.L3PeerLinkSpeed)
//	}
//
//	return &goapstra.LeafMlagInfo{
//		LeafLeafL3LinkCount:         leafLeafL3LinkCount,
//		LeafLeafL3LinkPortChannelId: leafLeafL3LinkPortChannelId,
//		LeafLeafL3LinkSpeed:         leafLeafL3LinkSpeed,
//		LeafLeafLinkCount:           int(o.PeerLinkCount),
//		LeafLeafLinkPortChannelId:   leafLeafLinkPortChannelId,
//		LeafLeafLinkSpeed:           goapstra.LogicalDevicePortSpeed(o.PeerLinkSpeed),
//		MlagVlanId:                  int(o.MlagKeepaliveVLan),
//	}
//}

func newMlagInfoObject(ctx context.Context, in *goapstra.LeafMlagInfo, diags *diag.Diagnostics) types.Object {
	if in == nil || in.LeafLeafLinkCount > 0 {
		return types.ObjectNull(mlagInfo{}.attrTypes())
	}

	var mi mlagInfo
	mi.loadApiResponse(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(mlagInfo{}.attrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, mi.attrTypes(), &mi)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(mlagInfo{}.attrTypes())
	}

	return result
}
