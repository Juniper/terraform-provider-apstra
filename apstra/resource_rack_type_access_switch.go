package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rRackTypeAccessSwitch struct {
	Name               types.String `tfsdk:"name"`
	LogicalDeviceId    types.String `tfsdk:"logical_device_id"`
	Count              types.Int64  `tfsdk:"count"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	Links              types.Map    `tfsdk:"links"`
	//LogicalDevice      types.Object `tfsdk:"logical_device"`
	//TagIds             types.Set    `tfsdk:"tag_ids"`
	//TagData            types.Set    `tfsdk:"tag_data"`
	//EsiLagInfo         types.Object `tfsdk:"esi_lag_info""`
}

func (o rRackTypeAccessSwitch) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Switch name, copied from map key, used when creating intra-rack links targeting this switch.",
			Computed:            true,
		},
		"logical_device_id": schema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"count": schema.Int64Attribute{
			MarkdownDescription: "Number of Access Switches of this type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"redundancy_protocol": schema.StringAttribute{
			MarkdownDescription: "Indicates whether the switch is a redundant pair.",
			Computed:            true,
		},
		"links": schema.MapNestedAttribute{
			MarkdownDescription: "Each Access Switch is required to have at least one Link to a Leaf Switch.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: schema.NestedAttributeObject{
				Attributes: rRackLink{}.attributes(),
			},
		},
		//"logical_device": logicalDeviceDataAttributeSchema(),
		//"tag_ids":        tagIdsAttributeSchema(),
		//"tag_data":       tagsDataAttributeSchema(),
		//"esi_lag_info": {
		//	MarkdownDescription: "Including this stanza converts the Access Switch into a redundant pair.",
		//	Optional:            true,
		//	Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		//		"l3_peer_link_count": {
		//			MarkdownDescription: "Number of L3 links between ESI-LAG devices.",
		//			Required:            true,
		//			Type:                types.Int64Type,
		//			Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
		//		},
		//		"l3_peer_link_speed": {
		//			MarkdownDescription: "Speed of l3 links between ESI-LAG devices.",
		//			Required:            true,
		//			Type:                types.StringType,
		//		},
		//	}),
		//},
	}
}

func (o rRackTypeAccessSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"logical_device_id":   types.StringType,
		"count":               types.Int64Type,
		"redundancy_protocol": types.StringType,
		"links":               types.MapType{ElemType: rRackLink{}.attrType()},
		//"logical_device":      logicalDeviceData{}.attrType(),
		//"tag_ids":             types.SetType{ElemType: types.StringType},
		//"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
		//"esi_lag_info":        esiLagInfo{}.attrType(),
	}
}

func (o rRackTypeAccessSwitch) attrType() attr.Type {
	return types.ObjectType{AttrTypes: o.attrTypes()}
}

func (o *rRackTypeAccessSwitch) copyWriteOnlyElements(src *rRackTypeAccessSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "rRackTypeAccessSwitch.copyWriteOnlyElements: attempt to copy from nil source")
		return
	}

	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	//o.TagIds = types.SetValueMust(types.StringType, src.TagIds.Elements())

	//for i, link := range o.Links {
	//	srcLink := src.linkByName(link.Name)
	//	if srcLink == nil {
	//		continue
	//	}
	//	o.Links[i].copyWriteOnlyElements(srcLink, diags)
	//	if diags.HasError() {
	//		return
	//	}
	//}
}

func (o *rRackTypeAccessSwitch) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackElementAccessSwitchRequest {
	redundancyProtocol := goapstra.AccessRedundancyProtocolNone
	if !o.RedundancyProtocol.IsNull() {
		err := redundancyProtocol.FromString(o.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path.AtMapKey("redundancy_protocol"),
				"error parsing redundancy_protocol", err.Error())
			return nil
		}
	}

	lacpActive := goapstra.RackLinkLagModeActive.String()

	links := make([]rRackLink, len(o.Links.Elements()))
	d := o.Links.ElementsAs(ctx, &links, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	linkRequests := make([]goapstra.RackLinkRequest, len(links))
	for i, link := range links {
		link.LagMode = types.StringValue(lacpActive)

		setVal, d := types.ObjectValueFrom(ctx, link.attrTypes(), &link)
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}

		linkReq := link.request(ctx, path.AtSetValue(setVal), rack, diags)
		if diags.HasError() {
			return nil
		}

		linkRequests[i] = *linkReq
	}

	//var tagIds []goapstra.ObjectId
	//if o.TagIds != nil {
	//	tagIds = make([]goapstra.ObjectId, len(o.TagIds))
	//	for i, tagId := range o.TagIds {
	//		tagIds[i] = goapstra.ObjectId(tagId)
	//	}
	//}

	//var esiLagInfo *goapstra.EsiLagInfo
	//if o.EsiLagInfo != nil {
	//	esiLagInfo.AccessAccessLinkCount = int(o.EsiLagInfo.L3PeerLinkCount)
	//	esiLagInfo.AccessAccessLinkSpeed = goapstra.LogicalDevicePortSpeed(o.EsiLagInfo.L3PeerLinkSpeed)
	//}

	return &goapstra.RackElementAccessSwitchRequest{
		Label:              o.Name.ValueString(),
		InstanceCount:      int(o.Count.ValueInt64()),
		RedundancyProtocol: redundancyProtocol,
		Links:              linkRequests,
		LogicalDeviceId:    goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
		//Tags:               tagIds,
		//EsiLagInfo:         esiLagInfo,
	}
}

func (o *rRackTypeAccessSwitch) validateConfig(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) {
	arp := goapstra.AccessRedundancyProtocolNone
	if !o.RedundancyProtocol.IsNull() {
		err := arp.FromString(o.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing redundancy protocol", err.Error())
		}
	}

	//if len(o.TagIds) != 0 {
	//	diags.AddAttributeError(path.AtName("tag_ids"), errInvalidConfig, "tag_ids not currently supported")
	//}

	//for i, link := range o.Links {
	//	link.validateConfigForAccessSwitch(ctx, arp, rack, path.AtListIndex(i), diags) // todo: Need AtSetValue() here
	//}
}

func (o *rRackTypeAccessSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Count = types.Int64Value(int64(in.InstanceCount))
	o.RedundancyProtocol = types.StringNull()
	if in.RedundancyProtocol != goapstra.AccessRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	//if in.EsiLagInfo != nil {
	//	o.EsiLagInfo = &esiLagInfo{}
	//	o.EsiLagInfo.parseApi(in.EsiLagInfo)
	//}
	//o.LogicalDevice.parseApi(in.LogicalDevice)

	//if len(in.Tags) > 0 {
	//	o.TagData = make([]tagData, len(in.Tags)) // populated below
	//	for i := range in.Tags {
	//		o.TagData[i].parseApi(&in.Tags[i])
	//	}
	//}

	links := newResourceLinkMap(ctx, in.Links, diags)
	if diags.HasError() {
		return
	}

	o.Links = links
}

//func (o *rRackTypeAccessSwitch) linkByName(desired string) *dRackLink {
//	for _, link := range o.Links {
//		if link.Name == desired {
//			return &link
//		}
//	}
//	return nil
//}
