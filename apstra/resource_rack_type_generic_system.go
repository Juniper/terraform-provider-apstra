package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rRackTypeGenericSystem struct {
	Name              types.String `tfsdk:"name"`
	Count             types.Int64  `tfsdk:"count"`
	PortChannelIdMin  types.Int64  `tfsdk:"port_channel_id_min"`
	PortChannelIdMax  types.Int64  `tfsdk:"port_channel_id_max"`
	LogicalDeviceId   types.String `tfsdk:"logical_device_id"`
	LogicalDeviceData types.Object `tfsdk:"logical_device"`
	Links             types.Map    `tfsdk:"links"`
	TagIds            types.Set    `tfsdk:"tag_ids"`
	TagData           types.Set    `tfsdk:"tag_data"`
}

func (o rRackTypeGenericSystem) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Generic System name, copied from map key.",
			Computed:            true,
		},
		"count": schema.Int64Attribute{
			MarkdownDescription: "Number of Generic Systems of this type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"logical_device_id": schema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Generic System.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device": schema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:          logicalDeviceData{}.schemaAsResourceReadOnly(),
		},
		"links": schema.MapNestedAttribute{
			MarkdownDescription: "Each Generic System is required to have at least one Link to a Leaf Switch or Access Switch.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: schema.NestedAttributeObject{
				Attributes: rRackLink{}.attributes(),
			},
		},
		"tag_ids": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Generic System",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Generic System",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.resourceAttributes(),
			},
		},
		"port_channel_id_min": schema.Int64Attribute{
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
		"port_channel_id_max": schema.Int64Attribute{
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
	}
}

func (o rRackTypeGenericSystem) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"logical_device_id":   types.StringType,
		"logical_device":      logicalDeviceData{}.attrType(),
		"links":               types.MapType{ElemType: rRackLink{}.attrType()},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
	}
}

func (o rRackTypeGenericSystem) attrType() attr.Type {
	return types.ObjectType{AttrTypes: o.attrTypes()}
}

func (o *rRackTypeGenericSystem) copyWriteOnlyElements(ctx context.Context, src *rRackTypeGenericSystem, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "rRackTypeGenericSystem.copyWriteOnlyElements: attempt to copy from nil source")
		return
	}

	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = setValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)

	var d diag.Diagnostics

	srcLinks := make(map[string]rRackLink, len(src.Links.Elements()))
	d = src.Links.ElementsAs(ctx, &srcLinks, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	dstLinks := make(map[string]rRackLink, len(o.Links.Elements()))
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
}

func (o *rRackTypeGenericSystem) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackElementGenericSystemRequest {
	var poIdMinVal, poIdMaxVal int
	if !o.PortChannelIdMin.IsNull() {
		poIdMinVal = int(o.PortChannelIdMin.ValueInt64())
	}
	if !o.PortChannelIdMax.IsNull() {
		poIdMaxVal = int(o.PortChannelIdMax.ValueInt64())
	}

	links := make(map[string]rRackLink, len(o.Links.Elements()))
	d := o.Links.ElementsAs(ctx, &links, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	linkRequests := make([]goapstra.RackLinkRequest, len(links))
	var i int
	for name, link := range links {
		linkReq := link.request(ctx, path.AtName("links").AtMapKey(name), rack, diags)
		if diags.HasError() {
			return nil
		}

		linkRequests[i] = *linkReq
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
		Label:            o.Name.ValueString(),
		Links:            linkRequests,
		LogicalDeviceId:  goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
	}
}

func (o *rRackTypeGenericSystem) loadApiResponse(ctx context.Context, in *goapstra.RackElementGenericSystem, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Count = types.Int64Value(int64(in.Count))
	o.PortChannelIdMin = types.Int64Value(int64(in.PortChannelIdMin))
	o.PortChannelIdMax = types.Int64Value(int64(in.PortChannelIdMax))

	o.LogicalDeviceData = newLogicalDeviceDataObject(ctx, in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}

	// todo: this might not be needed - also find similar instances
	// null set for now to avoid nil pointer dereference error because the API
	// response doesn't contain the tag IDs. See copyWriteOnlyElements() method.
	o.TagIds = types.SetNull(types.StringType)

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}

	o.Links = newResourceLinkMap(ctx, in.Links, diags)
	if diags.HasError() {
		return
	}
}
