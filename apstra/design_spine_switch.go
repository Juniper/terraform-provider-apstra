package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type spine struct {
	Count               types.Int64  `tfsdk:"count"`
	SuperSpineLinkSpeed types.String `tfsdk:"super_spine_link_speed"`
	SuperSpineLinkCount types.Int64  `tfsdk:"super_spine_link_count"`
	LogicalDeviceData   types.Object `tfsdk:"logical_device"`
	Tags                types.Set    `tfsdk:"tags"`
}

func (o spine) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of spine switches.",
			Computed:            true,
		},
		"super_spine_link_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of links to super spine switches.",
			Computed:            true,
		},
		"super_spine_link_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of links to each super spine switch.",
			Computed:            true,
		},
		"logical_device": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDeviceData{}.dataSourceAttributes(),
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to the Spine Switches.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributes(),
			},
		},
	}
}

func (o spine) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"count":                  types.Int64Type,
		"super_spine_link_speed": types.StringType,
		"super_spine_link_count": types.Int64Type,
		"logical_device":         types.ObjectType{AttrTypes: logicalDeviceData{}.attrTypes()},
		"tags":                   types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *spine) loadApiResponse(ctx context.Context, in *goapstra.Spine, diags *diag.Diagnostics) {
	o.Count = types.Int64Value(int64(in.Count))

	if in.LinkPerSuperspineSpeed == "" {
		o.SuperSpineLinkSpeed = types.StringNull()
		o.SuperSpineLinkCount = types.Int64Null()
	} else {
		o.SuperSpineLinkSpeed = types.StringValue(string(in.LinkPerSuperspineSpeed))
		o.SuperSpineLinkCount = types.Int64Value(int64(in.LinkPerSuperspineCount))
	}

	o.LogicalDeviceData = newLogicalDeviceDataObject(ctx, &in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}

	o.Tags = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}
}

func newDesignTemplateSpineObject(ctx context.Context, in *goapstra.Spine, diags *diag.Diagnostics) types.Object {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to generate spine object from nil source")
		return types.ObjectNull(spine{}.attrTypes())
	}

	var s spine
	s.loadApiResponse(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(spine{}.attrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, s.attrTypes(), &s)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(spine{}.attrTypes())
	}

	return result
}
