package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rackLink struct {
	Name             types.String `tfsdk:"name"`
	TargetSwitchName types.String `tfsdk:"target_switch_name"`
	LagMode          types.String `tfsdk:"lag_mode"`
	LinksPerSwitch   types.Int64  `tfsdk:"links_per_switch"`
	Speed            types.String `tfsdk:"speed"`
	SwitchPeer       types.String `tfsdk:"switch_peer"`
	TagIds           types.Set    `tfsdk:"tag_ids"`
	Tags             types.Set    `tfsdk:"tags"`
}

func (o rackLink) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of this link.",
			Computed:            true,
		},
		"target_switch_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			Computed:            true,
		},
		"lag_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link.",
			Computed:            true,
		},
		"links_per_switch": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Links to each switch.",
			Computed:            true,
		},
		"speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of this Link.",
			Computed:            true,
		},
		"switch_peer": dataSourceSchema.StringAttribute{
			MarkdownDescription: "For non-LAG connections to redundant switch pairs, this field selects the target switch.",
			Computed:            true,
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in data source contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Link.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o rackLink) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"target_switch_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"lag_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link.",
			Computed:            true,
			Optional:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.RackLinkLagModeActive.String(),
				goapstra.RackLinkLagModePassive.String(),
				goapstra.RackLinkLagModeStatic.String(),
			)},
		},
		"links_per_switch": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Links to each switch.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of this Link.",
			Required:            true,
		},
		"switch_peer": resourceSchema.StringAttribute{
			MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.RackLinkSwitchPeerFirst.String(),
				goapstra.RackLinkSwitchPeerSecond.String(),
			)},
		},
		"tag_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Link",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Link",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: tag{}.resourceAttributes(),
			},
		},
	}
}

func (o rackLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"switch_peer":        types.StringType,
		"tag_ids":            types.SetType{ElemType: types.StringType},
		"tags":               types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *rackLink) loadApiData(ctx context.Context, in *goapstra.RackLink, diags *diag.Diagnostics) {
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

	o.TagIds = types.SetNull(types.StringType)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

func newLinkSet(ctx context.Context, in []goapstra.RackLink, diags *diag.Diagnostics) types.Set {
	links := make([]rackLink, len(in))
	for i, link := range in {
		links[i].loadApiData(ctx, &link, diags)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: rackLink{}.attrTypes()})
		}
	}

	return setValueOrNull(ctx, types.ObjectType{AttrTypes: rackLink{}.attrTypes()}, links, diags)
}
