package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type spine struct {
	LogicalDeviceId     types.String `tfsdk:"logical_device_id"`
	LogicalDevice       types.Object `tfsdk:"logical_device"`
	Count               types.Int64  `tfsdk:"count"`
	SuperSpineLinkSpeed types.String `tfsdk:"super_spine_link_speed"`
	SuperSpineLinkCount types.Int64  `tfsdk:"super_spine_link_count"`
	TagIds              types.Set    `tfsdk:"tag_ids"`
	Tags                types.Set    `tfsdk:"tags"`
}

func (o spine) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"logical_device_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in data source contexts.",
			Computed:            true,
		},
		"logical_device": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDevice{}.dataSourceAttributesNested(),
		},
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
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in data source contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to the Spine Switches.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o spine) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Spine Switch.",
			Required:            true,
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDevice{}.resourceAttributesNested(),
		},
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Spine Switches.",
			Required:            true,
		},
		"super_spine_link_speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of links to super spine switches.",
			Optional:            true,
		},
		"super_spine_link_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Count of links to each super spine switch.",
			Optional:            true,
		},
		"tag_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Access Switch",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Spine Switch",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: tag{}.resourceAttributesNested(),
			},
		},
	}
}

func (o spine) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":      types.StringType,
		"logical_device":         types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"count":                  types.Int64Type,
		"super_spine_link_speed": types.StringType,
		"super_spine_link_count": types.Int64Type,
		"tag_ids":                types.SetType{ElemType: types.StringType},
		"tags":                   types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *spine) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.TemplateElementSpineRequest {
	tagIds := make([]goapstra.ObjectId, len(o.TagIds.Elements()))
	d := o.TagIds.ElementsAs(ctx, &tagIds, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return &goapstra.TemplateElementSpineRequest{
		Count:                  int(o.Count.ValueInt64()),
		LinkPerSuperspineSpeed: goapstra.LogicalDevicePortSpeed(o.SuperSpineLinkSpeed.ValueString()),
		LogicalDevice:          goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
		LinkPerSuperspineCount: int(o.SuperSpineLinkCount.ValueInt64()),
		Tags:                   tagIds,
	}
}

func (o *spine) loadApiData(ctx context.Context, in *goapstra.Spine, diags *diag.Diagnostics) {
	o.LogicalDevice = newLogicalDeviceObject(ctx, &in.LogicalDevice, diags)
	o.Count = types.Int64Value(int64(in.Count))

	if in.LinkPerSuperspineSpeed == "" {
		o.SuperSpineLinkSpeed = types.StringNull()
		o.SuperSpineLinkCount = types.Int64Null()
	} else {
		o.SuperSpineLinkSpeed = types.StringValue(string(in.LinkPerSuperspineSpeed))
		o.SuperSpineLinkCount = types.Int64Value(int64(in.LinkPerSuperspineCount))
	}

	o.TagIds = types.SetNull(types.StringType)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

func (o *spine) copyWriteOnlyElements(ctx context.Context, src *spine, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "spine.copyWriteOnlyElements: attempt to copy from nil source")
	}
	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = setValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)
}

func newDesignTemplateSpineObject(ctx context.Context, in *goapstra.Spine, diags *diag.Diagnostics) types.Object {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to generate spine object from nil source")
		return types.ObjectNull(spine{}.attrTypes())
	}

	var s spine
	s.LogicalDeviceId = types.StringNull()
	s.loadApiData(ctx, in, diags)
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
