package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rRackLink struct {
	Name             types.String `tfsdk:"name"`
	TargetSwitchName types.String `tfsdk:"target_switch_name"`
	LagMode          types.String `tfsdk:"lag_mode"`
	LinksPerSwitch   types.Int64  `tfsdk:"links_per_switch"`
	Speed            types.String `tfsdk:"speed"`
	SwitchPeer       types.String `tfsdk:"switch_peer"`
	TagData          types.Set    `tfsdk:"tag_data"`
	TagIds           types.Set    `tfsdk:"tag_data"`
}

func (o rRackLink) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		//			"name": {
		//				MarkdownDescription: "Name of this link.",
		//				Required:            true,
		//				Type:                types.StringType,
		//				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
		//			},
		//			"target_switch_name": {
		//				MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
		//				Required:            true,
		//				Type:                types.StringType,
		//				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
		//			},
		//			"lag_mode": {
		//				MarkdownDescription: "LAG negotiation mode of the Link.",
		//				Computed:            true,
		//				Optional:            true,
		//				Type:                types.StringType,
		//				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		//				Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(
		//					goapstra.RackLinkLagModeActive.String(),
		//					goapstra.RackLinkLagModePassive.String(),
		//					goapstra.RackLinkLagModeStatic.String())},
		//			},
		//			"links_per_switch": {
		//				MarkdownDescription: "Number of Links to each switch.",
		//				Optional:            true,
		//				Computed:            true,
		//				Type:                types.Int64Type,
		//				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		//				Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(2)},
		//			},
		//			"speed": {
		//				MarkdownDescription: "Speed of this Link.",
		//				Required:            true,
		//				Type:                types.StringType,
		//			},
		//			"switch_peer": {
		//				MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
		//				Optional:            true,
		//				Type:                types.StringType,
		//				Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(
		//					goapstra.RackLinkSwitchPeerFirst.String(),
		//					goapstra.RackLinkSwitchPeerSecond.String(),
		//				)},
		//			},
		//			"tag_ids":  tagIdsAttributeSchema(),
		//			"tag_data": tagsDataAttributeSchema(),
	}
}

func (o rRackLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"switch_peer":        types.StringType,
		"tag_ids":            types.SetType{ElemType: types.StringType},
		"tag_data":           types.SetType{ElemType: tagData{}.attrType()},
	}
}

func (o rRackLink) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *rRackLink) loadApiResponse(ctx context.Context, in *goapstra.RackLink, diags *diag.Diagnostics) {
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

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}
}

func newResourceLinkSet(ctx context.Context, in []goapstra.RackLink, diags *diag.Diagnostics) types.Set {
	if len(in) == 0 {
		return types.SetNull(rRackLink{}.attrType())
	}

	links := make([]rRackLink, len(in))
	for i, link := range in {
		links[i].loadApiResponse(ctx, &link, diags)
		if diags.HasError() {
			return types.SetNull(rRackLink{}.attrType())
		}
	}

	result, d := types.SetValueFrom(ctx, rRackLink{}.attrType(), &links)
	diags.Append(d...)

	return result
}
