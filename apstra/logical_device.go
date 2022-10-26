package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func logicalDeviceAttrType() attr.Type {
	return types.ObjectType{
		AttrTypes: logicalDeviceDataAttrTypes()}
}

func panelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":    types.Int64Type,
		"columns": types.Int64Type,
		"port_groups": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: portGroupAttrTypes(),
			},
		},
	}
}

func portGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port_count":     types.Int64Type,
		"port_speed_bps": types.Int64Type,
		"port_roles": types.SetType{
			ElemType: types.StringType,
		},
	}
}

func logicalDeviceDataAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
		Computed:            true,
		PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"panels": panelsAttributeSchema(),
			"name": {
				MarkdownDescription: "Logical device display name.",
				Computed:            true,
				Type:                types.StringType,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
		}),
	}
}

func panelsAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details physical layout of interfaces on the device.",
		Computed:            true,
		PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"rows": {
				MarkdownDescription: "Physical vertical dimension of the panel.",
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Type:                types.Int64Type,
			},
			"columns": {
				MarkdownDescription: "Physical horizontal dimension of the panel.",
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Type:                types.Int64Type,
			},
			"port_groups": {
				MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"port_count": {
						MarkdownDescription: "Number of ports in the group.",
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Type:                types.Int64Type,
					},
					"port_speed_bps": {
						MarkdownDescription: "Port speed in Gbps.",
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Type:                types.Int64Type,
					},
					"port_roles": {
						MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Type:                types.SetType{ElemType: types.StringType},
					},
				}),
			},
		}),
	}
}
