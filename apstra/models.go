package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DataAgentProfileId struct {
	Id    types.String   `tfsdk:"id"`
	Label types.String   `tfsdk:"label"`
	Tags  []types.String `tfsdk:"tags"`
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

type ResourceIp4PoolSubnet struct {
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

type ResourceWireframe struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Tags []types.String `tfsdk:"tags"`
}

// helper structs used by 'resource' and 'data source' objects follow

type Switch struct {
	InterfaceMap  types.String `tfsdk:"interface_map"`
	DeviceKey     types.String `tfsdk:"device_key"`
	DeviceProfile types.String `tfsdk:"device_profile"`
	SystemNodeId  types.String `tfsdk:"system_node_id"`
}

type tagLabels []types.String

func (o tagLabels) toGoapstraTagLabels() []string {
	result := make([]string, len(o))
	for i, tl := range o {
		result[i] = tl.Value
	}
	return result
}

type DSLeafSwitch struct {
	Name               types.String         `tfsdk:"name"`
	LinkPerSpineCount  types.Int64          `tfsdk:"spine_link_count"`
	LinkPerSpineSpeed  types.String         `tfsdk:"spine_link_speed"`
	RedundancyProtocol types.String         `tfsdk:"redundancy_protocol"`
	DisplayName        types.String         `tfsdk:"display_name"`
	TagLabels          tagLabels            `tfsdk:"tags"` // needs to be cloned from state on Read()
	TagData            []types.Object       `tfsdk:"tag_data"`
	MlagInfo           *MlagInfo            `tfsdk:"mlag_info"`
	Panels             []LogicalDevicePanel `tfsdk:"panels"`
}

type DSAccessSwitch struct {
	Name               types.String         `tfsdk:"name"`
	DisplayName        types.String         `tfsdk:"display_name"`
	Count              types.Int64          `tfsdk:"count"`
	RedundancyProtocol types.String         `tfsdk:"redundancy_protocol"`
	Links              []RackLink           `tfsdk:"links"`
	Panels             []LogicalDevicePanel `tfsdk:"panels"`
	Tags               []types.Object       `tfsdk:"tags"`
	EsiLagInfo         *EsiLagInfo          `tfsdk:"esi_lag_info"`
}

type DSGenericSystem struct {
	Name             types.String         `tfsdk:"name"`
	DisplayName      types.String         `tfsdk:"display_name"`
	Count            types.Int64          `tfsdk:"count"`
	PortChannelIdMin types.Int64          `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64          `tfsdk:"port_channel_id_max"`
	Tags             []types.Object       `tfsdk:"tags"`
	Panels           []LogicalDevicePanel `tfsdk:"panels"`
	Links            []RackLink           `tfsdk:"links"`
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
	Name               types.String   `tfsdk:"name"`
	TargetSwitchLabel  types.String   `tfsdk:"target_switch_name"`
	LagMode            types.String   `tfsdk:"lag_mode"`
	LinkPerSwitchCount types.Int64    `tfsdk:"links_per_switch"`
	Speed              types.String   `tfsdk:"speed"`
	SwitchPeer         types.String   `tfsdk:"switch_peer"`
	TagLabels          tagLabels      `tfsdk:"tags"` // needs to be cloned from state on Read()
	TagData            []types.Object `tfsdk:"tag_data"`
}

// todo: delete this eventually?
type LogicalDevicePanel struct {
	Rows    types.Int64 `tfsdk:"rows"`
	Columns types.Int64 `tfsdk:"columns"`
	//PortGroups []LogicalDevicePortGroup `tfsdk:"port_groups"`
}
