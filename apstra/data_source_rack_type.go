package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
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
	"os"
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
						MarkdownDescription: "Indicates the role of the switch within the rack, also used for targeting in-rack links.",
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
						MarkdownDescription: "Indicates the role of the switch within the rack, also used for targeting in-rack links.",
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
						MarkdownDescription: "Indicates the role of the generic system within the rack.",
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
	if config.Name.IsNull() == false { // fetch rack type by name
		rt, err = o.client.GetRackTypeByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with name '%s' does not exist", config.Name.ValueString()))
			return
		}
	}
	if config.Id.IsNull() == false { // fetch rack type by ID
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

	validateRackType(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//x, _ := json.MarshalIndent(rt, "", "  ")
	//resp.Diagnostics.AddWarning("rt", string(x))
	newState := &dRackType{}
	newState.parseApiResponse(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newStyle := &dRackType{}
	newStyle.parseApi(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newStateDump, err := json.MarshalIndent(newState, "", "  ")
	if err != nil {
		resp.Diagnostics.AddWarning("newStateDump Error", err.Error())
	}
	newStyleDump, err := json.MarshalIndent(newStyle, "", "  ")
	if err != nil {
		resp.Diagnostics.AddWarning("newStyleDump Error", err.Error())
	}
	os.WriteFile("/tmp/newState", newStateDump, 0644)
	os.WriteFile("/tmp/newStyle", newStyleDump, 0644)

	resp.Diagnostics.AddWarning("newStyle", newStyle.Id.String())
	resp.Diagnostics.AddWarning("newState", newState.Id.String())
	//resp.Diagnostics.AddWarning("description", newStyle.Description.String())
	//resp.Diagnostics.AddWarning("fcd", newStyle.FabricConnectivityDesign.String())
	//newState.LeafSwitches = types.SetNull(dRackTypeLeafSwitch{}.attrType())
	//newState.AccessSwitches = types.SetNull(dRackTypeAccessSwitch{}.attrType())
	//newState.GenericSystems = types.SetNull(dRackTypeGenericSystem{}.attrType())
	//
	//dump, err := json.MarshalIndent(newState, "", "  ")
	//resp.Diagnostics.AddWarning("dump", string(dump))

	//Set state
	diags = resp.State.Set(ctx, newStyle)
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

func dLeafSwitchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"mlag_info":           mlagInfoAttrType(),
		"logical_device":      logicalDeviceAttrType(),
		"tag_data":            tagDataAttrType(),
	}
}

func dAccessSwitchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"redundancy_protocol": types.StringType,
		"esi_lag_info":        esiLagInfoAttrType(),
		"logical_device":      logicalDeviceAttrType(),
		"tag_data":            tagDataAttrType(),
		"links":               dLinksElemType(),
	}
}

func dGenericSystemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
		"logical_device":      logicalDeviceAttrType(),
		"tag_data":            tagDataAttrType(),
		"links":               dLinksElemType(),
	}
}

func mlagInfoAttrType() attr.Type {
	return types.ObjectType{
		AttrTypes: mlagInfoAttrTypes()}
}

func mlagInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mlag_keepalive_vlan":          types.Int64Type,
		"peer_link_count":              types.Int64Type,
		"peer_link_speed":              types.StringType,
		"peer_link_port_channel_id":    types.Int64Type,
		"l3_peer_link_count":           types.Int64Type,
		"l3_peer_link_speed":           types.StringType,
		"l3_peer_link_port_channel_id": types.Int64Type,
	}
}

func esiLagInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"l3_peer_link_count": types.Int64Type,
		"l3_peer_link_speed": types.StringType,
	}
}

func esiLagInfoAttrType() attr.Type {
	return types.ObjectType{
		AttrTypes: esiLagInfoAttrTypes()}
}

func dLinksAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"switch_peer":        types.StringType,
		"tag_data":           tagDataAttrType(),
	}
}

func dLinksElemType() attr.Type {
	return types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: dLinksAttrTypes()}}
}

func newDLeafSwitchSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: dLeafSwitchAttrTypes()},
	}
}

func newDAccessSwitchSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: dAccessSwitchAttrTypes()},
	}
}

func newDGenericSystemSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: dGenericSystemAttrTypes()},
	}
}

func newLinkSet(size int) types.Set {
	return types.Set{
		Elems: make([]attr.Value, size),
		ElemType: types.ObjectType{
			AttrTypes: dLinksAttrTypes()},
	}
}

func parseApiLeafSwitchLinkPerSpineCountToTypesInt64(in *goapstra.RackElementLeafSwitch) types.Int64 {
	if in.LinkPerSpineCount == 0 {
		return types.Int64{Null: true}
	}
	return types.Int64{Value: int64(in.LinkPerSpineCount)}
}

func parseApiLeafSwitchLinkPerSpineSpeedToTypesString(in *goapstra.RackElementLeafSwitch) types.String {
	if in.LinkPerSpineCount == 0 {
		return types.String{Null: true}
	}
	return types.String{Value: string(in.LinkPerSpineSpeed)}
}

func parseApiLeafRedundancyProtocolToTypesString(in *goapstra.RackElementLeafSwitch) types.String {
	if in.RedundancyProtocol == goapstra.LeafRedundancyProtocolNone {
		return types.String{Null: true}
	}
	return types.String{Value: in.RedundancyProtocol.String()}
}

func parseApiLeafMlagInfoToTypesObject(in *goapstra.LeafMlagInfo) types.Object {
	if in == nil || (in.LeafLeafLinkCount == 0 && in.LeafLeafL3LinkCount == 0) {
		return types.Object{
			Null:      true,
			AttrTypes: mlagInfoAttrTypes(),
		}
	}

	var l3PeerLinkCount, l3PeerLinkPortChannelId types.Int64
	var l3PeerLinkSPeed types.String
	if in.LeafLeafL3LinkCount == 0 {
		// link count of zero means all L3 link descriptors should be null
		l3PeerLinkCount.Null = true
		l3PeerLinkSPeed.Null = true
		l3PeerLinkPortChannelId.Null = true
	} else {
		// we have links, so populate attributes from API response
		l3PeerLinkCount.Value = int64(in.LeafLeafL3LinkCount)
		l3PeerLinkSPeed.Value = string(in.LeafLeafL3LinkSpeed)
		if in.LeafLeafL3LinkPortChannelId == 0 {
			// Don't save PoId /0/ - use /null/ instead
			l3PeerLinkPortChannelId.Null = true
		} else {
			l3PeerLinkPortChannelId.Value = int64(in.LeafLeafL3LinkPortChannelId)
		}
	}

	var peerLinkPortChannelId types.Int64
	if in.LeafLeafLinkPortChannelId == 0 {
		// Don't save PoId /0/ - use /null/ instead
		peerLinkPortChannelId.Null = true
	} else {
		peerLinkPortChannelId.Value = int64(in.LeafLeafLinkPortChannelId)
	}

	return types.Object{
		AttrTypes: mlagInfoAttrTypes(),
		Attrs: map[string]attr.Value{
			"mlag_keepalive_vlan":          types.Int64{Value: int64(in.MlagVlanId)},
			"peer_link_count":              types.Int64{Value: int64(in.LeafLeafLinkCount)},
			"peer_link_speed":              types.String{Value: string(in.LeafLeafLinkSpeed)},
			"peer_link_port_channel_id":    peerLinkPortChannelId,
			"l3_peer_link_count":           l3PeerLinkCount,
			"l3_peer_link_speed":           l3PeerLinkSPeed,
			"l3_peer_link_port_channel_id": l3PeerLinkPortChannelId,
		},
	}
}

func parseApiAccessRedundancyProtocolToTypesString(in *goapstra.RackElementAccessSwitch) types.String {
	if in.RedundancyProtocol == goapstra.AccessRedundancyProtocolNone {
		return types.String{Null: true}
	} else {
		return types.String{Value: in.RedundancyProtocol.String()}
	}
}

func parseApiAccessEsiLagInfoToTypesObject(in *goapstra.EsiLagInfo) types.Object {
	if in == nil || in.AccessAccessLinkCount == 0 {
		return types.Object{
			Null:      true,
			AttrTypes: esiLagInfoAttrTypes(),
		}
	}

	return types.Object{
		AttrTypes: esiLagInfoAttrTypes(),
		Attrs: map[string]attr.Value{
			"l3_peer_link_count": types.Int64{Value: int64(in.AccessAccessLinkCount)},
			"l3_peer_link_speed": types.String{Value: string(in.AccessAccessLinkSpeed)},
		},
	}
}

func parseApiSliceRackLinkToTypesSetObject(links []goapstra.RackLink) types.Set {
	result := newLinkSet(len(links))
	for i, link := range links {
		var switchPeer types.String
		if link.SwitchPeer == goapstra.RackLinkSwitchPeerNone {
			switchPeer = types.String{Null: true}
		} else {
			switchPeer = types.String{Value: link.SwitchPeer.String()}
		}
		result.Elems[i] = types.Object{
			AttrTypes: dLinksAttrTypes(),
			Attrs: map[string]attr.Value{
				"name":               types.String{Value: link.Label},
				"target_switch_name": types.String{Value: link.TargetSwitchLabel},
				"lag_mode":           types.String{Value: link.LagMode.String()},
				"links_per_switch":   types.Int64{Value: int64(link.LinkPerSwitchCount)},
				"speed":              types.String{Value: string(link.LinkSpeed)},
				"switch_peer":        switchPeer,
				"tag_data":           parseApiSliceTagDataToTypesSetObject(link.Tags),
			},
		}
	}
	return result
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

	leafSwitches := make([]dRackTypeLeafSwitch, len(in.Data.LeafSwitches))
	for i := range in.Data.LeafSwitches {
		leafSwitches[i].parseApi(&in.Data.LeafSwitches[i])
	}
	leafSwitchSet, d := types.SetValueFrom(ctx, dRackTypeLeafSwitch{}.attrType(), leafSwitches)
	diags.Append(d...)

	accessSwitches := make([]dRackTypeAccessSwitch, len(in.Data.AccessSwitches))
	for i := range in.Data.AccessSwitches {
		accessSwitches[i].parseApi(&in.Data.AccessSwitches[i])
	}
	accessSwitchSet, d := types.SetValueFrom(ctx, dRackTypeAccessSwitch{}.attrType(), accessSwitches)
	diags.Append(d...)

	genericSystems := make([]dRackTypeGenericSystem, len(in.Data.GenericSystems))
	for i := range in.Data.GenericSystems {
		genericSystems[i].parseApi(&in.Data.GenericSystems[i])
	}
	genericSystemSet, d := types.SetValueFrom(ctx, dRackTypeGenericSystem{}.attrType(), genericSystems)
	diags.Append(d...)

	//dump, _ := json.MarshalIndent(in, "", "  ")
	//diags.AddWarning("goapstra dump", string(dump))

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Description = types.StringValue(in.Data.Description)
	o.FabricConnectivityDesign = types.StringValue(in.Data.FabricConnectivityDesign.String())
	o.LeafSwitches = leafSwitchSet
	o.AccessSwitches = accessSwitchSet
	o.GenericSystems = genericSystemSet

	////dump, _ := json.MarshalIndent(o, "", "  ")
	//diags.AddWarning("in.Id", string(in.Id))
	//diags.AddWarning("in.Data.DisplayName", in.Data.DisplayName)
	//diags.AddWarning("in.Data.Description", in.Data.Description)
	//diags.AddWarning("in.Data.FabricConnectivityDesign", in.Data.FabricConnectivityDesign.String())
}

type dRackTypeLeafSwitch struct {
	Name               string            `tfsdk:"name"`
	SpineLinkCount     int64             `tfsdk:"spine_link_count"`
	SpineLinkSpeed     string            `tfsdk:"spine_link_speed"`
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

func (o *dRackTypeLeafSwitch) parseApi(in *goapstra.RackElementLeafSwitch) {
	o.Name = in.Label
	o.SpineLinkCount = int64(in.LinkPerSpineCount)
	o.SpineLinkSpeed = string(in.LinkPerSpineSpeed)

	if in.RedundancyProtocol != goapstra.LeafRedundancyProtocolNone {
		redundancyProtocol := in.RedundancyProtocol.String()
		o.RedundancyProtocol = &redundancyProtocol
	}

	if in.MlagInfo != nil && in.MlagInfo.LeafLeafLinkCount > 0 {
		o.MlagInfo.parseApi(in.MlagInfo)
	}

	o.LogicalDevice.parseApi(in.LogicalDevice)
	o.TagData = make([]tagData, len(in.Tags)) // populated below

	for i := range in.Tags {
		o.TagData[i].parseApi(&in.Tags[i])
	}
}

type mlagInfo struct {
	MlagKeepaliveVLan       int64  `tfsdk:"mlag_keepalive_vlan"`
	PeerLinkCount           int64  `tfsdk:"peer_link_count"`
	PeerLinkSpeed           string `tfsdk:"peer_link_speed"`
	PeerLinkPortChannelId   int64  `tfsdk:"peer_link_port_channel_id"`
	L3PeerLinkCount         int64  `tfsdk:"l3_peer_link_count"`
	L3PeerLinkSpeed         string `tfsdk:"l3_peer_link_speed"`
	L3PeerLinkPortChannelId int64  `tfsdk:"l3_peer_link_port_channel_id"`
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
	o.MlagKeepaliveVLan = int64(in.MlagVlanId)
	o.PeerLinkCount = int64(in.LeafLeafLinkCount)
	o.PeerLinkSpeed = string(in.LeafLeafLinkSpeed)
	o.PeerLinkPortChannelId = int64(in.LeafLeafLinkPortChannelId)
	o.L3PeerLinkCount = int64(in.LeafLeafL3LinkCount)
	o.L3PeerLinkSpeed = string(in.LeafLeafL3LinkSpeed)
	o.L3PeerLinkPortChannelId = int64(in.LeafLeafL3LinkPortChannelId)
}

type dRackTypeAccessSwitch struct {
	Name               string             `tfsdk:"name"`
	Count              int64              `tfsdk:"count"`
	RedundancyProtocol string             `tfsdk:"redundancy_protocol"`
	EsiLagInfo         *esiLagInfo        `tfsdk:"esi_lag_info"`
	LogicalDevice      *logicalDeviceData `tfsdk:"logical_device"`
	TagData            []tagData          `tfsdk:"tag_data"`
	Links              []rackLink         `tfsdk:"links"`
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
			"links":               types.SetType{ElemType: rackLink{}.attrType()}}}
}

func (o *dRackTypeAccessSwitch) parseApi(in *goapstra.RackElementAccessSwitch) {
	o.Name = in.Label
	o.Count = int64(in.InstanceCount)
	o.RedundancyProtocol = in.RedundancyProtocol.String()
	if in.EsiLagInfo != nil {
		o.EsiLagInfo = &esiLagInfo{}
		o.EsiLagInfo.parseApi(in.EsiLagInfo)
	}
	if in.LogicalDevice != nil {
		o.LogicalDevice = &logicalDeviceData{}
		o.LogicalDevice.parseApi(in.LogicalDevice)
	}

	o.TagData = make([]tagData, len(in.Tags)) // populated below
	for i := range in.Tags {
		o.TagData[i].parseApi(&in.Tags[i])
	}

	o.Links = make([]rackLink, len(in.Links))
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

type rackLink struct {
	Name             string    `tfsdk:"name"`
	TargetSwitchName string    `tfsdk:"target_switch_name"`
	LagMode          string    `tfsdk:"lag_mode"`
	LinksPerSwitch   int64     `tfsdk:"links_per_switch"`
	Speed            string    `tfsdk:"speed"`
	SwitchPeer       string    `tfsdk:"switch_peer"`
	TagData          []tagData `tfsdk:"tag_data"`
}

func (o *rackLink) parseApi(in *goapstra.RackLink) {
	o.Name = in.Label
	o.TargetSwitchName = in.TargetSwitchLabel
	o.LagMode = in.LagMode.String()
	o.LinksPerSwitch = int64(in.LinkPerSwitchCount)
	o.Speed = string(in.LinkSpeed)
	o.SwitchPeer = in.SwitchPeer.String()
	o.TagData = make([]tagData, len(in.Tags))

	for i := range in.Tags {
		o.TagData[i].parseApi(&in.Tags[i])
	}
}

func (o rackLink) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":               types.StringType,
			"target_switch_name": types.StringType,
			"lag_mode":           types.StringType,
			"links_per_switch":   types.Int64Type,
			"speed":              types.StringType,
			"switch_peer":        types.StringType,
			"tag_data":           types.SetType{ElemType: tagData{}.attrType()},
		},
	}
}

type dRackTypeGenericSystem struct {
	Name             string            `tfsdk:"name"`
	Count            int64             `tfsdk:"count"`
	PortChannelIdMin int64             `tfsdk:"port_channel_id_min"`
	PortChannelIdMax int64             `tfsdk:"port_channel_id_max"`
	LogicalDevice    logicalDeviceData `tfsdk:"logical_device"`
	TagData          []tagData         `tfsdk:"tag_data"`
	Links            []rackLink        `tfsdk:"links"`
}

func (o *dRackTypeGenericSystem) parseApi(in *goapstra.RackElementGenericSystem) {
	o.Name = in.Label
	o.Count = int64(in.Count)
	o.PortChannelIdMin = int64(in.PortChannelIdMin)
	o.PortChannelIdMax = int64(in.PortChannelIdMax)
	o.LogicalDevice.parseApi(in.LogicalDevice)
	o.TagData = make([]tagData, len(in.Tags))
	o.Links = make([]rackLink, len(in.Links))

	for i := range in.Tags {
		o.TagData[i].parseApi(&in.Tags[i])
	}

	for i := range in.Links {
		o.Links[i].parseApi(&in.Links[i])
	}
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
			"links":               types.SetType{ElemType: rackLink{}.attrType()},
		},
	}
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

func (o *dRackType) parseApiResponse(ctx context.Context, rt *goapstra.RackType, diags *diag.Diagnostics) {
	o.Id = types.String{Value: string(rt.Id)}
	o.Name = types.String{Value: rt.Data.DisplayName}
	o.Description = types.String{Value: rt.Data.Description}
	o.FabricConnectivityDesign = types.String{Value: rt.Data.FabricConnectivityDesign.String()}
	o.parseApiResponseLeafSwitches(ctx, rt.Data.LeafSwitches, diags)
	o.parseApiResponseAccessSwitches(ctx, rt.Data.AccessSwitches, diags)
	o.parseApiResponseGenericSystems(ctx, rt.Data.GenericSystems, diags)
}

func (o *dRackType) parseApiResponseLeafSwitches(ctx context.Context, in []goapstra.RackElementLeafSwitch, diags *diag.Diagnostics) {
	o.LeafSwitches = newDLeafSwitchSet(len(in))
	for i, ls := range in {
		o.parseApiResponseLeafSwitch(ctx, &ls, i, diags)
	}
}

func (o *dRackType) parseApiResponseLeafSwitch(ctx context.Context, in *goapstra.RackElementLeafSwitch, idx int, diags *diag.Diagnostics) {
	o.LeafSwitches.Elems[idx] = types.Object{
		AttrTypes: dLeafSwitchAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: in.Label},
			"spine_link_count":    parseApiLeafSwitchLinkPerSpineCountToTypesInt64(in),
			"spine_link_speed":    parseApiLeafSwitchLinkPerSpineSpeedToTypesString(in),
			"redundancy_protocol": parseApiLeafRedundancyProtocolToTypesString(in),
			"logical_device":      parseApiLogicalDeviceToTypesObject(ctx, in.LogicalDevice, diags),
			"mlag_info":           parseApiLeafMlagInfoToTypesObject(in.MlagInfo),
			"tag_data":            parseApiSliceTagDataToTypesSetObject(in.Tags),
		},
	}
}

func (o *dRackType) parseApiResponseAccessSwitches(ctx context.Context, in []goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.AccessSwitches = newDAccessSwitchSet(len(in))
	for i, as := range in {
		o.parseApiResponseAccessSwitch(ctx, &as, i, diags)
	}
}

func (o *dRackType) parseApiResponseAccessSwitch(ctx context.Context, in *goapstra.RackElementAccessSwitch, idx int, diags *diag.Diagnostics) {
	o.AccessSwitches.Elems[idx] = types.Object{
		AttrTypes: dAccessSwitchAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: in.Label},
			"count":               types.Int64{Value: int64(in.InstanceCount)},
			"redundancy_protocol": parseApiAccessRedundancyProtocolToTypesString(in),
			"esi_lag_info":        parseApiAccessEsiLagInfoToTypesObject(in.EsiLagInfo),
			"logical_device":      parseApiLogicalDeviceToTypesObject(ctx, in.LogicalDevice, diags),
			"tag_data":            parseApiSliceTagDataToTypesSetObject(in.Tags),
			"links":               parseApiSliceRackLinkToTypesSetObject(in.Links),
		},
	}
}

func (o *dRackType) parseApiResponseGenericSystems(ctx context.Context, in []goapstra.RackElementGenericSystem, diags *diag.Diagnostics) {
	o.GenericSystems = newDGenericSystemSet(len(in))
	for i, gs := range in {
		o.parseApiResponseGenericSystem(ctx, &gs, i, diags)
	}
}

func (o *dRackType) parseApiResponseGenericSystem(ctx context.Context, in *goapstra.RackElementGenericSystem, idx int, diags *diag.Diagnostics) {
	o.GenericSystems.Elems[idx] = types.Object{
		AttrTypes: dGenericSystemAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: in.Label},
			"count":               types.Int64{Value: int64(in.Count)},
			"port_channel_id_min": types.Int64{Value: int64(in.PortChannelIdMin)},
			"port_channel_id_max": types.Int64{Value: int64(in.PortChannelIdMax)},
			"logical_device":      parseApiLogicalDeviceToTypesObject(ctx, in.LogicalDevice, diags),
			"tag_data":            parseApiSliceTagDataToTypesSetObject(in.Tags),
			"links":               parseApiSliceRackLinkToTypesSetObject(in.Links),
		},
	}
}
