package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dRackLink struct {
	Name             types.String `tfsdk:"name"`
	TargetSwitchName types.String `tfsdk:"target_switch_name"`
	LagMode          types.String `tfsdk:"lag_mode"`
	LinksPerSwitch   types.Int64  `tfsdk:"links_per_switch"`
	Speed            types.String `tfsdk:"speed"`
	SwitchPeer       types.String `tfsdk:"switch_peer"`
	TagData          types.Set    `tfsdk:"tag_data"`
}

func (o dRackLink) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of this link.",
			Computed:            true,
		},
		"target_switch_name": schema.StringAttribute{
			MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			Computed:            true,
		},
		"lag_mode": schema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link.",
			Computed:            true,
		},
		"links_per_switch": schema.Int64Attribute{
			MarkdownDescription: "Number of Links to each switch.",
			Computed:            true,
		},
		"speed": schema.StringAttribute{
			MarkdownDescription: "Speed of this Link.",
			Computed:            true,
		},
		"switch_peer": schema.StringAttribute{
			MarkdownDescription: "For non-LAG connections to redundant switch pairs, this field selects the target switch.",
			Computed:            true,
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Link.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.dataSourceAttributes(),
			},
		},
	}
}

func (o dRackLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"switch_peer":        types.StringType,
		"tag_data":           types.SetType{ElemType: types.ObjectType{AttrTypes: tagData{}.attrTypes()}},
	}
}

func (o *dRackLink) loadApiResponse(ctx context.Context, in *goapstra.RackLink, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.TargetSwitchName = types.StringValue(in.TargetSwitchLabel)
	o.LinksPerSwitch = types.Int64Value(int64(in.LinkPerSwitchCount))
	o.Speed = types.StringValue(string(in.LinkSpeed))

	if in.LagMode == goapstra.RackLinkLagModeNone {
		o.LagMode = types.StringNull()
	} else {
		o.LagMode = types.StringValue(in.LagMode.String())
	}

	if in.SwitchPeer == goapstra.RackLinkSwitchPeerNone {
		o.SwitchPeer = types.StringNull()
	} else {
		o.SwitchPeer = types.StringValue(in.SwitchPeer.String())
	}

	o.TagData = newTagDataSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}
}

func newDataSourceLinkSet(ctx context.Context, in []goapstra.RackLink, diags *diag.Diagnostics) types.Set {
	if len(in) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: dRackLink{}.attrTypes()})
	}

	links := make([]dRackLink, len(in))
	for i, link := range in {
		links[i].loadApiResponse(ctx, &link, diags)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: dRackLink{}.attrTypes()})
		}
	}

	result, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: dRackLink{}.attrTypes()}, &links)
	diags.Append(d...)

	return result
}
