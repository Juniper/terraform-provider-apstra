package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type logicalDevice struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (o logicalDevice) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"name":   types.StringType,
		"panels": types.ListType{ElemType: o.attrType()},
	}
}

func (o logicalDevice) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *logicalDevice) loadApiResponse(ctx context.Context, in *goapstra.LogicalDevice, diags *diag.Diagnostics) {
	panels := make([]logicalDevicePanel, len(in.Data.Panels))
	for i, panel := range in.Data.Panels {
		panels[i].loadApiResponse(ctx, &panel, diags)
		if diags.HasError() {
			return
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)

	if len(panels) > 0 {
		var d diag.Diagnostics
		o.Panels, d = types.ListValueFrom(ctx, logicalDevicePanel{}.attrType(), panels)
		diags.Append(d...)
	} else {
		o.Panels = types.ListNull(logicalDevicePanel{}.attrType())
	}
}

func (o *logicalDevice) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.LogicalDeviceData {
	var d diag.Diagnostics
	var panelElements []logicalDevicePanel
	d = o.Panels.ElementsAs(ctx, &panelElements, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	panels := make([]goapstra.LogicalDevicePanel, len(panelElements))
	for i, panel := range panelElements {
		panels[i] = *panel.request(ctx, diags)
	}
	return &goapstra.LogicalDeviceData{
		DisplayName: o.Name.ValueString(),
		Panels:      panels,
	}
}

type logicalDevicePanel struct {
	Rows       types.Int64 `tfsdk:"rows"`
	Columns    types.Int64 `tfsdk:"columns"`
	PortGroups types.List  `tfsdk:"port_groups"`
}

func (o logicalDevicePanel) attrType() attr.Type {
	var portGroup logicalDevicePanelPortGroup
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"rows":        types.Int64Type,
			"columns":     types.Int64Type,
			"port_groups": types.ListType{ElemType: portGroup.attrType()},
		},
	}
}

func (o *logicalDevicePanel) loadApiResponse(ctx context.Context, in *goapstra.LogicalDevicePanel, diags *diag.Diagnostics) {
	var portGroups []logicalDevicePanelPortGroup

	portGroups = make([]logicalDevicePanelPortGroup, len(in.PortGroups))
	for i := range in.PortGroups {
		portGroups[i].loadApiResponse(ctx, &in.PortGroups[i], diags)
		if diags.HasError() {
			return
		}
	}

	var bogusPG logicalDevicePanelPortGroup
	portGroupList, d := types.ListValueFrom(ctx, bogusPG.attrType(), portGroups)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.Rows = types.Int64Value(int64(in.PanelLayout.RowCount))
	o.Columns = types.Int64Value(int64(in.PanelLayout.ColumnCount))
	o.PortGroups = portGroupList
}

func (o *logicalDevicePanel) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.LogicalDevicePanel {
	tfPortGroups := make([]logicalDevicePanelPortGroup, len(o.PortGroups.Elements()))
	diags.Append(o.PortGroups.ElementsAs(ctx, &tfPortGroups, false)...)
	if diags.HasError() {
		return nil
	}

	reqPortGroups := make([]goapstra.LogicalDevicePortGroup, len(tfPortGroups))
	for i, pg := range tfPortGroups {
		roleStrings := make([]string, len(pg.PortRoles.Elements()))
		diags.Append(pg.PortRoles.ElementsAs(ctx, &roleStrings, false)...)
		if diags.HasError() {
			return nil
		}
		var reqRoles goapstra.LogicalDevicePortRoleFlags
		err := reqRoles.FromStrings(roleStrings)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing port roles: '%s'", strings.Join(roleStrings, "','")), err.Error())
		}
		reqPortGroups[i] = goapstra.LogicalDevicePortGroup{
			Count: int(pg.PortCount.ValueInt64()),
			Speed: goapstra.LogicalDevicePortSpeed(pg.PortSpeed.ValueString()),
			Roles: reqRoles,
		}
	}

	return &goapstra.LogicalDevicePanel{
		PanelLayout: goapstra.LogicalDevicePanelLayout{
			RowCount:    int(o.Rows.ValueInt64()),
			ColumnCount: int(o.Columns.ValueInt64()),
		},
		PortIndexing: goapstra.LogicalDevicePortIndexing{
			Order:      goapstra.PortIndexingHorizontalFirst,
			StartIndex: 1,
			Schema:     goapstra.PortIndexingSchemaAbsolute,
		},
		PortGroups: reqPortGroups,
	}
}

func (o *logicalDevicePanel) validate(ctx context.Context, i int, diags *diag.Diagnostics) {
	var panelPortsByDimensions, panelPortsByPortGroup int64
	panelPortsByDimensions = o.Rows.ValueInt64() * o.Columns.ValueInt64()

	portGroups := make([]logicalDevicePanelPortGroup, len(o.PortGroups.Elements()))
	diags.Append(o.PortGroups.ElementsAs(ctx, &portGroups, false)...)
	if diags.HasError() {
		return
	}

	for _, portGroup := range portGroups {
		panelPortsByPortGroup = panelPortsByPortGroup + portGroup.PortCount.ValueInt64()
	}
	if panelPortsByDimensions != panelPortsByPortGroup {
		diags.AddAttributeError(path.Root("panels").AtListIndex(i),
			errInvalidConfig,
			fmt.Sprintf("panel %d (%d by %d ports) has %d ports by dimensions, but the total by port group is %d",
				i, o.Rows.ValueInt64(), o.Columns.ValueInt64(), panelPortsByDimensions, panelPortsByPortGroup))
		return
	}
}

type logicalDevicePanelPortGroup struct {
	PortCount types.Int64  `tfsdk:"port_count"`
	PortSpeed types.String `tfsdk:"port_speed"`
	PortRoles types.Set    `tfsdk:"port_roles"`
}

func (o *logicalDevicePanelPortGroup) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"port_count": types.Int64Type,
			"port_speed": types.StringType,
			"port_roles": types.SetType{ElemType: types.StringType}}}
}

func (o *logicalDevicePanelPortGroup) loadApiResponse(ctx context.Context, in *goapstra.LogicalDevicePortGroup, diags *diag.Diagnostics) {
	portRoles, d := types.SetValueFrom(ctx, types.StringType, in.Roles.Strings())
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.PortCount = types.Int64Value(int64(in.Count))
	o.PortSpeed = types.StringValue(string(in.Speed))
	o.PortRoles = portRoles
}

// everything below here is suspect...

//func logicalDeviceAttrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: logicalDeviceDataAttrTypes()}
//}

//func (o logicalDeviceData) attrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: map[string]attr.Type{
//			"name": types.StringType,
//			"panels": types.ListType{
//				ElemType: logicalDevicePanel{}.attrType()}}}
//}

//func (o *logicalDeviceData) loadApiResponse(in *goapstra.LogicalDeviceData) {
//	o.Name = in.DisplayName
//	o.Panels = make([]logicalDevicePanel, len(in.Panels))
//
//	for i := range o.Panels {
//		o.Panels[i].loadApiResponse(&in.Panels[i])
//	}
//}

//func parseApiLogicalDeviceData(in *goapstra.LogicalDeviceData) *logicalDeviceData {
//	panels := make([]logicalDevicePanel, len(in.Panels))
//	for i := range in.Panels {
//		panels[i].loadApiResponse(&in.Panels[i])
//	}
//	return &logicalDeviceData{
//		Name:   in.DisplayName,
//		Panels: panels,
//	}
//}

//func logicalDeviceDataAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"name":   types.StringType,
//		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: panelAttrTypes()}},
//	}
//}

//func parseApiLogicalDeviceToTypesObject(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
//	structLogicalDeviceData := parseApiLogicalDeviceData(in)
//	result, d := types.ObjectValueFrom(ctx, logicalDeviceDataAttrTypes(), structLogicalDeviceData)
//	diags.Append(d...)
//	return result
//}

//func panelPortGroupAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"port_count": types.Int64Type,
//		"port_speed": types.StringType,
//		"port_roles": types.SetType{ElemType: types.StringType}}
//}

//func panelAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"rows":        types.Int64Type,
//		"columns":     types.Int64Type,
//		"port_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: panelPortGroupAttrTypes()}}}
//}

type logicalDeviceData struct {
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (o logicalDeviceData) schema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Logical device display name.",
				Computed:            true,
			},
			"panels": dPanelAttributeSchema(),
		},
	}
}

func (o logicalDeviceData) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"panels": types.ListType{ElemType: logicalDevicePanel{}.attrType()},
	}
}

func (o logicalDeviceData) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *logicalDeviceData) loadApiResponse(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) {
	panels := make([]logicalDevicePanel, len(in.Panels))
	for i, panel := range in.Panels {
		panels[i].loadApiResponse(ctx, &panel, diags)
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)

	if len(panels) > 0 {
		var d diag.Diagnostics
		o.Panels, d = types.ListValueFrom(ctx, logicalDevicePanel{}.attrType(), panels)
		diags.Append(d...)
	} else {
		o.Panels = types.ListNull(logicalDevicePanel{}.attrType())
	}
}

func newLogicalDeviceObject(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(logicalDevice{}.attrTypes())
	}

	var ld logicalDeviceData
	ld.loadApiResponse(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(logicalDevice{}.attrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, ld.attrTypes(), &ld)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(logicalDevice{}.attrTypes())
	}

	return result
}
