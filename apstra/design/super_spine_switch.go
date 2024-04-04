package design

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SuperSpine struct {
	LogicalDeviceId types.String `tfsdk:"logical_device_id"`
	LogicalDevice   types.Object `tfsdk:"logical_device"`
	PlaneCount      types.Int64  `tfsdk:"plane_count"`
	PerPlaneCount   types.Int64  `tfsdk:"per_plane_count"`
	TagIds          types.Set    `tfsdk:"tag_ids"`
	Tags            types.Set    `tfsdk:"tags"`
}

func (o SuperSpine) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id": types.StringType,
		"logical_device":    types.ObjectType{AttrTypes: LogicalDevice{}.AttrTypes()},
		"plane_count":       types.Int64Type,
		"per_plane_count":   types.Int64Type,
		"tag_ids":           types.SetType{ElemType: types.StringType},
		"tags":              types.SetType{ElemType: types.ObjectType{AttrTypes: Tag{}.AttrTypes()}},
	}
}

func (o SuperSpine) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
		"plane_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of planes.",
			Computed:            true,
		},
		"per_plane_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Super Spine switches per plane.",
			Computed:            true,
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in data source contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to the Super Spine Switches.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: Tag{}.DataSourceAttributesNested(),
			},
		},
	}
}

func (o SuperSpine) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Spine Switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          LogicalDevice{}.ResourceAttributesNested(),
		},
		"plane_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Permits creation of multi-planar 5-stage topologies. Default: 1",
			Computed:            true,
			Optional:            true,
			Default:             int64default.StaticInt64(1),
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"per_plane_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Super Spine switches per plane.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"tag_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to SuperSpine Switches",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to SuperSpine Switches",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: Tag{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o *SuperSpine) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.TemplateElementSuperspineRequest {
	tagIds := make([]apstra.ObjectId, len(o.TagIds.Elements()))
	diags.Append(o.TagIds.ElementsAs(ctx, &tagIds, false)...)
	if diags.HasError() {
		return nil
	}

	return &apstra.TemplateElementSuperspineRequest{
		PlaneCount:         int(o.PlaneCount.ValueInt64()),
		SuperspinePerPlane: int(o.PerPlaneCount.ValueInt64()),
		LogicalDeviceId:    apstra.ObjectId(o.LogicalDeviceId.ValueString()),
		Tags:               tagIds,
	}
}

func (o *SuperSpine) LoadApiData(ctx context.Context, in *apstra.Superspine, diags *diag.Diagnostics) {
	o.LogicalDevice = NewLogicalDeviceObject(ctx, &in.LogicalDevice, diags)
	o.PlaneCount = types.Int64Value(int64(in.PlaneCount))
	o.PerPlaneCount = types.Int64Value(int64(in.SuperspinePerPlane))
	o.Tags = NewTagSet(ctx, in.Tags, diags)
}

func (o *SuperSpine) CopyWriteOnlyElements(ctx context.Context, src *SuperSpine, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "SuperSpine.CopyWriteOnlyElements: attempt to copy from nil source")
		return
	}
	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = utils.SetValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)
}

func NewDesignTemplateSuperSpineObject(ctx context.Context, in *apstra.Superspine, diags *diag.Diagnostics) types.Object {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to generate SuperSpine object from nil source")
		return types.ObjectNull(Spine{}.AttrTypes())
	}

	var ss SuperSpine
	ss.LogicalDeviceId = types.StringNull()
	ss.TagIds = types.SetNull(types.StringType)

	ss.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(SuperSpine{}.AttrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, ss.AttrTypes(), &ss)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(SuperSpine{}.AttrTypes())
	}

	return result
}
