package design

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
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

type DeprecatedLogicalDevice struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (o DeprecatedLogicalDevice) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Logical Device. Required when `name` is omitted.",
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
			MarkdownDescription: "Web UI name of the Logical Device. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"panels": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Details physical layout of interfaces on the device.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DeprecatedLogicalDevicePanel{}.DataSourceAttributes(),
			},
		},
	}
}

func (o DeprecatedLogicalDevice) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
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
				Attributes: DeprecatedLogicalDevicePanel{}.DataSourceAttributes(),
			},
		},
	}
}

func (o DeprecatedLogicalDevice) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID number of the Logical Device",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Logical Device name displayed in the Apstra web UI",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"panels": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Details physical layout of interfaces on the device.",
			Required:            true,
			Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DeprecatedLogicalDevicePanel{}.ResourceAttributes(),
			},
		},
	}
}

func (o DeprecatedLogicalDevice) ResourceAttributesNested() map[string]resourceSchema.Attribute {
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
				Attributes: DeprecatedLogicalDevicePanel{}.ResourceAttributesReadOnly(),
			},
		},
	}
}

func (o DeprecatedLogicalDevice) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"name":   types.StringType,
		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: DeprecatedLogicalDevicePanel{}.AttrTypes()}},
	}
}

func (o *DeprecatedLogicalDevice) LoadApiData(ctx context.Context, in *apstra.LogicalDeviceData, diags *diag.Diagnostics) {
	panels := make([]DeprecatedLogicalDevicePanel, len(in.Panels))
	for i, panel := range in.Panels {
		panels[i].LoadApiData(ctx, &panel, diags)
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Panels = NewDeprecatedLogicalDevicePanelList(ctx, in.Panels, diags)

	if len(panels) > 0 {
		var d diag.Diagnostics
		o.Panels, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DeprecatedLogicalDevicePanel{}.AttrTypes()}, panels)
		diags.Append(d...)
	} else {
		o.Panels = types.ListNull(types.ObjectType{AttrTypes: DeprecatedLogicalDevicePanel{}.AttrTypes()})
	}
}

func (o *DeprecatedLogicalDevice) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.LogicalDeviceData {
	var d diag.Diagnostics
	var panelElements []DeprecatedLogicalDevicePanel
	d = o.Panels.ElementsAs(ctx, &panelElements, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	panels := make([]apstra.LogicalDevicePanel, len(panelElements))
	for i, panel := range panelElements {
		panels[i] = *panel.Request(ctx, diags)
	}
	return &apstra.LogicalDeviceData{
		DisplayName: o.Name.ValueString(),
		Panels:      panels,
	}
}

func (o *DeprecatedLogicalDevice) GetPanels(ctx context.Context, diags *diag.Diagnostics) []DeprecatedLogicalDevicePanel {
	panels := make([]DeprecatedLogicalDevicePanel, len(o.Panels.Elements()))
	diags.Append(o.Panels.ElementsAs(ctx, &panels, false)...)
	return panels
}

func NewLogicalDeviceObject(ctx context.Context, in *apstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(DeprecatedLogicalDevice{}.AttrTypes())
	}

	var ld DeprecatedLogicalDevice
	ld.Id = types.StringNull()
	ld.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(DeprecatedLogicalDevice{}.AttrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, DeprecatedLogicalDevice{}.AttrTypes(), &ld)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(DeprecatedLogicalDevice{}.AttrTypes())
	}

	return result
}
