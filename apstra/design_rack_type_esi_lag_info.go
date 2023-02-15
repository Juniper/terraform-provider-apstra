package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type esiLagInfo struct {
	L3PeerLinkCount types.Int64  `tfsdk:"l3_peer_link_count"`
	L3PeerLinkSpeed types.String `tfsdk:"l3_peer_link_speed"`
}

func (o esiLagInfo) schemaAsDataSource() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"l3_peer_link_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of L3 links between ESI peers.",
			Computed:            true,
		},
		"l3_peer_link_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of L3 links between ESI peers.",
			Computed:            true,
		},
	}
}

func (o esiLagInfo) schemaAsResource() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"l3_peer_link_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Count of L3 links between ESI peers.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"l3_peer_link_speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of L3 links between ESI peers.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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

func (o *esiLagInfo) loadApiResponse(_ context.Context, in *goapstra.EsiLagInfo, diags *diag.Diagnostics) {
	o.L3PeerLinkCount = types.Int64Value(int64(in.AccessAccessLinkCount))
	o.L3PeerLinkSpeed = types.StringValue(string(in.AccessAccessLinkSpeed))
}

func (o *esiLagInfo) request(_ context.Context, diags *diag.Diagnostics) *goapstra.EsiLagInfo {
	if o.L3PeerLinkSpeed.IsNull() && o.L3PeerLinkSpeed.IsNull() {
		return nil
	}

	if !o.L3PeerLinkSpeed.IsNull() && !o.L3PeerLinkCount.IsNull() {
		return &goapstra.EsiLagInfo{
			AccessAccessLinkCount: int(o.L3PeerLinkCount.ValueInt64()),
			AccessAccessLinkSpeed: goapstra.LogicalDevicePortSpeed(o.L3PeerLinkSpeed.ValueString()),
		}
	}

	diags.AddError(errProviderBug, "attempt to generate an EsiLagInfo request with some, but not all null fields")
	return nil
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
