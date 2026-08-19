package design

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/design"
	_ "github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type LogicalDevice struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Panels     types.List   `tfsdk:"panels"`
	Definition types.Object `tfsdk:"definition"`
}

func (ld LogicalDevice) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":         types.StringType,
		"name":       types.StringType,
		"panels":     types.ListType{ElemType: types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()}},
		"definition": types.ObjectType{AttrTypes: logicalDeviceDefinition{}.attrTypes()},
	}
}

func (ld LogicalDevice) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Logical Device. Required when `name` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{path.MatchRoot("name")}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Logical Device name displayed in the Apstra web UI. Required when `id` is omitted.",
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
		"definition": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Used in nested contexts.",
			Computed:            true,
			Attributes:          logicalDeviceDefinition{}.dataSourceAttributes(),
		},
	}
}

func (ld LogicalDevice) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Logical Device.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Logical Device name displayed in the Apstra web UI.",
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
		"definition": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Used in nested contexts.",
			Computed:            true,
			Attributes:          logicalDeviceDefinition{}.resourceAttributes(),
		},
	}
}

func (ld LogicalDevice) Validate(ctx context.Context, diags *diag.Diagnostics) {
	if ld.Panels.IsUnknown() {
		return // cannot validate unknown panels
	}

	for i, v := range ld.Panels.Elements() {
		if v.IsUnknown() {
			return // cannot validate unknown panel
		}

		var p logicalDevicePanel
		diags.Append(v.(types.Object).As(ctx, &p, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})...)
		if diags.HasError() {
			continue
		}

		p.validate(ctx, path.Root("panels").AtListIndex(i), diags)
	}
}

func (ld LogicalDevice) Request(ctx context.Context, diags *diag.Diagnostics) design.LogicalDevice {
	var result design.LogicalDevice
	if utils.HasValue(ld.ID) {
		result = design.NewLogicalDevice(ld.ID.ValueString())
	}

	result.Label = ld.Name.ValueString()
	result.Panels = make([]design.LogicalDevicePanel, 0, ld.Panels.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true, UnhandledUnknownAsZero: true}))

	for _, v := range ld.Panels.Elements() {
		var panel logicalDevicePanel
		diags.Append(v.(types.Object).As(ctx, &panel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return result
		}
		result.Panels = append(result.Panels, panel.request(ctx, diags))
	}

	return result
}

func (ld *LogicalDevice) LoadAPIData(ctx context.Context, in design.LogicalDevice, diags *diag.Diagnostics) {
	ld.ID = types.StringPointerValue(in.ID())
	ld.Name = types.StringValue(in.Label)
	ld.Panels = NewLogicalDevicePanelList(ctx, in.Panels, diags)
	ld.Definition = ld.DefinitionAsObject(ctx, diags)
}

func (ld LogicalDevice) DefinitionAsObject(ctx context.Context, diags *diag.Diagnostics) types.Object {
	return logicalDeviceDefinition{
		Name:   ld.Name,
		Panels: ld.Panels,
	}.asObject(ctx, diags)
}
