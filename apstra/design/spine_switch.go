package design

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/utils"
)

type Spine struct {
	LogicalDeviceId     types.String `tfsdk:"logical_device_id"`
	LogicalDevice       types.Object `tfsdk:"logical_device"`
	Count               types.Int64  `tfsdk:"count"`
	SuperSpineLinkSpeed types.String `tfsdk:"super_spine_link_speed"`
	SuperSpineLinkCount types.Int64  `tfsdk:"super_spine_link_count"`
	TagIds              types.Set    `tfsdk:"tag_ids"`
	Tags                types.Set    `tfsdk:"tags"`
}

func (o Spine) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"logical_device_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in data source contexts.",
			Computed:            true,
		},
		"logical_device": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          LogicalDevice{}.DataSourceAttributesNested(),
		},
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Spine switches.",
			Computed:            true,
		},
		"super_spine_link_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of links to super Spine switches.",
			Computed:            true,
		},
		"super_spine_link_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of links to each super Spine switch.",
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
				Attributes: Tag{}.DataSourceAttributesNested(),
			},
		},
	}
}

func (o Spine) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Spine Switch.",
			Required:            true,
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          LogicalDevice{}.ResourceAttributesNested(),
		},
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Spine Switches.",
			Required:            true,
		},
		"super_spine_link_speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of links to super Spine switches.",
			Optional:            true,
			Validators:          []validator.String{apstravalidator.ParseSpeed()},
		},
		"super_spine_link_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Count of links to each super Spine switch.",
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
				Attributes: Tag{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o Spine) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":      types.StringType,
		"logical_device":         types.ObjectType{AttrTypes: LogicalDevice{}.AttrTypes()},
		"count":                  types.Int64Type,
		"super_spine_link_speed": types.StringType,
		"super_spine_link_count": types.Int64Type,
		"tag_ids":                types.SetType{ElemType: types.StringType},
		"tags":                   types.SetType{ElemType: types.ObjectType{AttrTypes: Tag{}.AttrTypes()}},
	}
}

func (o *Spine) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.TemplateElementSpineRequest {
	tagIds := make([]apstra.ObjectId, len(o.TagIds.Elements()))
	d := o.TagIds.ElementsAs(ctx, &tagIds, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return &apstra.TemplateElementSpineRequest{
		Count:                  int(o.Count.ValueInt64()),
		LinkPerSuperspineSpeed: apstra.LogicalDevicePortSpeed(o.SuperSpineLinkSpeed.ValueString()),
		LogicalDevice:          apstra.ObjectId(o.LogicalDeviceId.ValueString()),
		LinkPerSuperspineCount: int(o.SuperSpineLinkCount.ValueInt64()),
		Tags:                   tagIds,
	}
}

func (o *Spine) LoadApiData(ctx context.Context, in *apstra.Spine, diags *diag.Diagnostics) {
	o.LogicalDevice = NewLogicalDeviceObject(ctx, &in.LogicalDevice, diags)
	o.Count = types.Int64Value(int64(in.Count))

	if in.LinkPerSuperspineSpeed == "" {
		o.SuperSpineLinkSpeed = types.StringNull()
		o.SuperSpineLinkCount = types.Int64Null()
	} else {
		o.SuperSpineLinkSpeed = types.StringValue(string(in.LinkPerSuperspineSpeed))
		o.SuperSpineLinkCount = types.Int64Value(int64(in.LinkPerSuperspineCount))
	}

	o.TagIds = types.SetNull(types.StringType)
	o.Tags = NewTagSet(ctx, in.Tags, diags)
}

func (o *Spine) CopyWriteOnlyElements(ctx context.Context, src *Spine, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "Spine.CopyWriteOnlyElements: attempt to copy from nil source")
		return
	}
	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = utils.SetValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)
}

func NewDesignTemplateSpineObject(ctx context.Context, in *apstra.Spine, diags *diag.Diagnostics) types.Object {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to generate Spine object from nil source")
		return types.ObjectNull(Spine{}.AttrTypes())
	}

	var s Spine
	s.LogicalDeviceId = types.StringNull()
	s.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(Spine{}.AttrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, s.AttrTypes(), &s)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(Spine{}.AttrTypes())
	}

	return result
}
