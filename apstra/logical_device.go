package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func logicalDeviceAttrType() attr.Type {
	return types.ObjectType{
		AttrTypes: logicalDeviceDataAttrTypes()}
}

type panelPortGroup struct {
	PortCount    int64    `tfsdk:"port_count"`
	PortSpeedBps int64    `tfsdk:"port_speed_bps"`
	PortRoles    []string `tfsdk:"port_roles"`
}

func parseApiPanelPortGroup(in goapstra.LogicalDevicePortGroup) *panelPortGroup {
	return &panelPortGroup{
		PortCount:    int64(in.Count),
		PortSpeedBps: in.Speed.BitsPerSecond(),
		PortRoles:    in.Roles.Strings(),
	}
}
func panelPortGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port_count":     types.Int64Type,
		"port_speed_bps": types.Int64Type,
		"port_roles":     types.SetType{ElemType: types.StringType},
	}
}

type panel struct {
	Rows       int64            `tfsdk:"rows"`
	Columns    int64            `tfsdk:"columns"`
	PortGroups []panelPortGroup `tfsdk:"port_groups"`
}

func parseApiPanel(in *goapstra.LogicalDevicePanel) *panel {
	portGroups := make([]panelPortGroup, len(in.PortGroups))
	for i := range in.PortGroups {
		portGroups[i] = *parseApiPanelPortGroup(in.PortGroups[i])
	}
	return &panel{
		Rows:       int64(in.PanelLayout.RowCount),
		Columns:    int64(in.PanelLayout.ColumnCount),
		PortGroups: portGroups,
	}
}
func panelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":        types.Int64Type,
		"columns":     types.Int64Type,
		"port_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: panelPortGroupAttrTypes()}},
	}
}

type logicalDeviceData struct {
	Panels []panel `tfsdk:"panels"'`
	Name   string  `tfsdk:"name"`
}

func parseApiLogicalDeviceData(in *goapstra.LogicalDeviceData) *logicalDeviceData {
	panels := make([]panel, len(in.Panels))
	for i := range in.Panels {
		panels[i] = *parseApiPanel(&in.Panels[i])
	}
	return &logicalDeviceData{
		Name:   in.DisplayName,
		Panels: panels,
	}
}
func logicalDeviceDataAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: panelAttrTypes()}},
	}
}

func logicalDevicePanelSchema() attr.Type {
	return types.ListType{
		ElemType: logicalDeviceDataPanelObjectSchema(),
	}
}

func logicalDeviceDataPanelObjectSchema() types.ObjectType {
	return types.ObjectType{
		AttrTypes: panelAttrTypes(),
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
