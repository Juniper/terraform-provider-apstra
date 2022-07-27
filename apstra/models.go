package apstra

import (
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
	OnBox          types.Bool   `tfsdk:"on_box"`
}

type ResourceRackType struct {
	Id                       types.String             `tfsdk:"id"`
	Name                     types.String             `tfsdk:"name"`
	Description              types.String             `tfsdk:"description"`
	FabricConnectivityDesign types.String             `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             map[string]LeafSwitch    `tfsdk:"leaf_switches"`
	GenericSystems           map[string]GenericSystem `tfsdk:"generic_systems"`
	//AccessSwitches         map[string]AccessSwitch  `tfsdk:"access_switches"`
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

type tagClone struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

type LeafSwitch struct {
	LogicalDeviceId    types.String `tfsdk:"logical_device_id"`
	LinkPerSpineCount  types.Int64  `tfsdk:"spine_link_count"`
	LinkPerSpineSpeed  types.String `tfsdk:"spine_link_speed"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"` // ["","mlag","esi"]
	//Tags               []types.String `tfsdk:"tags"` // todo: need to handle the read-only problem
	//MlagVlanId                  types.Int64  `tfsdk:"mlag_vlan_id"`                 // required w/"mlag"
	//LeafLeafLinkCount           types.Int64  `tfsdk:"peer_link_count"`              // required w/"mlag"
	//LeafLeafLinkSpeed           types.String `tfsdk:"peer_link_speed"`              // required w/"mlag"
	//LeafLeafLinkPortChannelId   types.Int64  `tfsdk:"peer_link_port_channel_id"`    // required w/"mlag"
	//LeafLeafL3LinkCount         types.Int64  `tfsdk:"l3_peer_link_count"`           // required w/"mlag"
	//LeafLeafL3LinkSpeed         types.String `tfsdk:"l3_peer_link_speed"`           // required w/"mlag"
	//LeafLeafL3LinkPortChannelId types.Int64  `tfsdk:"l3_peer_link_port_channel_id"` // required w/"mlag"
}

type GenericSystem struct {
	Count            types.Int64  `tfsdk:"count"`
	LogicalDeviceId  types.String `tfsdk:"logical_device_id"`
	PortChannelIdMin types.Int64  `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64  `tfsdk:"port_channel_id_max"`
	//Tags             []types.String `tfsdk:"tags"` // todo: need to handle the read-only problem
	Links []GSLink `tfsdk:"links"`
}

type GSLink struct {
	Name               types.String `tfsdk:"name"`               // none
	TargetSwitchLabel  types.String `tfsdk:"target_switch_name"` // none
	LagMode            types.String `tfsdk:"lag_mode"`           // none (omit)
	LinkPerSwitchCount types.Int64  `tfsdk:"links_per_switch"`   // none
	Speed              types.String `tfsdk:"speed"`              // none
	SwitchPeer         types.String `tfsdk:"switch_peer""`
	//Tags             []types.String `tfsdk:"tags"` // todo: need to handle the read-only problem
}

type AccessSwitch struct{} // todo
type ASLink struct{}       //todo
