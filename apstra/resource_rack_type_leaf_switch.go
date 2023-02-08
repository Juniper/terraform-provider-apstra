package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"strings"
)

type rRackTypeLeafSwitch struct {
	LogicalDeviceData types.Object `tfsdk:"logical_device"`
	LogicalDeviceId   types.String `tfsdk:"logical_device_id"`
	//MlagInfo           types.Object `tfsdk:"mlag_info""` // todo re-enable
	Name               types.String `tfsdk:"name"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	//TagIds             types.Set    `tfsdk:"tag_ids"` // todo re-enable
	//TagData            types.Set    `tfsdk:"tag_data"` // todo re-enable
}

func (o rRackTypeLeafSwitch) schema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device_id": schema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"spine_link_count": schema.Int64Attribute{
			MarkdownDescription: "Links per spine.",
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"spine_link_speed": schema.StringAttribute{
			MarkdownDescription: "Speed of spine-facing links, something like '10G'",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"redundancy_protocol": schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Enabling a redundancy protocol converts a single "+
				"Leaf Switch into a LAG-capable switch pair. Must be one of '%s'.",
				strings.Join(leafRedundancyModes(), "', '")),
			Optional:   true,
			Validators: []validator.String{stringvalidator.OneOf(leafRedundancyModes()...)},
		},
		"logical_device": schema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:          logicalDeviceData{}.schemaAsResourceReadOnly(),
		},
		//"tag_ids":        tagIdsAttributeSchema(), // todo re-enable
		//"tag_data":       tagsDataAttributeSchema(), // todo re-enable
		//"mlag_info": mlagInfo{}.schemaAsResource(), // todo re-enable
	}
}

func (o rRackTypeLeafSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"logical_device_id":   types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"logical_device":      logicalDeviceData{}.attrType(),
		//"tag_ids":             types.SetType{ElemType: types.StringType}, // todo re-enable
		//"tag_data":            types.SetType{ElemType: tagData{}.attrType()}, // todo re-enable
		//"mlag_info":           mlagInfo{}.attrType(), // todo re-enable
	}
}

func (o rRackTypeLeafSwitch) attrType() attr.Type {
	return types.ObjectType{AttrTypes: o.attrTypes()}
}

func (o *rRackTypeLeafSwitch) validateConfig(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) {
	fcd := rack.fabricConnectivityDesign(ctx, diags)
	if diags.HasError() {
		return
	}

	//if len(o.TagIds) != 0 {
	//	diags.AddAttributeError(path.AtName("tag_ids"), errInvalidConfig, "tag_ids not currently supported")
	//}

	switch fcd {
	case goapstra.FabricConnectivityDesignL3Clos:
		o.validateForL3Clos(ctx, path, diags) // todo: figure out how to use AtSetValue()
	case goapstra.FabricConnectivityDesignL3Collapsed:
		o.validateForL3Collapsed(ctx, path, diags)
	default:
		diags.AddAttributeError(path, errProviderBug, fmt.Sprintf("unknown fabric connectivity design '%s'", fcd.String()))
	}

	// todo re-enable this
	//if !o.RedundancyProtocol.IsNull() && o.RedundancyProtocol.ValueString() == goapstra.LeafRedundancyProtocolMlag.String() {
	//	o.validateMlagInfo(path, diags)
	//}
}

func (o *rRackTypeLeafSwitch) validateForL3Clos(ctx context.Context, path path.Path, diags *diag.Diagnostics) {
	if o.SpineLinkSpeed.IsNull() {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("'spine_link_speed' must be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Clos))
	}
}

func (o *rRackTypeLeafSwitch) validateForL3Collapsed(ctx context.Context, path path.Path, diags *diag.Diagnostics) {
	if !o.SpineLinkCount.IsNull() {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("'spine_link_count' must not be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Collapsed))
	}

	if !o.SpineLinkSpeed.IsNull() {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("'spine_link_speed' must bnot e specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Collapsed))
	}

	if !o.RedundancyProtocol.IsNull() {
		var redundancyProtocol goapstra.LeafRedundancyProtocol
		err := redundancyProtocol.FromString(o.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path.AtMapKey("redundancy_protocol"), "parse_error", err.Error())
			return
		}
		if redundancyProtocol == goapstra.LeafRedundancyProtocolMlag {
			diags.AddAttributeError(path, errInvalidConfig,
				fmt.Sprintf("'redundancy_protocol' = '%s' is not allowed when 'fabric_connectivity_design' = '%s'",
					goapstra.LeafRedundancyProtocolMlag, goapstra.FabricConnectivityDesignL3Collapsed))
		}
	}
}

//func (o *rRackTypeLeafSwitch) validateMlagInfo(path path.Path, diags *diag.Diagnostics) {
//	var redundancyProtocol goapstra.LeafRedundancyProtocol
//	err := redundancyProtocol.FromString(o.RedundancyProtocol.ValueString())
//	if err != nil {
//		diags.AddAttributeError(path.AtMapKey("redundancy_protocol"), "parse_error", err.Error())
//		return
//	}
//
//	if o.MlagInfo == nil && redundancyProtocol == goapstra.LeafRedundancyProtocolMlag {
//		diags.AddAttributeError(path, errInvalidConfig,
//			fmt.Sprintf("'mlag_info' required with 'redundancy_protocol' = '%s'", redundancyProtocol.String()))
//	}
//
//	if o.MlagInfo == nil {
//		return
//	}
//
//	if redundancyProtocol != goapstra.LeafRedundancyProtocolMlag {
//		diags.AddAttributeError(path, errInvalidConfig,
//			fmt.Sprintf("'mlag_info' incompatible with 'redundancy_protocol of '%s'", redundancyProtocol.String()))
//	}
//
//	if o.MlagInfo.PeerLinkPortChannelId != nil &&
//		o.MlagInfo.L3PeerLinkPortChannelId != nil &&
//		*o.MlagInfo.PeerLinkPortChannelId == *o.MlagInfo.L3PeerLinkPortChannelId {
//		diags.AddAttributeError(path, errInvalidConfig,
//			fmt.Sprintf("'peer_link_port_channel_id' and 'l3_peer_link_port_channel_id' cannot both use value %d",
//				*o.MlagInfo.PeerLinkPortChannelId))
//	}
//
//	if o.MlagInfo.L3PeerLinkCount != nil && o.MlagInfo.L3PeerLinkSpeed == nil {
//		diags.AddAttributeError(path, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_speed'")
//	}
//	if o.MlagInfo.L3PeerLinkSpeed != nil && o.MlagInfo.L3PeerLinkCount == nil {
//		diags.AddAttributeError(path, errInvalidConfig, "'l3_peer_link_speed' requires 'l3_peer_link_count'")
//	}
//
//	if o.MlagInfo.L3PeerLinkPortChannelId != nil && o.MlagInfo.L3PeerLinkCount == nil {
//		diags.AddAttributeError(path, errInvalidConfig, "'l3_peer_link_port_channel_id' requires 'l3_peer_link_count'")
//	}
//	if o.MlagInfo.L3PeerLinkCount != nil && o.MlagInfo.L3PeerLinkPortChannelId == nil {
//		diags.AddAttributeError(path, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_port_channel_id'")
//	}
//}

func (o *rRackTypeLeafSwitch) copyWriteOnlyElements(ctx context.Context, src *rRackTypeLeafSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "rRackTypeLeafSwitch.copyWriteOnlyElements: attempt to copy from nil source")
		return
	}
	// todo: these might not be safe, consider extracting / recreating the data
	o.LogicalDeviceId = src.LogicalDeviceId
	//o.TagIds = src.TagIds // todo re-enable
}

func (o *rRackTypeLeafSwitch) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackElementLeafSwitchRequest {
	fcd := rack.fabricConnectivityDesign(ctx, diags)

	var linkPerSpineCount int
	if o.SpineLinkCount.IsUnknown() && fcd == goapstra.FabricConnectivityDesignL3Clos {
		// config omits 'spine_link_count' set default value (1) for fabric designs which require it
		linkPerSpineCount = 1
	} else {
		// config includes 'spine_link_count' -- use the configured value
		linkPerSpineCount = int(o.SpineLinkCount.ValueInt64())
	}

	var linkPerSpineSpeed goapstra.LogicalDevicePortSpeed
	if !o.SpineLinkSpeed.IsNull() {
		linkPerSpineSpeed = goapstra.LogicalDevicePortSpeed(o.SpineLinkSpeed.ValueString())
	}

	redundancyProtocol := goapstra.LeafRedundancyProtocolNone
	if !o.RedundancyProtocol.IsNull() {
		err := redundancyProtocol.FromString(o.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path.AtMapKey("redundancy_protocol"),
				fmt.Sprintf("error parsing redundancy_protocol '%s'",
					o.RedundancyProtocol.ValueString()),
				err.Error())
			return nil
		}
	}

	// todo fix tags
	//var tagIds []goapstra.ObjectId
	//if o.TagIds != nil {
	//	tagIds = make([]goapstra.ObjectId, len(o.TagIds))
	//	for i, tagId := range o.TagIds {
	//		tagIds[i] = goapstra.ObjectId(tagId)
	//	}
	//}

	return &goapstra.RackElementLeafSwitchRequest{
		Label: o.Name.ValueString(),
		//MlagInfo:           o.MlagInfo.request(),
		LinkPerSpineCount:  linkPerSpineCount,
		LinkPerSpineSpeed:  linkPerSpineSpeed,
		RedundancyProtocol: redundancyProtocol,
		//Tags:               tagIds,
		LogicalDeviceId: goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
	}
}

func (o *rRackTypeLeafSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	if fcd != goapstra.FabricConnectivityDesignL3Collapsed {
		o.SpineLinkCount = types.Int64Value(int64(in.LinkPerSpineCount))
		o.SpineLinkSpeed = types.StringValue(string(in.LinkPerSpineSpeed))
	}

	if in.RedundancyProtocol == goapstra.LeafRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringNull()
	} else {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	//o.MlagInfo = newMlagInfoObject(ctx, in.MlagInfo, diags)
	//if diags.HasError() {
	//	return
	//}
	//
	//o.TagData = newTagSet(ctx, in.Tags, diags)
	//if diags.HasError() {
	//	return
	//}

	o.LogicalDeviceData = newLogicalDeviceDataObject(ctx, in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}

	//if len(in.Tags) > 0 {
	//	o.TagData = make([]tagData, len(in.Tags)) // populated below
	//	for i := range in.Tags {
	//		o.TagData[i].parseApi(&in.Tags[i])
	//	}
	//}
}
