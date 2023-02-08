package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type esiLagInfo struct {
	L3PeerLinkCount int64  `tfsdk:"l3_peer_link_count"`
	L3PeerLinkSpeed string `tfsdk:"l3_peer_link_speed"`
}

func (o esiLagInfo) schemaAsDataSource() dataSourceSchema.SingleNestedAttribute {
	return dataSourceSchema.SingleNestedAttribute{
		MarkdownDescription: "Interconnect information for Access Switches in ESI-LAG redundancy mode.",
		Computed:            true,
		Attributes: map[string]dataSourceSchema.Attribute{
			"l3_peer_link_count": dataSourceSchema.Int64Attribute{
				MarkdownDescription: "Count of L3 links to ESI peer.",
				Computed:            true,
			},
			"l3_peer_link_speed": dataSourceSchema.StringAttribute{
				MarkdownDescription: "Speed of L3 links to ESI peer.",
				Computed:            true,
			},
		},
	}
}

func (o esiLagInfo) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"l3_peer_link_count": types.Int64Type,
		"l3_peer_link_speed": types.StringType,
	}
}

func (o esiLagInfo) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *esiLagInfo) loadApiResponse(_ context.Context, in *goapstra.EsiLagInfo, _ *diag.Diagnostics) {
	o.L3PeerLinkCount = int64(in.AccessAccessLinkCount)
	o.L3PeerLinkSpeed = string(in.AccessAccessLinkSpeed)
}

func newEsiLagInfo(ctx context.Context, in *goapstra.EsiLagInfo, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(esiLagInfo{}.attrTypes())
	}

	var eli esiLagInfo
	eli.loadApiResponse(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(esiLagInfo{}.attrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, eli.attrTypes(), &eli)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(esiLagInfo{}.attrTypes())
	}

	return result
}
