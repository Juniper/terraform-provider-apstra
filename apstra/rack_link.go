package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	TagData          types.List   `tfsdk:"tag_data"`
}

func (o rackLink) schema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
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
				MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
				Computed:            true,
			},
			"tag_data": tagsDataAttributeSchema(),
		},
	}
}

//func (o dRackLink) attrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: map[string]attr.Type{
//			"name":               types.StringType,
//			"target_switch_name": types.StringType,
//			"lag_mode":           types.StringType,
//			"links_per_switch":   types.Int64Type,
//			"speed":              types.StringType,
//			"switch_peer":        types.StringType,
//			"tag_data":           types.SetType{ElemType: tagData{}.attrType()}}}
//}
//
//func (o *dRackLink) loadApiResponse(in *goapstra.RackLink) {
//	o.Name = in.Label
//	o.TargetSwitchName = in.TargetSwitchLabel
//	if in.LagMode != goapstra.RackLinkLagModeNone {
//		lagMode := in.LagMode.String()
//		o.LagMode = &lagMode
//	}
//	o.LinksPerSwitch = int64(in.LinkPerSwitchCount)
//	o.Speed = string(in.LinkSpeed)
//	if in.SwitchPeer != goapstra.RackLinkSwitchPeerNone {
//		switchPeer := in.SwitchPeer.String()
//		o.SwitchPeer = &switchPeer
//	}
//
//	if len(in.Tags) > 0 {
//		o.TagData = make([]tagData, len(in.Tags)) // populated below
//		for i := range in.Tags {
//			o.TagData[i].loadApiResponse(&in.Tags[i])
//		}
//	}
//}
//

func dLinksAttributeSchema() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		MarkdownDescription: "Details links from this Element to switches upstream switches within this Rack Type.",
		Computed:            true,
		Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
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
					MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
					Computed:            true,
				},
				"tag_data": tagsDataAttributeSchema(),
			},
		},
	}
}
