package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceRackType{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceRackType{}

type dataSourceRackType struct {
	client *goapstra.Client
}

func (o *dataSourceRackType) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_type"
}

func (o *dataSourceRackType) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataDetail,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *dataSourceRackType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details of a specific Rack Type.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one Rack Type. " +
			"Matching zero Rack Type or more than one Rack Type will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Rack Type id.  Required when the Rack Type name is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Rack Type name displayed in the Apstra web UI.  Required when Rack Type id is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"description": {
				MarkdownDescription: "Rack Type description displayed in the Apstra web UI.",
				Computed:            true,
				Type:                types.StringType,
			},
			"fabric_connectivity_design": {
				MarkdownDescription: "Indicates designs for which this Rack Type is intended.",
				Computed:            true,
				Type:                types.StringType,
			},
			"leaf_switches": {
				MarkdownDescription: "Details of Leaf Switches in this Rack Type.",
				Computed:            true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
						Computed:            true,
						Type:                types.StringType,
					},
					"spine_link_count": {
						MarkdownDescription: "Number of links to each spine switch.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"spine_link_speed": {
						MarkdownDescription: "Speed of links to spine switches.",
						Computed:            true,
						Type:                types.StringType,
					},
					"redundancy_protocol": {
						MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
						Computed:            true,
						Type:                types.StringType,
					},
					"mlag_info": {
						MarkdownDescription: "Details settings when the Leaf Switch is an MLAG-capable pair.",
						Computed:            true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"mlag_keepalive_vlan": {
								MarkdownDescription: "MLAG keepalive VLAN ID.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"peer_link_count": {
								MarkdownDescription: "Number of links between MLAG devices.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"peer_link_speed": {
								MarkdownDescription: "Speed of links between MLAG devices.",
								Computed:            true,
								Type:                types.StringType,
							},
							"peer_link_port_channel_id": {
								MarkdownDescription: "Peer link port-channel ID.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"l3_peer_link_count": {
								MarkdownDescription: "Number of L3 links between MLAG devices.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"l3_peer_link_speed": {
								MarkdownDescription: "Speed of l3 links between MLAG devices.",
								Computed:            true,
								Type:                types.StringType,
							},
							"l3_peer_link_port_channel_id": {
								MarkdownDescription: "L3 peer link port-channel ID.",
								Computed:            true,
								Type:                types.Int64Type,
							},
						}),
					},
					"logical_device": logicalDeviceDataAttributeSchema(),
					"tag_data":       tagsDataAttributeSchema(),
				}),
			},
			"access_switches": {
				MarkdownDescription: "Details of Access Switches in this Rack Type.",
				Computed:            true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
						Computed:            true,
						Type:                types.StringType,
					},
					"count": {
						MarkdownDescription: "Count of Access Switches of this type.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"redundancy_protocol": {
						MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
						Computed:            true,
						Type:                types.StringType,
					},
					"esi_lag_info": {
						MarkdownDescription: "Interconnect information for Access Switches in ESI-LAG redundancy mode.",
						Computed:            true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"l3_peer_link_count": {
								MarkdownDescription: "Count of L3 links to ESI peer.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"l3_peer_link_speed": {
								MarkdownDescription: "Speed of L3 links to ESI peer.",
								Computed:            true,
								Type:                types.StringType,
							},
						}),
					},
					"logical_device": logicalDeviceDataAttributeSchema(),
					"tag_data":       tagsDataAttributeSchema(),
					"links":          dLinksAttributeSchema(),
				}),
			},
			"generic_systems": {
				MarkdownDescription: "Details Generic Systems found in the Rack Type.",
				Computed:            true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Generic name, must be unique within the rack-type.",
						Computed:            true,
						Type:                types.StringType,
					},
					"count": {
						MarkdownDescription: "Number of Generic Systems of this type.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"port_channel_id_min": {
						MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"port_channel_id_max": {
						MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"logical_device": logicalDeviceDataAttributeSchema(),
					"tag_data":       tagsDataAttributeSchema(),
					"links":          dLinksAttributeSchema(),
				}),
			},
		},
	}, nil
}

func (o *dataSourceRackType) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dRackType
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'name' must be specified")
		return
	}
}

func (o *dataSourceRackType) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dRackType
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var rt *goapstra.RackType
	var ace goapstra.ApstraClientErr

	// maybe the config gave us the rack type name?
	if !config.Name.IsNull() { // fetch rack type by name
		rt, err = o.client.GetRackTypeByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with name '%s' does not exist", config.Name.ValueString()))
			return
		}
	}

	// maybe the config gave us the rack type id?
	if !config.Id.IsNull() { // fetch rack type by ID
		rt, err = o.client.GetRackType(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with id '%s' does not exist", config.Id.ValueString()))
			return
		}
	}

	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Rack Type", err.Error())
	}

	// catch problems which would crash the provider
	validateRackType(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newState := &dRackType{}
	newState.parseApi(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//Set state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func validateRackType(rt *goapstra.RackType, diags *diag.Diagnostics) {
	if rt.Data == nil {
		diags.AddError("rack type has no data", fmt.Sprintf("rack type '%s' data object is nil", rt.Id))
		return
	}

	for i := range rt.Data.LeafSwitches {
		validateLeafSwitch(rt, i, diags)
	}

	for i := range rt.Data.AccessSwitches {
		validateAccessSwitch(rt, i, diags)
	}

	for i := range rt.Data.GenericSystems {
		validateGenericSystem(rt, i, diags)
	}
}

func validateLeafSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	ls := rt.Data.LeafSwitches[i]
	if ls.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag && ls.MlagInfo == nil {
		diags.AddError("leaf switch MLAG Info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, ls.Label, ls.RedundancyProtocol.String()))
	}
	if ls.LogicalDevice == nil {
		diags.AddError("leaf switch logical device info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' logical device is nil",
				rt.Id, ls.Label))
	}
}

func validateAccessSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	as := rt.Data.AccessSwitches[i]
	if as.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi && as.EsiLagInfo == nil {
		diags.AddError("access switch ESI LAG Info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, as.Label, as.RedundancyProtocol.String()))
	}
	if as.LogicalDevice == nil {
		diags.AddError("access switch logical device info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' logical device is nil",
				rt.Id, as.Label))
	}
}

func validateGenericSystem(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	gs := rt.Data.GenericSystems[i]
	if gs.LogicalDevice == nil {
		diags.AddError("generic system logical device info missing",
			fmt.Sprintf("rack type '%s', generic system '%s' logical device is nil",
				rt.Id, gs.Label))
	}
}

type dRackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Set    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Set    `tfsdk:"access_switches"`
	GenericSystems           types.Set    `tfsdk:"generic_systems"`
}

func (o *dRackType) parseApi(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	var d diag.Diagnostics

	leafSwitchSet := types.SetNull(dRackTypeLeafSwitch{}.attrType())
	if len(in.Data.LeafSwitches) > 0 {
		leafSwitches := make([]dRackTypeLeafSwitch, len(in.Data.LeafSwitches))
		for i := range in.Data.LeafSwitches {
			leafSwitches[i].parseApi(&in.Data.LeafSwitches[i], in.Data.FabricConnectivityDesign)
		}
		leafSwitchSet, d = types.SetValueFrom(ctx, dRackTypeLeafSwitch{}.attrType(), leafSwitches)
		diags.Append(d...)
	}

	accessSwitchSet := types.SetNull(dRackTypeAccessSwitch{}.attrType())
	if len(in.Data.AccessSwitches) > 0 {
		accessSwitches := make([]dRackTypeAccessSwitch, len(in.Data.AccessSwitches))
		for i := range in.Data.AccessSwitches {
			accessSwitches[i].parseApi(&in.Data.AccessSwitches[i])
		}
		accessSwitchSet, d = types.SetValueFrom(ctx, dRackTypeAccessSwitch{}.attrType(), accessSwitches)
		diags.Append(d...)
	}

	genericSystemSet := types.SetNull(dRackTypeGenericSystem{}.attrType())
	if len(in.Data.GenericSystems) > 0 {
		genericSystems := make([]dRackTypeGenericSystem, len(in.Data.GenericSystems))
		for i := range in.Data.GenericSystems {
			genericSystems[i].parseApi(&in.Data.GenericSystems[i])
		}
		genericSystemSet, d = types.SetValueFrom(ctx, dRackTypeGenericSystem{}.attrType(), genericSystems)
		diags.Append(d...)
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Description = types.StringValue(in.Data.Description)
	o.FabricConnectivityDesign = types.StringValue(in.Data.FabricConnectivityDesign.String())
	o.LeafSwitches = leafSwitchSet
	o.AccessSwitches = accessSwitchSet
	o.GenericSystems = genericSystemSet
}

type dRackTypeLeafSwitch struct {
	Name               string            `tfsdk:"name"`
	SpineLinkCount     *int64            `tfsdk:"spine_link_count"`
	SpineLinkSpeed     *string           `tfsdk:"spine_link_speed"`
	RedundancyProtocol *string           `tfsdk:"redundancy_protocol"`
	MlagInfo           *mlagInfo         `tfsdk:"mlag_info"`
	LogicalDevice      logicalDeviceData `tfsdk:"logical_device"`
	TagData            []tagData         `tfsdk:"tag_data"`
}

func (o dRackTypeLeafSwitch) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                types.StringType,
			"spine_link_count":    types.Int64Type,
			"spine_link_speed":    types.StringType,
			"redundancy_protocol": types.StringType,
			"mlag_info":           mlagInfo{}.attrType(),
			"logical_device":      logicalDeviceData{}.attrType(),
			"tag_data":            types.SetType{ElemType: tagData{}.attrType()}}}
}

func (o *dRackTypeLeafSwitch) parseApi(in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign) {
	o.Name = in.Label
	if fcd != goapstra.FabricConnectivityDesignL3Collapsed {
		count := int64(in.LinkPerSpineCount)
		speed := string(in.LinkPerSpineSpeed)
		o.SpineLinkCount = &count
		o.SpineLinkSpeed = &speed
	}

	if in.RedundancyProtocol != goapstra.LeafRedundancyProtocolNone {
		redundancyProtocol := in.RedundancyProtocol.String()
		o.RedundancyProtocol = &redundancyProtocol
	}

	if in.MlagInfo != nil && in.MlagInfo.LeafLeafLinkCount > 0 {
		o.MlagInfo = &mlagInfo{}
		o.MlagInfo.parseApi(in.MlagInfo)
	}

	o.LogicalDevice.parseApi(in.LogicalDevice)

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}
}

type mlagInfo struct {
	MlagKeepaliveVLan       int64   `tfsdk:"mlag_keepalive_vlan"`
	PeerLinkCount           int64   `tfsdk:"peer_link_count"`
	PeerLinkSpeed           string  `tfsdk:"peer_link_speed"`
	PeerLinkPortChannelId   *int64  `tfsdk:"peer_link_port_channel_id"`
	L3PeerLinkCount         *int64  `tfsdk:"l3_peer_link_count"`
	L3PeerLinkSpeed         *string `tfsdk:"l3_peer_link_speed"`
	L3PeerLinkPortChannelId *int64  `tfsdk:"l3_peer_link_port_channel_id"`
}

func (o mlagInfo) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"mlag_keepalive_vlan":          types.Int64Type,
			"peer_link_count":              types.Int64Type,
			"peer_link_speed":              types.StringType,
			"peer_link_port_channel_id":    types.Int64Type,
			"l3_peer_link_count":           types.Int64Type,
			"l3_peer_link_speed":           types.StringType,
			"l3_peer_link_port_channel_id": types.Int64Type}}
}

func (o *mlagInfo) parseApi(in *goapstra.LeafMlagInfo) {
	var peerLinkPortChannelId *int64
	if in.LeafLeafLinkPortChannelId > 0 {
		x := int64(in.LeafLeafLinkPortChannelId)
		peerLinkPortChannelId = &x
	}

	var l3PeerLinkPortChannelId *int64
	if in.LeafLeafL3LinkPortChannelId > 0 {
		x := int64(in.LeafLeafL3LinkPortChannelId)
		l3PeerLinkPortChannelId = &x
	}

	var l3PeerLinkCount *int64
	if in.LeafLeafL3LinkCount > 0 {
		x := int64(in.LeafLeafL3LinkCount)
		l3PeerLinkCount = &x
	}

	var l3PeerLinkSpeed *string
	if in.LeafLeafL3LinkSpeed != "" {
		x := string(in.LeafLeafL3LinkSpeed)
		l3PeerLinkSpeed = &x
	}

	o.MlagKeepaliveVLan = int64(in.MlagVlanId)
	o.PeerLinkCount = int64(in.LeafLeafLinkCount)
	o.PeerLinkSpeed = string(in.LeafLeafLinkSpeed)
	o.PeerLinkPortChannelId = peerLinkPortChannelId
	o.L3PeerLinkCount = l3PeerLinkCount
	o.L3PeerLinkSpeed = l3PeerLinkSpeed
	o.L3PeerLinkPortChannelId = l3PeerLinkPortChannelId
}

func (o *mlagInfo) request() *goapstra.LeafMlagInfo {
	if o == nil {
		return nil
	}

	var leafLeafL3LinkCount int
	if o.L3PeerLinkCount != nil {
		leafLeafL3LinkCount = int(*o.L3PeerLinkCount)
	}

	var leafLeafL3LinkPortChannelId int
	if o.L3PeerLinkPortChannelId != nil {
		leafLeafL3LinkPortChannelId = int(*o.L3PeerLinkPortChannelId)
	}

	var leafLeafLinkPortChannelId int
	if o.PeerLinkPortChannelId != nil {
		leafLeafLinkPortChannelId = int(*o.PeerLinkPortChannelId)
	}

	var leafLeafL3LinkSpeed goapstra.LogicalDevicePortSpeed
	if o.L3PeerLinkSpeed != nil {
		leafLeafL3LinkSpeed = goapstra.LogicalDevicePortSpeed(*o.L3PeerLinkSpeed)
	}

	return &goapstra.LeafMlagInfo{
		LeafLeafL3LinkCount:         leafLeafL3LinkCount,
		LeafLeafL3LinkPortChannelId: leafLeafL3LinkPortChannelId,
		LeafLeafL3LinkSpeed:         leafLeafL3LinkSpeed,
		LeafLeafLinkCount:           int(o.PeerLinkCount),
		LeafLeafLinkPortChannelId:   leafLeafLinkPortChannelId,
		LeafLeafLinkSpeed:           goapstra.LogicalDevicePortSpeed(o.PeerLinkSpeed),
		MlagVlanId:                  int(o.MlagKeepaliveVLan),
	}
}

type dRackTypeAccessSwitch struct {
	Name               string            `tfsdk:"name"`
	Count              int64             `tfsdk:"count"`
	RedundancyProtocol *string           `tfsdk:"redundancy_protocol"`
	EsiLagInfo         *esiLagInfo       `tfsdk:"esi_lag_info"`
	LogicalDevice      logicalDeviceData `tfsdk:"logical_device"`
	TagData            []tagData         `tfsdk:"tag_data"`
	Links              []dRackLink       `tfsdk:"links"`
}

func (o dRackTypeAccessSwitch) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                types.StringType,
			"count":               types.Int64Type,
			"redundancy_protocol": types.StringType,
			"esi_lag_info":        esiLagInfo{}.attrType(),
			"logical_device":      logicalDeviceData{}.attrType(),
			"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
			"links":               types.SetType{ElemType: dRackLink{}.attrType()}}}
}

func (o *dRackTypeAccessSwitch) parseApi(in *goapstra.RackElementAccessSwitch) {
	o.Name = in.Label
	o.Count = int64(in.InstanceCount)
	if in.RedundancyProtocol != goapstra.AccessRedundancyProtocolNone {
		redundancyProtocol := in.RedundancyProtocol.String()
		o.RedundancyProtocol = &redundancyProtocol
	}
	if in.EsiLagInfo != nil {
		o.EsiLagInfo = &esiLagInfo{}
		o.EsiLagInfo.parseApi(in.EsiLagInfo)
	}
	o.LogicalDevice.parseApi(in.LogicalDevice)

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}

	o.Links = make([]dRackLink, len(in.Links))
	for i := range in.Links {
		o.Links[i].parseApi(&in.Links[i])
	}
}

type esiLagInfo struct {
	L3PeerLinkCount int64  `tfsdk:"l3_peer_link_count"`
	L3PeerLinkSpeed string `tfsdk:"l3_peer_link_speed"`
}

func (o esiLagInfo) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"l3_peer_link_count": types.Int64Type,
			"l3_peer_link_speed": types.StringType}}
}

func (o *esiLagInfo) parseApi(in *goapstra.EsiLagInfo) {
	o.L3PeerLinkCount = int64(in.AccessAccessLinkCount)
	o.L3PeerLinkSpeed = string(in.AccessAccessLinkSpeed)
}

func dLinksAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details links from this Element to switches upstream switches within this Rack Type.",
		Computed:            true,
		Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "Name of this link.",
				Computed:            true,
				Type:                types.StringType,
			},
			"target_switch_name": {
				MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
				Computed:            true,
				Type:                types.StringType,
			},
			"lag_mode": {
				MarkdownDescription: "LAG negotiation mode of the Link.",
				Computed:            true,
				Type:                types.StringType,
			},
			"links_per_switch": {
				MarkdownDescription: "Number of Links to each switch.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"speed": {
				MarkdownDescription: "Speed of this Link.",
				Computed:            true,
				Type:                types.StringType,
			},
			"switch_peer": {
				MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
				Computed:            true,
				Type:                types.StringType,
			},
			"tag_data": tagsDataAttributeSchema(),
		}),
	}
}

type dRackLink struct {
	Name             string    `tfsdk:"name"`
	TargetSwitchName string    `tfsdk:"target_switch_name"`
	LagMode          *string   `tfsdk:"lag_mode"`
	LinksPerSwitch   int64     `tfsdk:"links_per_switch"`
	Speed            string    `tfsdk:"speed"`
	SwitchPeer       *string   `tfsdk:"switch_peer"`
	TagData          []tagData `tfsdk:"tag_data"`
}

func (o dRackLink) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":               types.StringType,
			"target_switch_name": types.StringType,
			"lag_mode":           types.StringType,
			"links_per_switch":   types.Int64Type,
			"speed":              types.StringType,
			"switch_peer":        types.StringType,
			"tag_data":           types.SetType{ElemType: tagData{}.attrType()}}}
}

func (o *dRackLink) parseApi(in *goapstra.RackLink) {
	o.Name = in.Label
	o.TargetSwitchName = in.TargetSwitchLabel
	if in.LagMode != goapstra.RackLinkLagModeNone {
		lagMode := in.LagMode.String()
		o.LagMode = &lagMode
	}
	o.LinksPerSwitch = int64(in.LinkPerSwitchCount)
	o.Speed = string(in.LinkSpeed)
	if in.SwitchPeer != goapstra.RackLinkSwitchPeerNone {
		switchPeer := in.SwitchPeer.String()
		o.SwitchPeer = &switchPeer
	}

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}
}

type dRackTypeGenericSystem struct {
	Name             string            `tfsdk:"name"`
	Count            int64             `tfsdk:"count"`
	PortChannelIdMin int64             `tfsdk:"port_channel_id_min"`
	PortChannelIdMax int64             `tfsdk:"port_channel_id_max"`
	LogicalDevice    logicalDeviceData `tfsdk:"logical_device"`
	TagData          []tagData         `tfsdk:"tag_data"`
	Links            []dRackLink       `tfsdk:"links"`
}

func (o dRackTypeGenericSystem) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                types.StringType,
			"count":               types.Int64Type,
			"port_channel_id_min": types.Int64Type,
			"port_channel_id_max": types.Int64Type,
			"logical_device":      logicalDeviceData{}.attrType(),
			"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
			"links":               types.SetType{ElemType: dRackLink{}.attrType()}}}
}

func (o *dRackTypeGenericSystem) parseApi(in *goapstra.RackElementGenericSystem) {
	o.Name = in.Label
	o.Count = int64(in.Count)
	o.PortChannelIdMin = int64(in.PortChannelIdMin)
	o.PortChannelIdMax = int64(in.PortChannelIdMax)
	o.LogicalDevice.parseApi(in.LogicalDevice)
	o.Links = make([]dRackLink, len(in.Links))

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}

	for i := range in.Links {
		o.Links[i].parseApi(&in.Links[i])
	}
}
