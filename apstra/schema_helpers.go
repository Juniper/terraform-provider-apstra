package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func logicalDeviceAttributesSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Logical Device attributes describe the physical characteristics of a device.",
		Computed:            true,
		PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "Name of the Logical Device object in the global catalog.",
				Computed:            true,
				Type:                types.StringType,
			},
			"panels": {
				MarkdownDescription: "A set of attributes which describe each network connectivity panel on the Logical Device.",
				Computed:            true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"rows": {
						MarkdownDescription: "Physical vertical dimension of the panel (ports).",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"columns": {
						MarkdownDescription: "Physical horizontal dimension of the panel (ports).",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"port_groups": {
						MarkdownDescription: "Ordered logical groupings of interfaces within the panel, by speed or purpose.",
						Computed:            true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"port_count": {
								MarkdownDescription: "Number of ports in the group.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"port_speed_gbps": {
								MarkdownDescription: "Port speed in Gbps.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"port_roles": {
								MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
								Computed:            true,
								Type:                types.SetType{ElemType: types.StringType},
							},
						}),
					},
				}),
			},
		}),
	}
}

func tagDataAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details any tags applied to the Leaf Switch",
		Computed:            true,
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"label": {
				MarkdownDescription: "Tag label (name) field.",
				Computed:            true,
				Type:                types.StringType,
			},
			"description": {
				MarkdownDescription: "Tag description field.",
				Computed:            true,
				Type:                types.StringType,
			},
		}),
	}
}

//func linkAttributeSchema(devType rackElementType) tfsdk.Attribute {
//	lagModes := []string{
//		goapstra.RackLinkLagModeActive.String(),
//		goapstra.RackLinkLagModePassive.String(),
//		goapstra.RackLinkLagModeStatic.String(),
//	}
//	switchPeers := []string{
//		goapstra.RackLinkSwitchPeerFirst.String(),
//		goapstra.RackLinkSwitchPeerSecond.String(),
//	}
//	return tfsdk.Attribute{
//		MarkdownDescription: "Describe links between the device and other systems within the rack",
//		Required:            true,
//		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
//			"name": {
//				MarkdownDescription: "Link name",
//				Required:            true,
//				Type:                types.StringType,
//				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
//			},
//			"target_switch_name": {
//				MarkdownDescription: "Name of the Leaf Switch to which the link connects.",
//				Required:            true,
//				Type:                types.StringType,
//				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
//			},
//			"lag_mode": {
//				MarkdownDescription: fmt.Sprintf("Link LAG mode, must be one of '%s'.", strings.Join(lagModes, "', '")),
//				Optional:            devType != rackElementTypeAccessSwitch, // false for access switches (read-only provider computed)
//				Computed:            devType == rackElementTypeAccessSwitch, // true for access switches (provider always sets true)
//				Type:                types.StringType,
//				Validators:          []tfsdk.AttributeValidator{stringvalidator.OneOfCaseInsensitive(lagModes...)},
//			},
//			"links_per_switch": {
//				MarkdownDescription: "Minimum value '1'.",
//				Required:            true,
//				Type:                types.Int64Type,
//				Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
//			},
//			"speed": {
//				MarkdownDescription: "Link Speed, something like '10G'",
//				Required:            true,
//				Type:                types.StringType,
//			},
//			"tags": {
//				MarkdownDescription: "Labels of tags from the global catalog to be applied to this Link upon Rack Type creation",
//				Optional:            true,
//				Type:                types.SetType{ElemType: types.StringType},
//			},
//			"tag_data": tagDataAttributeSchema(),
//			"switch_peer": {
//				MarkdownDescription: fmt.Sprintf("For non-LAG links to redundant switches, must be one of '%s'.", strings.Join(switchPeers, "', '")),
//				Optional:            true,
//				Computed:            true,
//				Type:                types.StringType,
//				Validators:          []tfsdk.AttributeValidator{stringvalidator.OneOfCaseInsensitive(switchPeers...)},
//			},
//		}),
//	}
//}
