package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DataAgentProfile struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Platform    types.String `tfsdk:"platform"`
	HasUsername types.Bool   `tfsdk:"has_username"`
	HasPassword types.Bool   `tfsdk:"has_password"`
	Packages    types.Map    `tfsdk:"packages"`
	OpenOptions types.Map    `tfsdk:"open_options"`
}

type DataAgentProfileId struct {
	Id    types.String   `tfsdk:"id"`
	Label types.String   `tfsdk:"label"`
	Tags  []types.String `tfsdk:"tags"`
}

type DataAgentProfileIds struct {
	Ids []types.String `tfsdk:"ids"`
}

type DataAsnPool struct {
	Id             types.String   `tfsdk:"id"`
	Name           types.String   `tfsdk:"name"`
	Status         types.String   `tfsdk:"status"`
	Tags           []types.String `tfsdk:"tags"`
	Used           types.Int64    `tfsdk:"used"`
	UsedPercent    types.Float64  `tfsdk:"used_percentage"`
	CreatedAt      types.String   `tfsdk:"created_at"`
	LastModifiedAt types.String   `tfsdk:"last_modified_at"`
	Total          types.Int64    `tfsdk:"total"`
	Ranges         []AsnRange     `tfsdk:"ranges"`
}

type DataIp4Pool struct {
	Id             types.String   `tfsdk:"id"`
	Name           types.String   `tfsdk:"name"`
	Status         types.String   `tfsdk:"status"`
	Tags           []types.String `tfsdk:"tags"`
	Used           types.Int64    `tfsdk:"used"`
	UsedPercent    types.Float64  `tfsdk:"used_percentage"`
	CreatedAt      types.String   `tfsdk:"created_at"`
	LastModifiedAt types.String   `tfsdk:"last_modified_at"`
	Total          types.Int64    `tfsdk:"total"`
	Subnets        []Ip4Subnet    `tfsdk:"subnets"`
}

type DataAsnPoolId struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Tags []types.String `tfsdk:"tags"`
}

type DataAsnPoolIds struct {
	Ids []types.String `tfsdk:"ids"`
}

type DataIp4PoolId struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Tags []types.String `tfsdk:"tags"`
}

type DataIp4PoolIds struct {
	Ids []types.String `tfsdk:"ids"`
}

type DataLogicalDevice struct {
	Id     types.String         `tfsdk:"id"`
	Name   types.String         `tfsdk:"name"`
	Panels []LogicalDevicePanel `tfsdk:"panels"`
}

type DataRackType struct {
	Id                       types.String      `tfsdk:"id"`
	Name                     types.String      `tfsdk:"name"`
	Description              types.String      `tfsdk:"description"`
	FabricConnectivityDesign types.String      `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             []DSLeafSwitch    `tfsdk:"leaf_switches"`
	AccessSwitches           []DSAccessSwitch  `tfsdk:"access_switches"`
	GenericSystems           []DSGenericSystem `tfsdk:"generic_systems"`
}

type DataTag struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type ResourceAgentProfile struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Platform    types.String `tfsdk:"platform"`
	Packages    types.Map    `tfsdk:"packages"`
	OpenOptions types.Map    `tfsdk:"open_options"`
}

type ResourceAsnPool struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Tags []types.String `tfsdk:"tags"`
}

type ResourceAsnPoolRange struct {
	PoolId types.String `tfsdk:"pool_id"`
	First  types.Int64  `tfsdk:"first"`
	Last   types.Int64  `tfsdk:"last"`
}

type ResourceBlueprint struct {
	Id         types.String      `tfsdk:"id"`
	Name       types.String      `tfsdk:"name"`
	TemplateId types.String      `tfsdk:"template_id"`
	LeafAsns   []types.String    `tfsdk:"leaf_asn_pool_ids"`
	LeafIp4s   []types.String    `tfsdk:"leaf_ip_pool_ids"`
	LinkIp4s   []types.String    `tfsdk:"link_ip_pool_ids"`
	SpineAsns  []types.String    `tfsdk:"spine_asn_pool_ids"`
	SpineIp4s  []types.String    `tfsdk:"spine_ip_pool_ids"`
	Switches   map[string]Switch `tfsdk:"switches"`
}

type ResourceIp4Pool struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Tags []types.String `tfsdk:"tags"`
}

type ResourceIp4Subnet struct {
	PoolId types.String `tfsdk:"pool_id"`
	Cidr   types.String `tfsdk:"cidr"`
}

type ResourceManagedDevice struct {
	AgentId        types.String `tfsdk:"agent_id"`
	SystemId       types.String `tfsdk:"system_id"`
	ManagementIp   types.String `tfsdk:"management_ip"`
	DeviceKey      types.String `tfsdk:"device_key"`
	AgentProfileId types.String `tfsdk:"agent_profile_id"`
	AgentLabel     types.String `tfsdk:"agent_label"`
	Location       types.String `tfsdk:"location"`
	OffBox         types.Bool   `tfsdk:"off_box"`
}

type ResourceRackType struct {
	Id                       types.String     `tfsdk:"id"`
	Name                     types.String     `tfsdk:"name"`
	Description              types.String     `tfsdk:"description"`
	FabricConnectivityDesign types.String     `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             []RLeafSwitch    `tfsdk:"leaf_switches"`
	GenericSystems           []RGenericSystem `tfsdk:"generic_systems"`
	AccessSwitches           []RAccessSwitch  `tfsdk:"access_switches"`
}

type ResourceWireframe struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Tags []types.String `tfsdk:"tags"`
}

// helper structs used by 'resource' and 'data source' objects follow

type AsnRange struct {
	Status         types.String  `tfsdk:"status"`
	First          types.Int64   `tfsdk:"first"`
	Last           types.Int64   `tfsdk:"last"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

type Ip4Subnet struct {
	Status         types.String  `tfsdk:"status"`
	Network        types.String  `tfsdk:"network"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

type LogicalDevicePanel struct {
	Rows       types.Int64              `tfsdk:"rows"`
	Columns    types.Int64              `tfsdk:"columns"`
	PortGroups []LogicalDevicePortGroup `tfsdk:"port_groups"`
}

type LogicalDevicePortGroup struct {
	Count types.Int64    `tfsdk:"port_count"`
	Speed types.Int64    `tfsdk:"port_speed_gbps"`
	Roles []types.String `tfsdk:"port_roles"`
}

type Switch struct {
	InterfaceMap  types.String `tfsdk:"interface_map"`
	DeviceKey     types.String `tfsdk:"device_key"`
	DeviceProfile types.String `tfsdk:"device_profile"`
	SystemNodeId  types.String `tfsdk:"system_node_id"`
}

type tagLabels []types.String

func (o tagLabels) toGoapstraTagLabels() []goapstra.TagLabel {
	result := make([]goapstra.TagLabel, len(o))
	for i, tl := range o {
		result[i] = goapstra.TagLabel(tl.Value)
	}
	return result
}

type rackLinks []RackLink

func (o rackLinks) toGoapstraRackLinkRequests() []goapstra.RackLinkRequest {
	result := make([]goapstra.RackLinkRequest, len(o))
	for i, link := range o {
		result[i] = goapstra.RackLinkRequest{
			Label:              link.Name.Value,
			TargetSwitchLabel:  link.TargetSwitchLabel.Value,
			LagMode:            linkLagModeToGoapstraLagMode(link.LagMode),
			LinkPerSwitchCount: int(link.LinkPerSwitchCount.Value),
			LinkSpeed:          goapstra.LogicalDevicePortSpeed(link.Speed.Value),
			SwitchPeer:         linkSwitchPeerToGoapstraSwitchPeer(link.SwitchPeer),
			AttachmentType:     linkLagModeToGoapstraAttachmentType(link.LagMode),
			Tags:               link.TagLabels.toGoapstraTagLabels(),
		}
	}
	return result
}

type RLeafSwitch struct {
	Name               types.String         `tfsdk:"name"`
	LogicalDeviceId    types.String         `tfsdk:"logical_device_id"`
	LinkPerSpineCount  types.Int64          `tfsdk:"spine_link_count"`
	LinkPerSpineSpeed  types.String         `tfsdk:"spine_link_speed"`
	RedundancyProtocol types.String         `tfsdk:"redundancy_protocol"`
	DisplayName        types.String         `tfsdk:"display_name"`
	Panels             []LogicalDevicePanel `tfsdk:"panels"`
	MlagInfo           *MlagInfo            `tfsdk:"mlag_info"`
	TagLabels          tagLabels            `tfsdk:"tags"`
	TagData            []TagData            `tfsdk:"tag_data"`
}

func (o *RLeafSwitch) redundancyProtocol() goapstra.LeafRedundancyProtocol {
	switch o.RedundancyProtocol.Value {
	case goapstra.LeafRedundancyProtocolMlag.String():
		return goapstra.LeafRedundancyProtocolMlag
	case goapstra.LeafRedundancyProtocolEsi.String():
		return goapstra.LeafRedundancyProtocolEsi
	default:
		return goapstra.LeafRedundancyProtocolNone
	}
}

func (o *RLeafSwitch) checkMlagInfoPresent(diags *diag.Diagnostics) {
	if o.RedundancyProtocol.Value == goapstra.LeafRedundancyProtocolMlag.String() && o.MlagInfo == nil {
		diags.AddError("invalid configuration",
			fmt.Sprintf("'mlag_info' must be set when 'redundancy_mode' is %s",
				o.RedundancyProtocol.Value))
	}
}

func (o *RLeafSwitch) copyWriteOnlyAttributesFrom(src *RLeafSwitch, diags *diag.Diagnostics) {
	// copy Logical Device ID
	o.LogicalDeviceId = types.String{Value: src.LogicalDeviceId.Value}
	// copy Tag Labels
	o.TagLabels = make(tagLabels, len(src.TagLabels))
	copy(o.TagLabels, src.TagLabels)
}

type RAccessSwitch struct {
	Name               types.String         `tfsdk:"name"`
	Count              types.Int64          `tfsdk:"count"`
	LogicalDeviceId    types.String         `tfsdk:"logical_device_id"` // needs to be cloned from state on Read()
	RedundancyProtocol types.String         `tfsdk:"redundancy_protocol"`
	DisplayName        types.String         `tfsdk:"display_name"`
	Links              rackLinks            `tfsdk:"links"`
	TagLabels          tagLabels            `tfsdk:"tags"` // needs to be cloned from state on Read()
	TagData            []TagData            `tfsdk:"tag_data"`
	EsiLagInfo         *EsiLagInfo          `tfsdk:"esi_lag_info"`
	Panels             []LogicalDevicePanel `tfsdk:"panels"`
}

func (o *RAccessSwitch) redundancyProtocol() goapstra.AccessRedundancyProtocol {
	switch o.RedundancyProtocol.Value {
	case goapstra.AccessRedundancyProtocolEsi.String():
		return goapstra.AccessRedundancyProtocolEsi
	default:
		return goapstra.AccessRedundancyProtocolNone
	}
}

func (o *RAccessSwitch) checkEsiLagInfoPresent(diags *diag.Diagnostics) {
	if o.RedundancyProtocol.Value == goapstra.AccessRedundancyProtocolEsi.String() && o.EsiLagInfo == nil {
		diags.AddError("invalid configuration",
			fmt.Sprintf("'esi_lag_info' must be set when 'redundancy_mode' is %s",
				o.RedundancyProtocol.Value))
	}
}

func (o *RAccessSwitch) checkLinksTargetLeafs(rt *ResourceRackType, diags *diag.Diagnostics) {
	for _, link := range o.Links {
		_, reType := rt.findDeviceIndexAndTypeByName(link.TargetSwitchLabel.Value, diags)
		if reType != rackElementTypeLeafSwitch {
			diags.AddError("invalid configuration",
				fmt.Sprintf("Access Switch '%s' Link '%s' targets something other than a Leaf Switch",
					o.Name.Value, link.Name.Value))
		}
	}
}

func (o *RAccessSwitch) checkLinksLagConfig(rt *ResourceRackType, diags *diag.Diagnostics) {
	for _, link := range o.Links {
		link.checkLinkLagConfig(rt, link.Name.Value, diags)
	}
}

//goland:noinspection GoUnusedParameter
func (o *RAccessSwitch) copyWriteOnlyAttributesFrom(src *RAccessSwitch, diags *diag.Diagnostics) {
	// copy Logical Device ID
	o.LogicalDeviceId = types.String{Value: src.LogicalDeviceId.Value}
	// copy Tag Labels
	o.TagLabels = make(tagLabels, len(src.TagLabels))
	copy(o.TagLabels, src.TagLabels)
}

type RGenericSystem struct {
	Name             types.String         `tfsdk:"name"`
	Count            types.Int64          `tfsdk:"count"`
	LogicalDeviceId  types.String         `tfsdk:"logical_device_id"`
	PortChannelIdMin types.Int64          `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64          `tfsdk:"port_channel_id_max"`
	TagLabels        tagLabels            `tfsdk:"tags"` // needs to be cloned from state on Read()
	TagData          []TagData            `tfsdk:"tag_data"`
	Links            rackLinks            `tfsdk:"links"`
	DisplayName      types.String         `tfsdk:"display_name"` //todo add to schema
	Panels           []LogicalDevicePanel `tfsdk:"panels"`
	//AsnDomain:       // not exposed in WebUI, so skipping
	//ManagementLevel: // not exposed in WebUI, so skipping
	//Loopback:        // not exposed in WebUI, so skipping

}

func (o *RGenericSystem) getRackLinkIndexByName(name string, diags *diag.Diagnostics) int {
	for i, rl := range o.Links {
		if rl.Name.Value == name {
			return i
		}
	}
	diags.AddError("generic system link not found",
		fmt.Sprintf("generic system link named '%s' was not found in the Generic System definition returned by Apstra",
			name))
	return -1
}

func (o *RGenericSystem) checkPoIdMinMax(diags *diag.Diagnostics) {
	if o.PortChannelIdMin.IsNull() && o.PortChannelIdMax.IsNull() {
		return
	}
	if !o.PortChannelIdMin.IsNull() || !o.PortChannelIdMax.IsNull() {
		diags.AddError("invalid configuration",
			"if one of 'port_channel_id_min' and 'port_channel_id_max' is set, then both must be set")
	}
	if o.PortChannelIdMax.Value < o.PortChannelIdMin.Value {
		diags.AddError("invalid configuration",
			fmt.Sprintf("Generic System '%s' 'port_channel_id_max' is less than 'port_channel_id_min'",
				o.Name.Value))
	}
}

func (o *RGenericSystem) checkLinksLagConfig(rt *ResourceRackType, diags *diag.Diagnostics) {
	for _, link := range o.Links {
		link.checkLinkLagConfig(rt, link.Name.Value, diags)
	}
}

func (o *RGenericSystem) copyWriteOnlyAttributesFrom(src *RGenericSystem, diags *diag.Diagnostics) {
	// copy Logical Device ID
	o.LogicalDeviceId = types.String{Value: src.LogicalDeviceId.Value}
	// copy Tag Labels
	o.TagLabels = make(tagLabels, len(src.TagLabels))
	copy(o.TagLabels, src.TagLabels)
	// copy Tag Labels of each Link
	for _, srcRackLink := range src.Links {
		dstRackLinkIndex := o.getRackLinkIndexByName(srcRackLink.Name.Value, diags)
		if diags.HasError() {
			return
		}
		o.Links[dstRackLinkIndex].TagLabels = make(tagLabels, len(srcRackLink.TagLabels))
		copy(o.Links[dstRackLinkIndex].TagLabels, srcRackLink.TagLabels)
	}
}

type DSLeafSwitch struct {
	Name               types.String         `tfsdk:"name"`
	LinkPerSpineCount  types.Int64          `tfsdk:"spine_link_count"`
	LinkPerSpineSpeed  types.String         `tfsdk:"spine_link_speed"`
	RedundancyProtocol types.String         `tfsdk:"redundancy_protocol"`
	DisplayName        types.String         `tfsdk:"display_name"`
	TagLabels          tagLabels            `tfsdk:"tags"` // needs to be cloned from state on Read()
	TagData            []TagData            `tfsdk:"tag_data"`
	MlagInfo           *MlagInfo            `tfsdk:"mlag_info"`
	Panels             []LogicalDevicePanel `tfsdk:"panels"`
}

type DSAccessSwitch struct {
	Name               types.String         `tfsdk:"name"`
	DisplayName        types.String         `tfsdk:"display_name"`
	Count              types.Int64          `tfsdk:"count"`
	RedundancyProtocol types.String         `tfsdk:"redundancy_protocol"`
	Links              rackLinks            `tfsdk:"links"`
	Panels             []LogicalDevicePanel `tfsdk:"panels"`
	Tags               []TagData            `tfsdk:"tags"`
	EsiLagInfo         *EsiLagInfo          `tfsdk:"esi_lag_info"`
}

type DSGenericSystem struct {
	Name             types.String         `tfsdk:"name"`
	DisplayName      types.String         `tfsdk:"display_name"`
	Count            types.Int64          `tfsdk:"count"`
	PortChannelIdMin types.Int64          `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64          `tfsdk:"port_channel_id_max"`
	Tags             []TagData            `tfsdk:"tags"`
	Panels           []LogicalDevicePanel `tfsdk:"panels"`
	Links            rackLinks            `tfsdk:"links"`
}

type MlagInfo struct {
	VlanId                      types.Int64  `tfsdk:"mlag_keepalive_vlan"`
	LeafLeafLinkCount           types.Int64  `tfsdk:"peer_link_count"`
	LeafLeafLinkSpeed           types.String `tfsdk:"peer_link_speed"`
	LeafLeafLinkPortChannelId   types.Int64  `tfsdk:"peer_link_port_channel_id"`
	LeafLeafL3LinkCount         types.Int64  `tfsdk:"l3_peer_link_count"`
	LeafLeafL3LinkSpeed         types.String `tfsdk:"l3_peer_link_speed"`
	LeafLeafL3LinkPortChannelId types.Int64  `tfsdk:"l3_peer_link_port_channel_id"`
}

type EsiLagInfo struct {
	AccessAccessLinkCount types.Int64  `tfsdk:"peer_link_count"`
	AccessAccessLinkSpeed types.String `tfsdk:"peer_link_speed"`
}

// TagData is tag data w/out ID element, as cloned into some other design object.
type TagData struct {
	Label       types.String `tfsdk:"label"`
	Description types.String `tfsdk:"description"`
}

type RackLink struct {
	Name               types.String `tfsdk:"name"`
	TargetSwitchLabel  types.String `tfsdk:"target_switch_name"`
	LagMode            types.String `tfsdk:"lag_mode"`
	LinkPerSwitchCount types.Int64  `tfsdk:"links_per_switch"`
	Speed              types.String `tfsdk:"speed"`
	SwitchPeer         types.String `tfsdk:"switch_peer"`
	TagLabels          tagLabels    `tfsdk:"tags"` // needs to be cloned from state on Read()
	TagData            []TagData    `tfsdk:"tag_data"`
}

func (o *RackLink) checkLinkLagConfig(rt *ResourceRackType, originDevice string, diags *diag.Diagnostics) {
	if o.LagMode.IsNull() {
		// no LAG
		if o.SwitchPeer.IsNull() && rt.switchIsRedundant(o.TargetSwitchLabel.Value, diags) {
			// no LAG, no switch peer, but switch is redundant
			diags.AddError("invalid configuration",
				fmt.Sprintf("Rack Element '%s', link '%s', 'switch_peer' required with non-redundant switch '%s'",
					originDevice, o.Name.Value, o.TargetSwitchLabel.Value))
		}
	} else {
		// LAG
		if !o.SwitchPeer.IsNull() {
			// LAG, but switch peer specified
			diags.AddError("invalid configuration",
				fmt.Sprintf("Rack Element '%s', link '%s', 'switch_peer' cannot be used with LAG enabled",
					originDevice, o.Name.Value))
		}
	}
}
