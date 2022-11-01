package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type logicalDevice struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (o logicalDevice) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":     types.StringType,
			"name":   types.StringType,
			"panels": types.ListType{ElemType: logicalDevicePanel{}.attrType()}}}
}

func (o *logicalDevice) parseApi(ctx context.Context, in *goapstra.LogicalDevice, diags *diag.Diagnostics) {
	var panels []logicalDevicePanel

	if len(in.Data.Panels) != 0 {
		panels = make([]logicalDevicePanel, len(in.Data.Panels))
		for i := range in.Data.Panels {
			panels[i].parseApi(&in.Data.Panels[i])
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)

	var d diag.Diagnostics
	o.Panels, d = types.ListValueFrom(ctx, logicalDevicePanel{}.attrType(), panels)

	diags.Append(d...)
}

func (o *logicalDevice) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.LogicalDeviceData {
	var elements []logicalDevicePanel
	o.Panels.ElementsAs(ctx, &elements, false)
	panels := make([]goapstra.LogicalDevicePanel, len(elements))
	for i, panel := range elements {
		panels[i] = *panel.request(diags)
	}
	return &goapstra.LogicalDeviceData{
		DisplayName: o.Name.ValueString(),
		Panels:      panels,
	}
}

type logicalDevicePanel struct {
	Rows       int64                         `tfsdk:"rows"`
	Columns    int64                         `tfsdk:"columns"`
	PortGroups []logicalDevicePanelPortGroup `tfsdk:"port_groups"`
}

func (o logicalDevicePanel) attrType() attr.Type {
	var portGroups logicalDevicePanelPortGroup
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"rows":        types.Int64Type,
			"columns":     types.Int64Type,
			"port_groups": types.ListType{ElemType: portGroups.attrType()}}}
}

func (o *logicalDevicePanel) parseApi(in *goapstra.LogicalDevicePanel) {
	var portGroups []logicalDevicePanelPortGroup

	if len(in.PortGroups) != 0 {
		portGroups = make([]logicalDevicePanelPortGroup, len(in.PortGroups))
		for i := range in.PortGroups {
			portGroups[i].parseApi(&in.PortGroups[i])
		}
	}

	o.Rows = int64(in.PanelLayout.RowCount)
	o.Columns = int64(in.PanelLayout.ColumnCount)
	o.PortGroups = portGroups
}

func (o *logicalDevicePanel) request(diags *diag.Diagnostics) *goapstra.LogicalDevicePanel {
	portGroups := make([]goapstra.LogicalDevicePortGroup, len(o.PortGroups))
	for i, pg := range o.PortGroups {
		var roles goapstra.LogicalDevicePortRoleFlags
		err := roles.FromStrings(pg.PortRoles)
		if err != nil {
			diags.AddError("error parsing port roles", err.Error())
		}

		portGroups[i] = goapstra.LogicalDevicePortGroup{
			Count: int(pg.PortCount),
			Speed: goapstra.LogicalDevicePortSpeed(pg.PortSpeed),
			Roles: roles,
		}
	}

	return &goapstra.LogicalDevicePanel{
		PanelLayout: goapstra.LogicalDevicePanelLayout{
			RowCount:    int(o.Rows),
			ColumnCount: int(o.Columns),
		},
		PortIndexing: goapstra.LogicalDevicePortIndexing{
			Order:      goapstra.PortIndexingHorizontalFirst,
			StartIndex: 1,
			Schema:     goapstra.PortIndexingSchemaAbsolute,
		},
		PortGroups: portGroups,
	}
}

type logicalDevicePanelPortGroup struct {
	PortCount int64    `tfsdk:"port_count"`
	PortSpeed string   `tfsdk:"port_speed"`
	PortRoles []string `tfsdk:"port_roles"`
}

func (o *logicalDevicePanelPortGroup) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"port_count": types.Int64Type,
			"port_speed": types.StringType,
			"port_roles": types.SetType{ElemType: types.StringType}}}
}

func (o *logicalDevicePanelPortGroup) parseApi(in *goapstra.LogicalDevicePortGroup) {
	o.PortCount = int64(in.Count)
	o.PortSpeed = string(in.Speed)
	o.PortRoles = in.Roles.Strings()
}

// everything below here is suspect...

func logicalDeviceAttrType() attr.Type {
	return types.ObjectType{
		AttrTypes: logicalDeviceDataAttrTypes()}
}

type logicalDeviceData struct {
	Name   string               `tfsdk:"name"`
	Panels []logicalDevicePanel `tfsdk:"panels"'`
}

func (o logicalDeviceData) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"panels": types.ListType{
				ElemType: logicalDevicePanel{}.attrType()}}}
}

func (o *logicalDeviceData) parseApi(in *goapstra.LogicalDeviceData) {
	o.Name = in.DisplayName
	o.Panels = make([]logicalDevicePanel, len(in.Panels))

	for i := range o.Panels {
		o.Panels[i].parseApi(&in.Panels[i])
	}
}

func parseApiLogicalDeviceData(in *goapstra.LogicalDeviceData) *logicalDeviceData {
	panels := make([]logicalDevicePanel, len(in.Panels))
	for i := range in.Panels {
		panels[i].parseApi(&in.Panels[i])
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

func logicalDeviceDataAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
		Computed:            true,
		PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"panels": dPanelsAttributeSchema(),
			"name": {
				MarkdownDescription: "Logical device display name.",
				Computed:            true,
				Type:                types.StringType,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
		}),
	}
}

func parseApiLogicalDeviceToTypesObject(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
	structLogicalDeviceData := parseApiLogicalDeviceData(in)
	result, d := types.ObjectValueFrom(ctx, logicalDeviceDataAttrTypes(), structLogicalDeviceData)
	diags.Append(d...)
	return result
}

func panelPortGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port_count": types.Int64Type,
		"port_speed": types.StringType,
		"port_roles": types.SetType{ElemType: types.StringType}}
}

func panelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":        types.Int64Type,
		"columns":     types.Int64Type,
		"port_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: panelPortGroupAttrTypes()}}}
}
