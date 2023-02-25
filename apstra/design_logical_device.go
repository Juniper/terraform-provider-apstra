package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type logicalDevice struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (o logicalDevice) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Logical Device by ID. Required when `name`is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Logical Device by name. Required when `id`is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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

func (o logicalDevice) dataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
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

func (o logicalDevice) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID number of the resource pool",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Pool name displayed in the Apstra web UI",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"panels": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Details physical layout of interfaces on the device.",
			Required:            true,
			Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: logicalDevicePanel{}.resourceAttributes(),
			},
		},
	}
}

func (o logicalDevice) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"name":   types.StringType,
		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()}},
	}
}

func (o *logicalDevice) loadApiData(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) {
	panels := make([]logicalDevicePanel, len(in.Panels))
	for i, panel := range in.Panels {
		panels[i].loadApiResponse(ctx, &panel, diags)
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Panels = newLogicalDevicePanelList(ctx, in.Panels, diags)

	if len(panels) > 0 {
		var d diag.Diagnostics
		o.Panels, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()}, panels)
		diags.Append(d...)
	} else {
		o.Panels = types.ListNull(types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()})
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
		o.Panels, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()}, panels)
		diags.Append(d...)
	} else {
		o.Panels = types.ListNull(types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()})
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

func (o *logicalDevice) panels(ctx context.Context, diags *diag.Diagnostics) []logicalDevicePanel {
	panels := make([]logicalDevicePanel, len(o.Panels.Elements()))
	diags.Append(o.Panels.ElementsAs(ctx, &panels, false)...)
	return panels
}

func newLogicalDeviceObject(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(logicalDevice{}.attrTypes())
	}

	var ld logicalDevice
	ld.Id = types.StringNull()
	ld.loadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(logicalDevice{}.attrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, logicalDevice{}.attrTypes(), &ld)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(logicalDevice{}.attrTypes())
	}

	return result
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
		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()}},
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
		return types.ObjectNull(logicalDeviceData{}.attrTypes())
	}

	var ld logicalDeviceData
	ld.loadApiResponse(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(logicalDeviceData{}.attrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, ld.attrTypes(), &ld)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(logicalDeviceData{}.attrTypes())
	}

	return result
}

func newLogicalDevicePortGroupList(ctx context.Context, in []goapstra.LogicalDevicePortGroup, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: logicalDevicePanelPortGroup{}.attrTypes()})
	}

	portGroups := make([]logicalDevicePanelPortGroup, len(in))
	for i, portGroup := range in {
		portRoles, d := types.SetValueFrom(ctx, types.StringType, portGroup.Roles.Strings())
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: logicalDevicePanelPortGroup{}.attrTypes()})
		}
		portGroups[i] = logicalDevicePanelPortGroup{
			PortCount: types.Int64Value(int64(portGroup.Count)),
			PortSpeed: types.StringValue(string(portGroup.Speed)),
			PortRoles: portRoles,
		}
	}

	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: logicalDevicePanelPortGroup{}.attrTypes()}, portGroups)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: logicalDevicePanelPortGroup{}.attrTypes()})
	}
	return result
}
