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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func validateAccessSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	as := rt.Data.AccessSwitches[i]
	if as.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi && as.EsiLagInfo == nil {
		diags.AddError("access switch ESI LAG Info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, as.Label, as.RedundancyProtocol.String()))
		return
	}
	if as.LogicalDevice == nil {
		diags.AddError("access switch logical device info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' logical device is nil",
				rt.Id, as.Label))
		return
	}
}

type accessSwitch struct {
	LogicalDeviceId    types.String `tfsdk:"logical_device_id"`
	LogicalDevice      types.Object `tfsdk:"logical_device"`
	EsiLagInfo         types.Object `tfsdk:"esi_lag_info"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	Count              types.Int64  `tfsdk:"count"`
	Links              types.Map    `tfsdk:"links"`
	TagIds             types.Set    `tfsdk:"tag_ids"`
	Tags               types.Set    `tfsdk:"tags"`
}

func (o accessSwitch) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
		"esi_lag_info": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Interconnect information for Access Switches in ESI-LAG redundancy mode.",
			Computed:            true,
			Attributes:          esiLagInfo{}.dataSourceAttributes(),
		},
		"redundancy_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
			Computed:            true,
		},
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of Access Switches of this type.",
			Computed:            true,
		},
		"links": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Details links from this Access Switch to upstream switches within this Rack Type.",
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
			MarkdownDescription: "Details any tags applied to this Access Switch.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o accessSwitch) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Access Switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:          logicalDevice{}.resourceAttributesNested(),
		},
		"esi_lag_info": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Including this stanza converts the Access Switch into a redundant pair.",
			Optional:            true,
			Attributes:          esiLagInfo{}.schemaAsResource(),
		},
		"redundancy_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: "Indicates whether the switch is a redundant pair.",
			Computed:            true,
		},
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Access Switches of this type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"links": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Each Access Switch is required to have at least one Link to a Leaf Switch.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: rackLink{}.resourceAttributes(),
			},
		},
		"tag_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag IDs to be applied to this Access Switch",
			Optional:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			ElementType:         types.StringType,
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Access Switch",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: tag{}.resourceAttributesNested(),
			},
		},
	}
}

func (o accessSwitch) resourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          logicalDevice{}.resourceAttributesNested(),
		},
		"esi_lag_info": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Defines connectivity between ESI LAG peers when "+
				"`redundancy_protocol` is set to `%s`.", goapstra.AccessRedundancyProtocolEsi.String()),
			Computed:   true,
			Attributes: esiLagInfo{}.schemaAsResource(),
		},
		"redundancy_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: "Indicates whether the switch is a redundant pair.",
			Computed:            true,
		},
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Access Switches of this type.",
			Computed:            true,
		},
		"links": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Each Access Switch is required to have at least one Link to a Leaf Switch.",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: rackLink{}.resourceAttributesNested(),
			},
		},
		"tag_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in nested contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Access Switch",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: tag{}.resourceAttributesNested(),
			},
		},
	}
}

func (o accessSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":   types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"esi_lag_info":        types.ObjectType{AttrTypes: esiLagInfo{}.attrTypes()},
		"redundancy_protocol": types.StringType,
		"count":               types.Int64Type,
		"links":               types.MapType{ElemType: types.ObjectType{AttrTypes: rackLink{}.attrTypes()}},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *accessSwitch) request(ctx context.Context, path path.Path, rack *rackType, diags *diag.Diagnostics) *goapstra.RackElementAccessSwitchRequest {
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

	var eli esiLagInfo
	diags.Append(o.EsiLagInfo.As(ctx, &eli, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &goapstra.RackElementAccessSwitchRequest{
		InstanceCount:      int(o.Count.ValueInt64()),
		RedundancyProtocol: redundancyProtocol,
		Links:              linkRequests,
		LogicalDeviceId:    goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
		Tags:               tagIds,
		EsiLagInfo:         eli.request(ctx, diags),
	}
}

func (o *accessSwitch) loadApiData(ctx context.Context, in *goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.LogicalDeviceId = types.StringNull()
	o.LogicalDevice = newLogicalDeviceObject(ctx, in.LogicalDevice, diags)
	o.EsiLagInfo = newEsiLagInfo(ctx, in.EsiLagInfo, diags)
	o.RedundancyProtocol = stringValueWithNull(ctx, in.RedundancyProtocol.String(), goapstra.AccessRedundancyProtocolNone.String(), diags)
	o.Count = types.Int64Value(int64(in.InstanceCount))
	o.Links = newLinkMap(ctx, in.Links, diags)
	o.TagIds = types.SetNull(types.StringType)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

func (o *accessSwitch) links(ctx context.Context, diags *diag.Diagnostics) map[string]rackLink {
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

func (o *accessSwitch) copyWriteOnlyElements(ctx context.Context, src *accessSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "accessSwitch.copyWriteOnlyElements: attempt to copy from nil source")
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

func newAccessSwitchMap(ctx context.Context, in []goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) types.Map {
	accessSwitches := make(map[string]accessSwitch, len(in))
	for _, accessIn := range in {
		var as accessSwitch
		as.loadApiData(ctx, &accessIn, diags)
		accessSwitches[accessIn.Label] = as
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: accessSwitch{}.attrTypes()})
		}
	}

	return mapValueOrNull(ctx, types.ObjectType{AttrTypes: accessSwitch{}.attrTypes()}, accessSwitches, diags)
}
