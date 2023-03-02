package design

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

type LogicalDevice struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (o LogicalDevice) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
				Attributes: LogicalDevicePanel{}.DataSourceAttributes(),
			},
		},
	}
}

func (o LogicalDevice) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
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
				Attributes: LogicalDevicePanel{}.DataSourceAttributes(),
			},
		},
	}
}

func (o LogicalDevice) ResourceAttributes() map[string]resourceSchema.Attribute {
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
				Attributes: LogicalDevicePanel{}.ResourceAttributes(),
			},
		},
	}
}

func (o LogicalDevice) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
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
				Attributes: LogicalDevicePanel{}.ResourceAttributesReadOnly(),
			},
		},
	}
}

func (o LogicalDevice) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"name":   types.StringType,
		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: LogicalDevicePanel{}.AttrTypes()}},
	}
}

func (o *LogicalDevice) LoadApiData(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) {
	panels := make([]LogicalDevicePanel, len(in.Panels))
	for i, panel := range in.Panels {
		panels[i].LoadApiData(ctx, &panel, diags)
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Panels = NewLogicalDevicePanelList(ctx, in.Panels, diags)

	if len(panels) > 0 {
		var d diag.Diagnostics
		o.Panels, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: LogicalDevicePanel{}.AttrTypes()}, panels)
		diags.Append(d...)
	} else {
		o.Panels = types.ListNull(types.ObjectType{AttrTypes: LogicalDevicePanel{}.AttrTypes()})
	}
}

func (o *LogicalDevice) Request(ctx context.Context, diags *diag.Diagnostics) *goapstra.LogicalDeviceData {
	var d diag.Diagnostics
	var panelElements []LogicalDevicePanel
	d = o.Panels.ElementsAs(ctx, &panelElements, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	panels := make([]goapstra.LogicalDevicePanel, len(panelElements))
	for i, panel := range panelElements {
		panels[i] = *panel.Request(ctx, diags)
	}
	return &goapstra.LogicalDeviceData{
		DisplayName: o.Name.ValueString(),
		Panels:      panels,
	}
}

func (o *LogicalDevice) GetPanels(ctx context.Context, diags *diag.Diagnostics) []LogicalDevicePanel {
	panels := make([]LogicalDevicePanel, len(o.Panels.Elements()))
	diags.Append(o.Panels.ElementsAs(ctx, &panels, false)...)
	return panels
}

func NewLogicalDeviceObject(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(LogicalDevice{}.AttrTypes())
	}

	var ld LogicalDevice
	ld.Id = types.StringNull()
	ld.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(LogicalDevice{}.AttrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, LogicalDevice{}.AttrTypes(), &ld)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(LogicalDevice{}.AttrTypes())
	}

	return result
}
