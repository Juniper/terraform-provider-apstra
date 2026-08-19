package design

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type logicalDeviceDefinition struct {
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (ldd logicalDeviceDefinition) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: DeprecatedLogicalDevicePanel{}.AttrTypes()}},
	}
}

func (ldd logicalDeviceDefinition) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
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

func (ldd logicalDeviceDefinition) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
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

func (ldd logicalDeviceDefinition) asObject(ctx context.Context, diags *diag.Diagnostics) types.Object {
	result, d := types.ObjectValueFrom(ctx, ldd.attrTypes(), ldd)
	diags.Append(d...)
	return result
}
