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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type rRackTypeAccessSwitch struct {
	Count              types.Int64  `tfsdk:"count"`
	EsiLagInfo         types.Object `tfsdk:"esi_lag_info"`
	Links              types.Map    `tfsdk:"links"`
	LogicalDeviceData  types.Object `tfsdk:"logical_device"`
	LogicalDeviceId    types.String `tfsdk:"logical_device_id"`
	Name               types.String `tfsdk:"name"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	TagIds             types.Set    `tfsdk:"tag_ids"`
	TagData            types.Set    `tfsdk:"tag_data"`
}

func (o rRackTypeAccessSwitch) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Access Switch name, copied from map key, used when creating intra-rack links targeting this switch.",
			Computed:            true,
		},
		"count": schema.Int64Attribute{
			MarkdownDescription: "Number of Access Switches of this type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"logical_device_id": schema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Access Switch.",
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
			MarkdownDescription: "Each Access Switch is required to have at least one Link to a Leaf Switch.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: schema.NestedAttributeObject{
				Attributes: rRackLink{}.attributes(),
			},
		},
		"tag_ids": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Access Switch",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Access Switch",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.resourceAttributes(),
			},
		},
		"redundancy_protocol": schema.StringAttribute{
			MarkdownDescription: "Indicates whether the switch is a redundant pair.",
			Computed:            true,
		},
		"esi_lag_info": schema.SingleNestedAttribute{
			MarkdownDescription: "Including this stanza converts the Access Switch into a redundant pair.",
			Optional:            true,
			Attributes:          esiLagInfo{}.schemaAsResource(),
		},
	}
}

func (o rRackTypeAccessSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"logical_device_id":   types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: logicalDeviceData{}.attrTypes()},
		"links":               types.MapType{ElemType: types.ObjectType{AttrTypes: rRackLink{}.attrTypes()}},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tag_data":            types.SetType{ElemType: types.ObjectType{AttrTypes: tagData{}.attrTypes()}},
		"redundancy_protocol": types.StringType,
		"esi_lag_info":        types.ObjectType{AttrTypes: esiLagInfo{}.attrTypes()},
	}
}

func (o *rRackTypeAccessSwitch) copyWriteOnlyElements(ctx context.Context, src *rRackTypeAccessSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "rRackTypeAccessSwitch.copyWriteOnlyElements: attempt to copy from nil source")
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

	o.Links = mapValueOrNull(ctx, types.ObjectType{AttrTypes: rRackLink{}.attrTypes()}, dstLinks, diags)
	if diags.HasError() {
		return
	}
}

func (o *rRackTypeAccessSwitch) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackElementAccessSwitchRequest {
	redundancyProtocol := goapstra.AccessRedundancyProtocolNone
	if !o.EsiLagInfo.IsNull() {
		redundancyProtocol = goapstra.AccessRedundancyProtocolEsi
	}

	lacpActive := goapstra.RackLinkLagModeActive.String()

	links := o.links(ctx, diags)
	if diags.HasError() {
		return nil
	}

	linkRequests := make([]goapstra.RackLinkRequest, len(links))
	i := 0
	for name, link := range links {
		link.LagMode = types.StringValue(lacpActive)
		lr := link.request(ctx, path.AtName("links").AtMapKey(name), rack, diags)
		if diags.HasError() {
			return nil
		}

		linkRequests[i] = *lr
		i++
	}

	var tagIds []goapstra.ObjectId
	tagIds = make([]goapstra.ObjectId, len(o.TagIds.Elements()))
	o.TagIds.ElementsAs(ctx, &tagIds, false)

	var eli esiLagInfo
	diags.Append(o.EsiLagInfo.As(ctx, &eli, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &goapstra.RackElementAccessSwitchRequest{
		Label:              o.Name.ValueString(),
		InstanceCount:      int(o.Count.ValueInt64()),
		RedundancyProtocol: redundancyProtocol,
		Links:              linkRequests,
		LogicalDeviceId:    goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
		Tags:               tagIds,
		EsiLagInfo:         eli.request(ctx, diags),
	}
}

func (o *rRackTypeAccessSwitch) links(ctx context.Context, diags *diag.Diagnostics) map[string]rRackLink {
	links := make(map[string]rRackLink, len(o.Links.Elements()))
	d := o.Links.ElementsAs(ctx, &links, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the link name from the map key into the object's Name field
	for name, link := range links {
		link.Name = types.StringValue(name)
		links[name] = link
	}
	return links
}

func (o *rRackTypeAccessSwitch) linkByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackLink {
	links := o.links(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if link, ok := links[requested]; ok {
		return &link
	}

	return nil
}

func (o *rRackTypeAccessSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Count = types.Int64Value(int64(in.InstanceCount))
	o.RedundancyProtocol = types.StringNull()
	if in.RedundancyProtocol != goapstra.AccessRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	o.LogicalDeviceData = newLogicalDeviceDataObject(ctx, in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}

	o.EsiLagInfo = newEsiLagInfo(ctx, in.EsiLagInfo, diags)
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
