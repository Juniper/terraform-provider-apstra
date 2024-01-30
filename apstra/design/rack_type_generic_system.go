package design

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ValidateGenericSystem(rt *apstra.RackType, i int, diags *diag.Diagnostics) {
	gs := rt.Data.GenericSystems[i]
	if gs.LogicalDevice == nil {
		diags.AddError("generic system logical device info missing",
			fmt.Sprintf("rack type '%s', generic system '%s' logical device is nil",
				rt.Id, gs.Label))
		return
	}
}

type GenericSystem struct {
	LogicalDeviceId  types.String `tfsdk:"logical_device_id"`
	LogicalDevice    types.Object `tfsdk:"logical_device"`
	PortChannelIdMin types.Int64  `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64  `tfsdk:"port_channel_id_max"`
	Count            types.Int64  `tfsdk:"count"`
	Links            types.Map    `tfsdk:"links"`
	TagIds           types.Set    `tfsdk:"tag_ids"`
	Tags             types.Set    `tfsdk:"tags"`
}

func (o GenericSystem) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
		"port_channel_id_min": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"port_channel_id_max": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Generic Systems of this type.",
			Computed:            true,
		},
		"links": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Details links from this Generic System to upstream switches within this Rack Type.",
			Computed:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: RackLink{}.DataSourceAttributes(),
			},
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in data source contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Generic System.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: Tag{}.DataSourceAttributesNested(),
			},
		},
	}
}

func (o GenericSystem) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Generic System.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          LogicalDevice{}.ResourceAttributesNested(),
		},
		"port_channel_id_min": resourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.Between(PoIdMin, PoIdMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("port_channel_id_max")),
				int64validator.AtMostSumOf(path.MatchRelative().AtParent().AtName("port_channel_id_max")),
			},
		},
		"port_channel_id_max": resourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.Between(PoIdMin, PoIdMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("port_channel_id_min")),
				int64validator.AtLeastSumOf(path.MatchRelative().AtParent().AtName("port_channel_id_min")),
			},
		},
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Generic Systems of this type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"links": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Each Generic System is required to have at least one Link to a Leaf Switch or Access Switch.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: RackLink{}.ResourceAttributes(),
			},
		},
		"tag_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Generic System",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Generic System",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: Tag{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o GenericSystem) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          LogicalDevice{}.ResourceAttributesNested(),
		},
		"port_channel_id_min": resourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"port_channel_id_max": resourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Generic Systems of this type.",
			Computed:            true,
		},
		"links": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Each Generic System is required to have at least one Link to a Leaf Switch or Access Switch.",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: RackLink{}.ResourceAttributesNested(),
			},
		},
		"tag_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in nested contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Generic System",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: Tag{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o GenericSystem) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":   types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: LogicalDevice{}.AttrTypes()},
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
		"count":               types.Int64Type,
		"links":               types.MapType{ElemType: types.ObjectType{AttrTypes: RackLink{}.AttrTypes()}},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: Tag{}.AttrTypes()}},
	}
}

func (o *GenericSystem) Request(ctx context.Context, path path.Path, rack *RackType, diags *diag.Diagnostics) *apstra.RackElementGenericSystemRequest {
	var poIdMinVal, poIdMaxVal int
	if !o.PortChannelIdMin.IsNull() {
		poIdMinVal = int(o.PortChannelIdMin.ValueInt64())
	}
	if !o.PortChannelIdMax.IsNull() {
		poIdMaxVal = int(o.PortChannelIdMax.ValueInt64())
	}

	links := o.GetLinks(ctx, diags)
	if diags.HasError() {
		return nil
	}

	linkRequests := make([]apstra.RackLinkRequest, len(links))
	i := 0
	for name, link := range links {
		lr := link.Request(ctx, path.AtName("links").AtMapKey(name), rack, diags)
		if diags.HasError() {
			return nil
		}
		lr.Label = name
		if diags.HasError() {
			return nil
		}

		linkRequests[i] = *lr
		i++
	}

	tagIds := make([]apstra.ObjectId, len(o.TagIds.Elements()))
	o.TagIds.ElementsAs(ctx, &tagIds, false)

	return &apstra.RackElementGenericSystemRequest{
		Count:            int(o.Count.ValueInt64()),
		AsnDomain:        apstra.FeatureSwitchDisabled,
		ManagementLevel:  apstra.SystemManagementLevelUnmanaged,
		PortChannelIdMin: poIdMinVal,
		PortChannelIdMax: poIdMaxVal,
		Loopback:         apstra.FeatureSwitchDisabled,
		Tags:             tagIds,
		Links:            linkRequests,
		LogicalDeviceId:  apstra.ObjectId(o.LogicalDeviceId.ValueString()),
	}
}

func (o *GenericSystem) LoadApiData(ctx context.Context, in *apstra.RackElementGenericSystem, diags *diag.Diagnostics) {
	o.LogicalDeviceId = types.StringNull()
	o.LogicalDevice = NewLogicalDeviceObject(ctx, in.LogicalDevice, diags)
	o.PortChannelIdMin = types.Int64Value(int64(in.PortChannelIdMin))
	o.PortChannelIdMax = types.Int64Value(int64(in.PortChannelIdMax))
	o.Count = types.Int64Value(int64(in.Count))
	o.Links = NewLinkMap(ctx, in.Links, diags)
	o.TagIds = types.SetNull(types.StringType)
	o.Tags = NewTagSet(ctx, in.Tags, diags)
}

func (o *GenericSystem) GetLinks(ctx context.Context, diags *diag.Diagnostics) map[string]RackLink {
	links := make(map[string]RackLink, len(o.Links.Elements()))
	d := o.Links.ElementsAs(ctx, &links, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the link name from the map key into the object's Name field
	for name, link := range links {
		links[name] = link
	}
	return links
}

func (o *GenericSystem) CopyWriteOnlyElements(ctx context.Context, src *GenericSystem, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "GenericSystem.CopyWriteOnlyElements: attempt to copy from nil source")
		return
	}

	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = utils.SetValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)

	var d diag.Diagnostics

	srcLinks := make(map[string]RackLink, len(src.Links.Elements()))
	d = src.Links.ElementsAs(ctx, &srcLinks, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	dstLinks := make(map[string]RackLink, len(o.Links.Elements()))
	d = o.Links.ElementsAs(ctx, &dstLinks, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for name, dstLink := range dstLinks {
		if srcLink, ok := srcLinks[name]; ok {
			dstLink.CopyWriteOnlyElements(ctx, &srcLink, diags)
			dstLinks[name] = dstLink
		}
	}

	o.Links = utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: RackLink{}.AttrTypes()}, dstLinks, diags)
	if diags.HasError() {
		return
	}

}

func NewGenericSystemMap(ctx context.Context, in []apstra.RackElementGenericSystem, diags *diag.Diagnostics) types.Map {
	genericSystems := make(map[string]GenericSystem, len(in))
	for _, genericIn := range in {
		var gs GenericSystem
		gs.LoadApiData(ctx, &genericIn, diags)
		genericSystems[genericIn.Label] = gs
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: GenericSystem{}.AttrTypes()})
		}
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: GenericSystem{}.AttrTypes()}, genericSystems, diags)
}
