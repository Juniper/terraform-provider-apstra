package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
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
					"tags":           tagsDataAttributeSchema(),
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
							"l3_link_count": {
								MarkdownDescription: "Count of L3 links to ESI peer.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"l3_link_speed": {
								MarkdownDescription: "Speed of L3 links to ESI peer.",
								Computed:            true,
								Type:                types.StringType,
							},
						}),
					},
					"logical_device": logicalDeviceDataAttributeSchema(),
					"tags":           tagsDataAttributeSchema(),
					"links":          linksAttributeSchema(),
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
					"tags":           tagsDataAttributeSchema(),
					"links":          linksAttributeSchema(),
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

	if (config.Name.Null && config.Id.Null) || (!config.Name.Null && !config.Id.Null) { // XOR
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
	if config.Name.Null == false { // fetch rack type by name
		rt, err = o.client.GetRackTypeByName(ctx, config.Name.Value)
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with name '%s' does not exist", config.Name.Value))
			return
		}
	}
	if config.Id.Null == false { // fetch rack type by ID
		rt, err = o.client.GetRackType(ctx, goapstra.ObjectId(config.Id.Value))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with id '%s' does not exist", config.Id.Value))
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

	newState := &dRackType{}
	newState.parseApiResponse(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
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

func dLeafSwitchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"mlag_info": types.ObjectType{
			AttrTypes: mlagInfoAttrTypes()},
		"logical_device": types.ObjectType{
			AttrTypes: logicalDeviceDataElementAttrTypes()},
		"tags": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: tagDataAttrTypes()}},
	}
}

func dAccessSwitchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"redundancy_protocol": types.StringType,
		"esi_lag_info": types.ObjectType{
			AttrTypes: esiLagInfoAttrTypes()},
		"logical_device": types.ObjectType{
			AttrTypes: logicalDeviceDataElementAttrTypes()},
		"tags": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: tagDataAttrTypes()}},
		"links": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: linksAttrTypes()}},
	}
}

func dGenericSystemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
		"logical_device": types.ObjectType{
			AttrTypes: logicalDeviceDataElementAttrTypes()},
		"tags": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: tagDataAttrTypes()}},
		"links": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: linksAttrTypes()}},
	}
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
		"l3_link_count": types.Int64Type,
		"l3_link_speed": types.StringType,
	}
}

func linksAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"switch_peer":        types.StringType,
	}
}

func newLeafSwitchSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: dLeafSwitchAttrTypes()},
	}
}

func newAccessSwitchSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: dAccessSwitchAttrTypes()},
	}
}

func newGenericSystemSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: dGenericSystemAttrTypes()},
	}
}

func newLinkSet(size int) types.Set {
	return types.Set{
		Elems: make([]attr.Value, size),
		ElemType: types.ObjectType{
			AttrTypes: linksAttrTypes()},
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
		l3PeerLinkCount.Null = true
		l3PeerLinkSPeed.Null = true
		l3PeerLinkPortChannelId.Null = true
	} else {
		l3PeerLinkCount.Value = int64(in.LeafLeafL3LinkCount)
		l3PeerLinkSPeed.Value = string(in.LeafLeafL3LinkSpeed)
		l3PeerLinkPortChannelId.Value = int64(in.LeafLeafL3LinkPortChannelId)
	}

	return types.Object{
		AttrTypes: mlagInfoAttrTypes(),
		Attrs: map[string]attr.Value{
			"mlag_keepalive_vlan":          types.Int64{Value: int64(in.MlagVlanId)},
			"peer_link_count":              types.Int64{Value: int64(in.LeafLeafLinkCount)},
			"peer_link_speed":              types.String{Value: string(in.LeafLeafLinkSpeed)},
			"peer_link_port_channel_id":    types.Int64{Value: int64(in.LeafLeafLinkPortChannelId)},
			"l3_peer_link_count":           l3PeerLinkCount,
			"l3_peer_link_speed":           l3PeerLinkSPeed,
			"l3_peer_link_port_channel_id": l3PeerLinkPortChannelId,
		},
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
			"l3_link_count": types.Int64{Value: int64(in.AccessAccessLinkCount)},
			"l3_link_speed": types.String{Value: string(in.AccessAccessLinkSpeed)},
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
			AttrTypes: linksAttrTypes(),
			Attrs: map[string]attr.Value{
				"name":               types.String{Value: link.Label},
				"target_switch_name": types.String{Value: link.TargetSwitchLabel},
				"lag_mode":           types.String{Value: link.LagMode.String()},
				"links_per_switch":   types.Int64{Value: int64(link.LinkPerSwitchCount)},
				"speed":              types.String{Value: string(link.LinkSpeed)},
				"switch_peer":        switchPeer,
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

func (o *dRackType) parseApiResponse(rt *goapstra.RackType, diags *diag.Diagnostics) {
	o.Id = types.String{Value: string(rt.Id)}
	o.Name = types.String{Value: rt.Data.DisplayName}
	o.Description = types.String{Value: rt.Data.Description}
	o.FabricConnectivityDesign = types.String{Value: rt.Data.FabricConnectivityDesign.String()}
	o.parseApiResponseLeafSwitches(rt.Data.LeafSwitches, diags)
	o.parseApiResponseAccessSwitches(rt.Data.AccessSwitches, diags)
	o.parseApiResponseGenericSystems(rt.Data.GenericSystems, diags)
}

func (o *dRackType) parseApiResponseLeafSwitches(in []goapstra.RackElementLeafSwitch, diags *diag.Diagnostics) {
	o.LeafSwitches = newLeafSwitchSet(len(in))
	for i, ls := range in {
		o.parseApiResponseLeafSwitch(&ls, i, diags)
	}
}

func (o *dRackType) parseApiResponseLeafSwitch(in *goapstra.RackElementLeafSwitch, idx int, diags *diag.Diagnostics) {
	o.LeafSwitches.Elems[idx] = types.Object{
		AttrTypes: dLeafSwitchAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: in.Label},
			"spine_link_count":    parseApiLeafSwitchLinkPerSpineCountToTypesInt64(in),
			"spine_link_speed":    parseApiLeafSwitchLinkPerSpineSpeedToTypesString(in),
			"redundancy_protocol": parseApiLeafRedundancyProtocolToTypesString(in),
			"logical_device":      parseApiLogicalDeviceToTypesObject(in.LogicalDevice),
			"mlag_info":           parseApiLeafMlagInfoToTypesObject(in.MlagInfo),
			"tags":                parseApiSliceTagDataToTypesSetObject(in.Tags),
		},
	}
}

func (o *dRackType) parseApiResponseAccessSwitches(in []goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.AccessSwitches = newAccessSwitchSet(len(in))
	for i, as := range in {
		o.parseApiResponseAccessSwitch(&as, i, diags)
	}
}

func (o *dRackType) parseApiResponseAccessSwitch(in *goapstra.RackElementAccessSwitch, idx int, diags *diag.Diagnostics) {
	var redundancyProtocol types.String
	if in.RedundancyProtocol == goapstra.AccessRedundancyProtocolNone {
		redundancyProtocol = types.String{Null: true}
	} else {
		redundancyProtocol = types.String{Value: in.RedundancyProtocol.String()}
	}

	o.AccessSwitches.Elems[idx] = types.Object{
		AttrTypes: dAccessSwitchAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: in.Label},
			"count":               types.Int64{Value: int64(in.InstanceCount)},
			"redundancy_protocol": redundancyProtocol,
			"esi_lag_info":        parseApiAccessEsiLagInfoToTypesObject(in.EsiLagInfo),
			"logical_device":      parseApiLogicalDeviceToTypesObject(in.LogicalDevice),
			"tags":                parseApiSliceTagDataToTypesSetObject(in.Tags),
			"links":               parseApiSliceRackLinkToTypesSetObject(in.Links),
		},
	}
}

func (o *dRackType) parseApiResponseGenericSystems(in []goapstra.RackElementGenericSystem, diags *diag.Diagnostics) {
	o.GenericSystems = newGenericSystemSet(len(in))
	for i, gs := range in {
		o.parseApiResponseGenericSystem(&gs, i, diags)
	}
}

func (o *dRackType) parseApiResponseGenericSystem(in *goapstra.RackElementGenericSystem, idx int, diagnostics *diag.Diagnostics) {
	o.GenericSystems.Elems[idx] = types.Object{
		AttrTypes: dGenericSystemAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: in.Label},
			"count":               types.Int64{Value: int64(in.Count)},
			"port_channel_id_min": types.Int64{Value: int64(in.PortChannelIdMin)},
			"port_channel_id_max": types.Int64{Value: int64(in.PortChannelIdMax)},
			"logical_device":      parseApiLogicalDeviceToTypesObject(in.LogicalDevice),
			"tags":                parseApiSliceTagDataToTypesSetObject(in.Tags),
			"links":               parseApiSliceRackLinkToTypesSetObject(in.Links),
		},
	}
}

func linksAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details links from this Access Switch to other switches in this Rack Type.",
		Computed:            true,
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
			//"tags": {
			//	MarkdownDescription: "Details any tags applied to the Link",
			//	Computed:            true,
			//	Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//		"label": {
			//			MarkdownDescription: "Tag label (name) field.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"description": {
			//			MarkdownDescription: "Tag description field.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//	}),
			//},
		}),
	}
}
