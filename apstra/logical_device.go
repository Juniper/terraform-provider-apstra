package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func (o logicalDevicePanel) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"rows": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Physical vertical dimension of the panel.",
			Computed:            true,
		},
		"columns": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Physical horizontal dimension of the panel.",
			Computed:            true,
		},
		"port_groups": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: logicalDevicePanelPortGroup{}.dataSourceAttributes(),
			},
		},
	}
}

func (o logicalDevicePanel) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"rows": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical vertical dimension of the panel.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"columns": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical horizontal dimension of the panel.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"port_groups": resourceSchema.ListNestedAttribute{
			Required:            true,
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: logicalDevicePanelPortGroup{}.schemaAsResource(),
			},
		},
	}
}

func (o logicalDevicePanel) resourceAttributesReadOnly() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"rows": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical vertical dimension of the panel.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"columns": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical horizontal dimension of the panel.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"port_groups": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Computed:            true,
			PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: logicalDevicePanelPortGroup{}.resourceAttributesReadOnly(),
			},
		},
	}
}

func (o logicalDevicePanel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":        types.Int64Type,
		"columns":     types.Int64Type,
		"port_groups": types.ListType{ElemType: logicalDevicePanelPortGroup{}.attrType()},
	}
}

func (o logicalDevicePanel) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
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

func (o logicalDevicePanelPortGroup) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"port_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of ports in the group.",
			Computed:            true,
		},
		"port_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Port speed.",
			Computed:            true,
		},
		"port_roles": dataSourceSchema.SetAttribute{
			MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o logicalDevicePanelPortGroup) schemaAsResource() map[string]resourceSchema.Attribute {
	var allRoleFlagsSet goapstra.LogicalDevicePortRoleFlags
	allRoleFlagsSet.SetAll()

	return map[string]resourceSchema.Attribute{
		"port_count": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "Number of ports in the group.",
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"port_speed": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Port speed.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(2),
			},
		},
		"port_roles": schema.SetAttribute{
			ElementType:         types.StringType,
			Required:            true,
			MarkdownDescription: fmt.Sprintf("One or more of: '%s'", strings.Join(allRoleFlagsSet.Strings(), "', '")),
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.OneOf(allRoleFlagsSet.Strings()...)),
			},
		},
	}
}

func (o logicalDevicePanelPortGroup) resourceAttributesReadOnly() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"port_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of ports in the group.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"port_speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Port speed.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"port_roles": resourceSchema.SetAttribute{
			MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		},
	}
}

func (o logicalDevicePanelPortGroup) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port_count": types.Int64Type,
		"port_speed": types.StringType,
		"port_roles": types.SetType{ElemType: types.StringType},
	}
}

func (o logicalDevicePanelPortGroup) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
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

func (o logicalDeviceData) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Logical device display name.",
			Computed:            true,
		},
		"panels": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Details physical layout of interfaces on the device.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: logicalDevicePanel{}.dataSourceAttributes(),
			},
		},
	}
}

func (o logicalDeviceData) schemaAsResourceReadOnly() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Logical device display name.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"panels": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Details physical layout of interfaces on the device.",
			Computed:            true,
			PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: logicalDevicePanel{}.resourceAttributesReadOnly(),
			},
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
	o.Name = types.StringValue(in.DisplayName)
	o.Panels = newLogicalDevicePanelList(ctx, in.Panels, diags)
	if diags.HasError() {
		return
	}
}

func newLogicalDeviceDataObject(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
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

func newLogicalDevicePanelList(ctx context.Context, in []goapstra.LogicalDevicePanel, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(logicalDevicePanel{}.attrType())
	}

	panels := make([]logicalDevicePanel, len(in))
	for i, panel := range in {
		panels[i] = logicalDevicePanel{
			Rows:       types.Int64Value(int64(panel.PanelLayout.RowCount)),
			Columns:    types.Int64Value(int64(panel.PanelLayout.ColumnCount)),
			PortGroups: newLogicalDevicePortGroupList(ctx, panel.PortGroups, diags),
		}
		if diags.HasError() {
			return types.ListNull(logicalDevicePanel{}.attrType())
		}
	}

	result, d := types.ListValueFrom(ctx, logicalDevicePanel{}.attrType(), &panels)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(logicalDevicePanel{}.attrType())
	}

	return result
}

func newLogicalDevicePortGroupList(ctx context.Context, in []goapstra.LogicalDevicePortGroup, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(logicalDevicePanelPortGroup{}.attrType())
	}

	portGroups := make([]logicalDevicePanelPortGroup, len(in))
	for i, portGroup := range in {
		portRoles, d := types.SetValueFrom(ctx, types.StringType, portGroup.Roles.Strings())
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(logicalDevicePanelPortGroup{}.attrType())
		}
		portGroups[i] = logicalDevicePanelPortGroup{
			PortCount: types.Int64Value(int64(portGroup.Count)),
			PortSpeed: types.StringValue(string(portGroup.Speed)),
			PortRoles: portRoles,
		}
	}

	result, d := types.ListValueFrom(ctx, logicalDevicePanelPortGroup{}.attrType(), portGroups)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(logicalDevicePanelPortGroup{}.attrType())
	}
	return result
}
