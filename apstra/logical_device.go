package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		"port_count":      types.Int64Type,
		"port_speed_gbps": types.Int64Type,
		"port_roles": types.SetType{
			ElemType: types.StringType,
		},
	}
}

func newPanelSetFromSliceLogicalDevicePanel(panels []goapstra.LogicalDevicePanel) types.List {
	result := newPanelList(len(panels))
	for i, panel := range panels {
		result.Elems[i] = types.Object{
			AttrTypes: panelAttrTypes(),
			Attrs: map[string]attr.Value{
				"rows":        types.Int64{Value: int64(panel.PanelLayout.RowCount)},
				"columns":     types.Int64{Value: int64(panel.PanelLayout.ColumnCount)},
				"port_groups": newPortGroupListFromSliceLogicalDevicePortGroup(panel.PortGroups),
			},
		}
	}
	return result
}

func newPortGroupListFromSliceLogicalDevicePortGroup(portGroups []goapstra.LogicalDevicePortGroup) types.List {
	result := newPortGroupList(len(portGroups))
	for i, portGroup := range portGroups {
		result.Elems[i] = types.Object{
			AttrTypes: portGroupAttrTypes(),
			Attrs: map[string]attr.Value{
				"port_count":      types.Int64{Value: int64(portGroup.Count)},
				"port_speed_gbps": types.Int64{Value: portGroup.Speed.BitsPerSecond()},
				"port_roles":      newPortRolesSetFromLogicalDevicePortRoleFlags(portGroup.Roles),
			},
		}
	}
	return result
}

func newPortRolesSetFromLogicalDevicePortRoleFlags(roles goapstra.LogicalDevicePortRoleFlags) types.Set {
	roleStrings := roles.Strings()
	result := newPortRolesSet(len(roleStrings))
	for i, s := range roleStrings {
		result.Elems[i] = types.String{Value: s}
	}
	return result
}

func newPortRolesSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.StringType,
	}
}

func newPortGroupList(size int) types.List {
	return types.List{
		Elems: make([]attr.Value, size),
		ElemType: types.ObjectType{
			AttrTypes: portGroupAttrTypes(),
		},
	}
}

func newPanelList(size int) types.List {
	return types.List{
		Elems: make([]attr.Value, size),
		ElemType: types.ObjectType{
			AttrTypes: panelAttrTypes(),
		},
	}
}

func panelsSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details physical layout of interfaces on the device.",
		Computed:            true,
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"rows": {
				MarkdownDescription: "Physical vertical dimension of the panel.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"columns": {
				MarkdownDescription: "Physical horizontal dimension of the panel.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"port_groups": {
				MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
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
	}
}
