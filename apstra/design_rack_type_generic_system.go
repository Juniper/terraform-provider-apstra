package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateGenericSystem(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	gs := rt.Data.GenericSystems[i]
	if gs.LogicalDevice == nil {
		diags.AddError("generic system logical device info missing",
			fmt.Sprintf("rack type '%s', generic system '%s' logical device is nil",
				rt.Id, gs.Label))
		return
	}
}

type genericSystem struct {
	LogicalDeviceId  types.String `tfsdk:"logical_device_id"`
	LogicalDevice    types.Object `tfsdk:"logical_device"`
	PortChannelIdMin types.Int64  `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64  `tfsdk:"port_channel_id_max"`
	Count            types.Int64  `tfsdk:"count"`
	Links            types.Map    `tfsdk:"links"`
	TagIds           types.Set    `tfsdk:"tag_ids"`
	Tags             types.Set    `tfsdk:"tags"`
}

func (o genericSystem) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
				Attributes: rackLink{}.dataSourceAttributes(),
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
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o genericSystem) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Generic System.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:          logicalDevice{}.resourceAttributesNested(),
		},
		"port_channel_id_min": resourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.Between(poIdMin, poIdMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("port_channel_id_max")),
				int64validator.AtMostSumOf(path.MatchRelative().AtParent().AtName("port_channel_id_max")),
			},
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"port_channel_id_max": resourceSchema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.Between(poIdMin, poIdMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("port_channel_id_min")),
				int64validator.AtLeastSumOf(path.MatchRelative().AtParent().AtName("port_channel_id_min")),
			},
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
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
				Attributes: rackLink{}.resourceAttributes(),
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
				Attributes: tag{}.resourceAttributesNested(),
			},
		},
	}
}

func (o genericSystem) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":   types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
		"count":               types.Int64Type,
		"links":               types.MapType{ElemType: types.ObjectType{AttrTypes: rackLink{}.attrTypes()}},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *genericSystem) request(ctx context.Context, path path.Path, rack *rackType, diags *diag.Diagnostics) *goapstra.RackElementGenericSystemRequest {
	var poIdMinVal, poIdMaxVal int
	if !o.PortChannelIdMin.IsNull() {
		poIdMinVal = int(o.PortChannelIdMin.ValueInt64())
	}
	if !o.PortChannelIdMax.IsNull() {
		poIdMaxVal = int(o.PortChannelIdMax.ValueInt64())
	}

	links := o.links(ctx, diags)
	if diags.HasError() {
		return nil
	}

	linkRequests := make([]goapstra.RackLinkRequest, len(links))
	i := 0
	for name, link := range links {
		lr := link.request(ctx, path.AtName("links").AtMapKey(name), rack, diags)
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

	var tagIds []goapstra.ObjectId
	tagIds = make([]goapstra.ObjectId, len(o.TagIds.Elements()))
	o.TagIds.ElementsAs(ctx, &tagIds, false)

	return &goapstra.RackElementGenericSystemRequest{
		Count:            int(o.Count.ValueInt64()),
		AsnDomain:        goapstra.FeatureSwitchDisabled,
		ManagementLevel:  goapstra.GenericSystemUnmanaged,
		PortChannelIdMin: poIdMinVal,
		PortChannelIdMax: poIdMaxVal,
		Loopback:         goapstra.FeatureSwitchDisabled,
		Tags:             tagIds,
		Links:            linkRequests,
		LogicalDeviceId:  goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
	}
}

func (o *genericSystem) loadApiData(ctx context.Context, in *goapstra.RackElementGenericSystem, diags *diag.Diagnostics) {
	o.LogicalDeviceId = types.StringNull()
	o.LogicalDevice = newLogicalDeviceObject(ctx, in.LogicalDevice, diags)
	o.PortChannelIdMin = types.Int64Value(int64(in.PortChannelIdMin))
	o.PortChannelIdMax = types.Int64Value(int64(in.PortChannelIdMax))
	o.Count = types.Int64Value(int64(in.Count))
	o.Links = newLinkMap(ctx, in.Links, diags)
	o.TagIds = types.SetNull(types.StringType)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

func (o *genericSystem) links(ctx context.Context, diags *diag.Diagnostics) map[string]rackLink {
	links := make(map[string]rackLink, len(o.Links.Elements()))
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

func (o *genericSystem) copyWriteOnlyElements(ctx context.Context, src *genericSystem, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "genericSystem.copyWriteOnlyElements: attempt to copy from nil source")
		return
	}

	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = setValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)

	var d diag.Diagnostics

	srcLinks := make(map[string]rackLink, len(src.Links.Elements()))
	d = src.Links.ElementsAs(ctx, &srcLinks, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	dstLinks := make(map[string]rackLink, len(o.Links.Elements()))
	d = o.Links.ElementsAs(ctx, &dstLinks, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for name, dstLink := range dstLinks {
		if srcLink, ok := srcLinks[name]; ok {
			dstLink.copyWriteOnlyElements(ctx, &srcLink, diags)
			dstLinks[name] = dstLink
		}
	}

	o.Links = mapValueOrNull(ctx, types.ObjectType{AttrTypes: rackLink{}.attrTypes()}, dstLinks, diags)
	if diags.HasError() {
		return
	}

}

func newGenericSystemMap(ctx context.Context, in []goapstra.RackElementGenericSystem, diags *diag.Diagnostics) types.Map {
	genericSystems := make(map[string]genericSystem, len(in))
	for _, genericIn := range in {
		var gs genericSystem
		gs.loadApiData(ctx, &genericIn, diags)
		genericSystems[genericIn.Label] = gs
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: genericSystem{}.attrTypes()})
		}
	}

	return mapValueOrNull(ctx, types.ObjectType{AttrTypes: genericSystem{}.attrTypes()}, genericSystems, diags)
}
