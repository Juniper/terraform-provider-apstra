package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
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
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
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
			"tag_ids":  tagIdsAttributeSchema(),
			"tag_data": tagsDataAttributeSchema(),
		}),
	}
}
