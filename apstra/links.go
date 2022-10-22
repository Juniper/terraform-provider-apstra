package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func linksElemType() attr.Type {
	return types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: linksAttrTypes()}}
}

func linkAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"attachment_type":    types.StringType,
		"switch_peer":        types.StringType,
		"tag_names":          tagNameElemType(),
		"tag_data":           tagDataElemType(),
	}
}

func rRackLinkAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Link details for any connection from a Rack Type element " +
			"(Access Switch or Generic System) to the upstream switch providing connectivity to that element.",
		Required:   true,
		Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "Name of this link.",
				Required:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"target_switch_name": {
				MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
				Required:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"lag_mode": {
				MarkdownDescription: "LAG negotiation mode of the Link.",
				Computed:            true,
				Optional:            true,
				Type:                types.StringType,
				Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(
					goapstra.RackLinkLagModeActive.String(),
					goapstra.RackLinkLagModePassive.String(),
					goapstra.RackLinkLagModeStatic.String())},
			},
			"links_per_switch": {
				MarkdownDescription: "Number of Links to each switch.",
				Computed:            true,
				Optional:            true,
				Type:                types.Int64Type,
				Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
			},
			"speed": {
				MarkdownDescription: "Speed of this Link.",
				Required:            true,
				Type:                types.StringType,
			},
			"switch_peer": {
				MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
				Optional:            true,
				Type:                types.StringType,
			},
			"tag_names": tagLabelsAttributeSchema(),
			"tag_data":  tagsDataAttributeSchema(),
		}),
	}
}
