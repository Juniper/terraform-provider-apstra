package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
