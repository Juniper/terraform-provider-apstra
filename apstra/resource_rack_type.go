package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"regexp"
	"strings"
)

type rackElementType int

const (
	vlanMin = 1
	vlanMax = 4094

	poIdMin = 0
	poIdMax = 4096

	rackElementTypeLeafSwitch = rackElementType(iota)
	rackElementTypeAccessSwitch
	rackElementTypeGenericSystem
)

type resourceRackTypeType struct{}

func (r resourceRackTypeType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	// permitted fabric_connectivity_design mode strings
	fcpModes := []string{
		goapstra.FabricConnectivityDesignL3Clos.String(),
		goapstra.FabricConnectivityDesignL3Collapsed.String()}

	// permitted leaf_switch.redundancy_protocol mode strings
	leafSwitchRedundancyProtocols := []string{
		goapstra.LeafRedundancyProtocolEsi.String(),
		goapstra.LeafRedundancyProtocolMlag.String(),
	}

	//// permitted access_switch.redundancy_protocol mode strings
	//accessSwitchRedundancyProtocols := []string{
	//	goapstra.LeafRedundancyProtocolEsi.String(),
	//}
	//
	//// permitted generic_system.links[*].lag_mode mode strings
	//gsLinkLagModes := []string{
	//	goapstra.RackLinkLagModeActive.String(),
	//	goapstra.RackLinkLagModePassive.String(),
	//	goapstra.RackLinkLagModeStatic.String(),
	//}
	//
	//// permitted access_switch.links[*].switch_peer and generic_system.links[*].switch_peer id strings
	//linkSwitchPeers := []string{
	//	goapstra.RackLinkSwitchPeerFirst.String(),
	//	goapstra.RackLinkSwitchPeerSecond.String(),
	//}

	// regex for validating fabric_connectivity_design mode string
	fcdRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(fcpModes, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling fabric connectivity design regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	// regex for validating leaf_switch.redundancy_protocol mode strings
	leafRedundancyRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(leafSwitchRedundancyProtocols, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling leaf redundancy regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	//// regex for validating access_switch.redundancy_protocol mode strings
	//accessRedundancyRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
	//	strings.Join(accessSwitchRedundancyProtocols, "$|^")))
	//if err != nil {
	//	diagnostics := diag.Diagnostics{}
	//	diagnostics.AddError("error compiling access redundancy regex", err.Error())
	//	return tfsdk.Schema{}, diagnostics
	//}
	//
	//// regex for validating access_switch.link[*].switch_peer and generic_system.link[*].switch_peer id strings
	//linkSwitchPeerRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
	//	strings.Join(linkSwitchPeers, "$|^")))
	//if err != nil {
	//	diagnostics := diag.Diagnostics{}
	//	diagnostics.AddError("error compiling link switch peer regex", err.Error())
	//	return tfsdk.Schema{}, diagnostics
	//}
	//
	//// regex for validating generic_system.links[*].lag_mode mode strings
	//gsLinkLagModeRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
	//	strings.Join(gsLinkLagModes, "$|^")))
	//if err != nil {
	//	diagnostics := diag.Diagnostics{}
	//	diagnostics.AddError("error compiling generic system LAG mode regex", err.Error())
	//	return tfsdk.Schema{}, diagnostics
	//}

	return tfsdk.Schema{
		MarkdownDescription: "This resource creates an Apstra Rack Type.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Object ID for the Rack Type, assigned by Apstra.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Rack Type name, displayed in the Apstra web UI.",
				Type:                types.StringType,
				Required:            true,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"description": {
				MarkdownDescription: "Rack Type description, displayed in the Apstra web UI.",
				Type:                types.StringType,
				Optional:            true,
			},
			"fabric_connectivity_design": {
				MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(fcpModes, "', '")),
				Type:                types.StringType,
				Required:            true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
					fcdRegexp,
					fmt.Sprintf("fabric_connectivity_design must be one of: '%s'",
						strings.Join(fcpModes, "', '")))},
			},
			"leaf_switches": {
				MarkdownDescription: "Each Rack Type is required to have at least one Leaf Switch.",
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Switch name, used when creating intra-rack links.",
						Type:                types.StringType,
						Required:            true,
					},
					"logical_device_id": {
						MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
						Type:                types.StringType,
						Required:            true,
					},
					"spine_link_count": {
						MarkdownDescription: "Links per spine.",
						Type:                types.Int64Type,
						Required:            true,
						Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
					},
					"spine_link_speed": {
						MarkdownDescription: "Speed of spine-facing links, something like '10G'",
						Type:                types.StringType,
						Required:            true,
					},
					"redundancy_protocol": {
						MarkdownDescription: fmt.Sprintf("When set, must be one of '%s'.", strings.Join(leafSwitchRedundancyProtocols, "', '")),
						Type:                types.StringType,
						Optional:            true,
						Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
							leafRedundancyRegexp,
							fmt.Sprintf("redundancy_protocol must be one of: '%s'",
								strings.Join(leafSwitchRedundancyProtocols, "', '")))},
					},
					"display_name": {
						MarkdownDescription: "Name copied from the Logical Device upon which this Leaf Switch was modeled.",
						Computed:            true,
						Type:                types.StringType,
					},
					"panels": {
						MarkdownDescription: "Details physical layout of interfaces on the device.",
						Computed:            true,
						//Optional:            true,
						//PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"rows": {
								MarkdownDescription: "Physical vertical dimension of the panel.",
								Computed:            true,
								//Optional:            true,
								//PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
								Type: types.Int64Type,
							},
							"columns": {
								MarkdownDescription: "Physical horizontal dimension of the panel.",
								Computed:            true,
								//Optional:            true,
								//PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
								Type: types.Int64Type,
							},
							//"port_groups": {
							//	MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
							//	Computed:            true,
							//	Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							//		"port_count": {
							//			MarkdownDescription: "Number of ports in the group.",
							//			Computed:            true,
							//			Type:                types.Int64Type,
							//		},
							//		"port_speed_gbps": {
							//			MarkdownDescription: "Port speed in Gbps.",
							//			Computed:            true,
							//			Type:                types.Int64Type,
							//		},
							//		"port_roles": {
							//			MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
							//			Computed:            true,
							//			Type:                types.SetType{ElemType: types.StringType},
							//		},
							//	}),
							//},
						}),
					},
					"mlag_info": {
						MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, "+
							"defines the connectivity between MLAG peers.", goapstra.LeafRedundancyProtocolMlag.String()),
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"mlag_keepalive_vlan": {
								MarkdownDescription: "MLAG keepalive VLAN ID.",
								Required:            true,
								Type:                types.Int64Type,
								Validators: []tfsdk.AttributeValidator{
									int64validator.Between(vlanMin, vlanMax),
								},
							},
							"peer_link_count": {
								MarkdownDescription: "Number of links between MLAG devices.",
								Required:            true,
								Type:                types.Int64Type,
								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
							},
							"peer_link_speed": {
								MarkdownDescription: "Speed of links between MLAG devices.",
								Required:            true,
								Type:                types.StringType,
							},
							"peer_link_port_channel_id": {
								MarkdownDescription: "Peer link port-channel ID.",
								Required:            true,
								Type:                types.Int64Type,
								Validators: []tfsdk.AttributeValidator{
									int64validator.Between(poIdMin, poIdMax),
								},
							},
							"l3_peer_link_count": {
								MarkdownDescription: "Number of L3 links between MLAG devices.",
								Required:            true,
								Type:                types.Int64Type,
							},
							"l3_peer_link_speed": {
								MarkdownDescription: "Speed of l3 links between MLAG devices.",
								Required:            true,
								Type:                types.StringType,
							},
							"l3_peer_link_port_channel_id": {
								MarkdownDescription: "L3 peer link port-channel ID.",
								Required:            true,
								Type:                types.Int64Type,
								Validators: []tfsdk.AttributeValidator{
									int64validator.Between(poIdMin, poIdMax),
								},
							},
						}),
					},
					"tags": {
						MarkdownDescription: "Labels of tags from the global catalog to be applied to this Leaf Switch upon Rack Type creation",
						Optional:            true,
						Type:                types.SetType{ElemType: types.StringType},
					},
					//		"tag_data": {
					//			MarkdownDescription: "Details any tags applied to the Leaf Switch",
					//			Computed:            true,
					//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					//				"label": {
					//					MarkdownDescription: "Tag label (name) field.",
					//					Optional:            true,
					//					Type:                types.StringType,
					//				},
					//				"description": {
					//					MarkdownDescription: "Tag description field.",
					//					Optional:            true,
					//					Type:                types.StringType,
					//				},
					//			}),
					//		},
				})},
			//"access_switches": {
			//	Optional:            true,
			//	MarkdownDescription: "Template for servers and similar which will be created upon rack instantiation.",
			//	Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//		"name": {
			//			MarkdownDescription: "Name for instances of this Access Switch Type.",
			//			Type:                types.StringType,
			//			Required:            true,
			//		},
			//		"count": {
			//			MarkdownDescription: "Number of Access Switches of this type.",
			//			Type:                types.Int64Type,
			//			Required:            true,
			//			Validators: []tfsdk.AttributeValidator{
			//				int64validator.AtLeast(1),
			//			},
			//		},
			//		"logical_device_id": {
			//			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Access Switch.",
			//			Type:                types.StringType,
			//			Required:            true,
			//		},
			//		"redundancy_protocol": {
			//			MarkdownDescription: fmt.Sprintf("When set, must be one of '%s'.", strings.Join(accessSwitchRedundancyProtocols, "', '")),
			//			Type:                types.StringType,
			//			Optional:            true,
			//			Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
			//				accessRedundancyRegexp,
			//				fmt.Sprintf("redundancy_protocol must be one of: '%s'",
			//					strings.Join(accessSwitchRedundancyProtocols, "', '")))},
			//		},
			//		"tags": {
			//			MarkdownDescription: "Labels of tags from the global catalog to be applied to this Access Switch upon Rack Type creation",
			//			Optional:            true,
			//			Type:                types.SetType{ElemType: types.StringType},
			//		},
			//		"tag_data": {
			//			MarkdownDescription: "Details any tags applied to the Access Switch",
			//			Computed:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"label": {
			//					MarkdownDescription: "Tag label (name) field.",
			//					Optional:            true,
			//					Type:                types.StringType,
			//				},
			//				"description": {
			//					MarkdownDescription: "Tag description field.",
			//					Optional:            true,
			//					Type:                types.StringType,
			//				},
			//			}),
			//		},
			//		"display_name": {
			//			MarkdownDescription: "Name copied from the Logical Device upon which this Access Switch was modeled.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"panels": {
			//			MarkdownDescription: "Details physical layout of interfaces on the device.",
			//			Computed:            true,
			//			Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			//				"rows": {
			//					MarkdownDescription: "Physical vertical dimension of the panel.",
			//					Computed:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"columns": {
			//					MarkdownDescription: "Physical horizontal dimension of the panel.",
			//					Computed:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"port_groups": {
			//					MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			//					Computed:            true,
			//					Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			//						"port_count": {
			//							MarkdownDescription: "Number of ports in the group.",
			//							Computed:            true,
			//							Type:                types.Int64Type,
			//						},
			//						"port_speed_gbps": {
			//							MarkdownDescription: "Port speed in Gbps.",
			//							Computed:            true,
			//							Type:                types.Int64Type,
			//						},
			//						"port_roles": {
			//							MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
			//							Computed:            true,
			//							Type:                types.SetType{ElemType: types.StringType},
			//						},
			//					}),
			//				},
			//			}),
			//		},
			//		"links": {
			//			MarkdownDescription: "Describe links between the Access Switch and other systems within the rack",
			//			Required:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"name": {
			//					MarkdownDescription: "Link name",
			//					Required:            true,
			//					Type:                types.StringType,
			//				},
			//				"target_switch_name": {
			//					MarkdownDescription: "Name of the Leaf Switch to which the Access Switch connects.",
			//					Required:            true,
			//					Type:                types.StringType,
			//				},
			//				"lag_mode": {
			//					MarkdownDescription: "Link LAG mode",
			//					Computed:            true, // always lacp active mode for access->leaf links
			//					Type:                types.StringType,
			//				},
			//				"links_per_switch": {
			//					MarkdownDescription: "Default value '1'.",
			//					Required:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"speed": {
			//					MarkdownDescription: "Link Speed, something like '10G'",
			//					Required:            true,
			//					Type:                types.StringType,
			//				},
			//				"tags": {
			//					MarkdownDescription: "Labels of tags from the global catalog to be applied to this Access Switch upon Rack Type creation",
			//					Optional:            true,
			//					Type:                types.SetType{ElemType: types.StringType},
			//				},
			//				"tag_data": {
			//					MarkdownDescription: "Details any tags applied to the Link",
			//					Computed:            true,
			//					Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//						"label": {
			//							MarkdownDescription: "Tag label (name) field.",
			//							Optional:            true,
			//							Type:                types.StringType,
			//						},
			//						"description": {
			//							MarkdownDescription: "Tag description field.",
			//							Optional:            true,
			//							Type:                types.StringType,
			//						},
			//					}),
			//				},
			//				"switch_peer": {
			//					MarkdownDescription: fmt.Sprintf("For non-LAG links to redundant switches, must be one of '%s'.", strings.Join(linkSwitchPeers, "', '")),
			//					Optional:            true,
			//					Computed:            true,
			//					Type:                types.StringType,
			//					Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
			//						linkSwitchPeerRegexp,
			//						fmt.Sprintf("link switch_peer must be one of: '%s'",
			//							strings.Join(linkSwitchPeers, "', '")))},
			//				},
			//			}),
			//		},
			//		"esi_lag_info": {
			//			MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, defines the connectivity between peers",
			//				goapstra.AccessRedundancyProtocolEsi.String()),
			//			Optional: true,
			//			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			//				"peer_link_count": {
			//					MarkdownDescription: "Number of L3 links between ESI-LAG devices.",
			//					Required:            true,
			//					Type:                types.Int64Type,
			//					Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
			//				},
			//				"peer_link_speed": {
			//					MarkdownDescription: "Speed of l3 links between ESI-LAG devices.",
			//					Required:            true,
			//					Type:                types.StringType,
			//				},
			//			}),
			//		},
			//	}),
			//},
			//"generic_systems": {
			//	Optional:            true,
			//	MarkdownDescription: "Template for servers and similar which will be created upon rack instantiation.",
			//	Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//		"name": {
			//			MarkdownDescription: "Name for instances of this Generic System type.",
			//			Type:                types.StringType,
			//			Required:            true,
			//		},
			//		"count": {
			//			MarkdownDescription: "Number of Generic Systems of this type.",
			//			Type:                types.Int64Type,
			//			Required:            true,
			//			Validators: []tfsdk.AttributeValidator{
			//				int64validator.AtLeast(1),
			//			},
			//		},
			//		"logical_device_id": {
			//			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Generic System.",
			//			Type:                types.StringType,
			//			Required:            true,
			//		},
			//		"port_channel_id_min": {
			//			MarkdownDescription: "Port Channel ID Min. Required when 'port_channel_id_max' is set.",
			//			Type:                types.Int64Type,
			//			Optional:            true,
			//			Validators: []tfsdk.AttributeValidator{
			//				int64validator.Between(poIdMin, poIdMax),
			//			},
			//		},
			//		"port_channel_id_max": {
			//			MarkdownDescription: "Port Channel ID Max. Required when 'port_channel_id_min' is set.",
			//			Type:                types.Int64Type,
			//			Optional:            true,
			//			Validators: []tfsdk.AttributeValidator{
			//				int64validator.Between(poIdMin, poIdMax),
			//			},
			//		},
			//		"tags": {
			//			MarkdownDescription: "Labels of tags from the global catalog to be applied to this Generic System upon Rack Type creation",
			//			Optional:            true,
			//			Type:                types.SetType{ElemType: types.StringType},
			//		},
			//		"tag_data": {
			//			MarkdownDescription: "Details any tags applied to the Generic System",
			//			Computed:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"label": {
			//					MarkdownDescription: "Tag label (name) field.",
			//					Optional:            true,
			//					Type:                types.StringType,
			//				},
			//				"description": {
			//					MarkdownDescription: "Tag description field.",
			//					Optional:            true,
			//					Type:                types.StringType,
			//				},
			//			}),
			//		},
			//		"display_name": {
			//			MarkdownDescription: "Name copied from the Logical Device upon which this Access Switch was modeled.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"panels": {
			//			MarkdownDescription: "Details physical layout of interfaces on the device.",
			//			Computed:            true,
			//			Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			//				"rows": {
			//					MarkdownDescription: "Physical vertical dimension of the panel.",
			//					Computed:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"columns": {
			//					MarkdownDescription: "Physical horizontal dimension of the panel.",
			//					Computed:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"port_groups": {
			//					MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			//					Computed:            true,
			//					Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			//						"port_count": {
			//							MarkdownDescription: "Number of ports in the group.",
			//							Computed:            true,
			//							Type:                types.Int64Type,
			//						},
			//						"port_speed_gbps": {
			//							MarkdownDescription: "Port speed in Gbps.",
			//							Computed:            true,
			//							Type:                types.Int64Type,
			//						},
			//						"port_roles": {
			//							MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
			//							Computed:            true,
			//							Type:                types.SetType{ElemType: types.StringType},
			//						},
			//					}),
			//				},
			//			}),
			//		},
			//		"links": {
			//			MarkdownDescription: "Describe links between the generic system and other systems within the rack",
			//			Required:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"name": {
			//					MarkdownDescription: "Link name",
			//					Required:            true,
			//					Type:                types.StringType,
			//				},
			//				"target_switch_name": {
			//					MarkdownDescription: "Name of the Leaf or Access switch to which the Generic System connects.",
			//					Required:            true,
			//					Type:                types.StringType,
			//				},
			//				"lag_mode": {
			//					MarkdownDescription: fmt.Sprintf("LAG Mode must be one of: '%s'.", strings.Join(gsLinkLagModes, "', '")),
			//					Optional:            true,
			//					Type:                types.StringType,
			//					Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
			//						gsLinkLagModeRegexp,
			//						fmt.Sprintf("lag_mode must be one of: '%s'",
			//							strings.Join(gsLinkLagModes, "', '")))},
			//				},
			//				"links_per_switch": {
			//					MarkdownDescription: "Default value '1'.",
			//					Required:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"speed": {
			//					MarkdownDescription: "Link Speed, something like '10G'",
			//					Required:            true,
			//					Type:                types.StringType,
			//				},
			//				"tags": {
			//					MarkdownDescription: "Labels of tags from the global catalog to be applied to this Leaf Switch upon Rack Type creation",
			//					Optional:            true,
			//					Type:                types.SetType{ElemType: types.StringType},
			//				},
			//				"tag_data": {
			//					MarkdownDescription: "Details any tags applied to the Link",
			//					Computed:            true,
			//					Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//						"label": {
			//							MarkdownDescription: "Tag label (name) field.",
			//							Optional:            true,
			//							Type:                types.StringType,
			//						},
			//						"description": {
			//							MarkdownDescription: "Tag description field.",
			//							Optional:            true,
			//							Type:                types.StringType,
			//						},
			//					}),
			//				},
			//				"switch_peer": {
			//					MarkdownDescription: fmt.Sprintf("For non-LAG links to redundant switches, must be one of '%s'.", strings.Join(linkSwitchPeers, "', '")),
			//					Optional:            true,
			//					Computed:            true,
			//					Type:                types.StringType,
			//					Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
			//						linkSwitchPeerRegexp,
			//						fmt.Sprintf("link switch_peer must be one of: '%s'",
			//							strings.Join(linkSwitchPeers, "', '")))},
			//				},
			//			}),
			//		},
			//	}),
			//},
		},
	}, nil
}

func (r resourceRackTypeType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceRackType{
		p: *(p.(*apstraProvider)),
	}, nil
}

type resourceRackType struct {
	p apstraProvider
}

func (r resourceRackType) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config ResourceRackType
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// minimum leaf switch check
	if len(config.LeafSwitches) == 0 {
		resp.Diagnostics.AddError(
			"missing required configuration element",
			"at least one 'leaf_switches' element is required")
		return
	}

	// Leaf switch checks
	for _, ls := range config.LeafSwitches {
		ls.checkMlagInfoPresent(&resp.Diagnostics)
	}

	// todo restore
	//// access switch checks
	//for _, as := range config.AccessSwitches {
	//	as.checkEsiLagInfoPresent(&resp.Diagnostics)
	//	as.checkLinksLagConfig(&config, &resp.Diagnostics)
	//	as.checkLinksTargetLeafs(&config, &resp.Diagnostics)
	//}
	//
	// todo restore
	//// generic system checks
	//for _, gs := range config.GenericSystems {
	//	gs.checkPoIdMinMax(&resp.Diagnostics)
	//	gs.checkLinksLagConfig(&config, &resp.Diagnostics)
	//}

	config.detectDeviceNameCollisions(&resp.Diagnostics)
}

func (r resourceRackType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	//resp.Diagnostics.Append(tfsdk.ValueFrom(ctx, panels, types.ListType{ElemType: types.StringType}, &model.Attribute)...)
	// Retrieve values from plan
	plan := &ResourceRackType{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := os.WriteFile("/tmp/out", []byte("hello\n"), 0644)
	if err != nil {
		resp.Diagnostics.AddError("file write error", err.Error())
	}

	// force non-negotiable values
	plan.forceValues(&resp.Diagnostics)
	if diags.HasError() {
		return
	}

	// Prepare a goapstra.RackTypeRequest
	rtReq := plan.goapstraRequest(&resp.Diagnostics)
	if diags.HasError() {
		return
	}

	// send the request to Apstra
	id, err := r.p.client.CreateRackType(ctx, rtReq)
	if err != nil {
		resp.Diagnostics.AddError("error creating rack type", err.Error())
		return
	}

	// fetch the newly-created rack from Apstra
	rt, err := r.p.client.GetRackType(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving just-created rack type", err.Error())
		return
	}

	// convert apstra response into terraform state object
	newState := goapstraRackTypeToResourceRackType(rt, &resp.Diagnostics)
	jsonData, err := json.Marshal(newState)
	if err != nil {
		resp.Diagnostics.AddWarning("error! (1)", err.Error())
	}
	resp.Diagnostics.AddWarning("newstate", string(jsonData))

	if resp.Diagnostics.HasError() {
		return
	}

	// copy write-only state elements from plan
	newState.copyWriteOnlyAttributesFrom(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	jsonData, err = json.Marshal(newState)
	if err != nil {
		resp.Diagnostics.AddWarning("error! (2)", err.Error())
	}
	resp.Diagnostics.AddWarning("newstate (again)", string(jsonData))

	// commit the state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r resourceRackType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	oldState := &ResourceRackType{}
	diags := req.State.Get(ctx, oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rt, err := r.p.client.GetRackType(ctx, goapstra.ObjectId(oldState.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading rack type from API", err.Error())
		return
	}

	newState := goapstraRackTypeToResourceRackType(rt, &resp.Diagnostics)

	// copy write-only / un-fetchable elements of old state into new state
	newState.copyWriteOnlyAttributesFrom(oldState, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//o, _ := json.Marshal(oldState)
	//n, _ := json.Marshal(newState)
	//resp.Diagnostics.AddWarning("o", string(o))
	//resp.Diagnostics.AddWarning("n", string(n))

	// Set state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Get current state
	state := &ResourceRackType{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	plan := &ResourceRackType{}
	diags = req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// force non-negotiable values
	plan.forceValues(&resp.Diagnostics)
	if diags.HasError() {
		return
	}

	// Prepare a goapstra.RackTypeRequest
	rtReq := plan.goapstraRequest(&resp.Diagnostics)
	if diags.HasError() {
		return
	}

	// Send the request to Apstra as an update
	err := r.p.client.UpdateRackType(ctx, goapstra.ObjectId(state.Id.Value), rtReq)
	if err != nil {
		resp.Diagnostics.AddError("error updating rack type", err.Error())
		return
	}

	// fetch the updated rack from Apstra
	rt, err := r.p.client.GetRackType(ctx, goapstra.ObjectId(plan.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error retrieving just-created rack type", err.Error())
		return
	}

	// convert apstra response into terraform state object
	newState := goapstraRackTypeToResourceRackType(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy write-only state elements from plan
	newState.copyWriteOnlyAttributesFrom(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// commit the state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (r resourceRackType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state ResourceRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete rack type by calling API
	err := r.p.client.DeleteRackType(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"error deleting Rack Type",
			fmt.Sprintf("could not delete Rack Type '%s' - %s", state.Id.Value, err),
		)
		return
	}
}

// copyWriteOnlyAttributesFrom duplicates user input (from the saved state) into the
// new state object instantiated during Read() and Create() operations.
// Currently, those elements are:
// - leaf_switch:
//   - logical_device IDs
//            because the `logical_device` returned in the rack type object does
//            not relate to the create-time `logical_device` ID found in the
//            global catalog
//   - tag_labels:
//            because the tag `label` returned in the rack type object does not
//            relate to the create-time tag `label` found in the global catalog
// - access_switch
//   - logical_device IDs
//            because the `logical_device` returned in the rack type object does
//            not relate to the create-time `logical_device` ID found in the
//            global catalog
//   - tag_labels:
//            because the tag `label` returned in the rack type object does not
//            relate to the create-time tag `label` found in the global catalog
//   - links:
//     - tag_labels:
//            because the tag `label` returned in the rack type object does not
//            relate to the create-time tag `label` found in the global catalog
// - generic_system
//   - logical_device IDs
//            because the `logical_device` returned in the rack type object does
//            not relate to the create-time `logical_device` ID found in the
//            global catalog
//   - tag_labels:
//            because the tag `label` returned in the rack type object does not
//            relate to the create-time tag `label` found in the global catalog
//   - links:
//     - tag labels:
//            because the tag `label` returned in the rack type object does not
//            relate to the create-time tag `label` found in the global catalog
func (o *ResourceRackType) copyWriteOnlyAttributesFrom(src *ResourceRackType, diags *diag.Diagnostics) {
	// todo restore
	for _, srcLeafSwitch := range src.LeafSwitches {
		dstLeafSwitchIndex, _ := o.findDeviceIndexAndTypeByName(srcLeafSwitch.Name.Value, diags)
		if diags.HasError() {
			return
		}
		o.LeafSwitches[dstLeafSwitchIndex].copyWriteOnlyAttributesFrom(&srcLeafSwitch, diags)
	}
	// todo restore
	//// iterate over Access Switches in the Rack Type
	//for _, srcAccessSwitch := range src.AccessSwitches {
	//	dstAccessSwitchIndex, _ := o.findDeviceIndexAndTypeByName(srcAccessSwitch.Name.Value, diags)
	//	if diags.HasError() {
	//		return
	//	}
	//	o.AccessSwitches[dstAccessSwitchIndex].copyWriteOnlyAttributesFrom(&srcAccessSwitch, diags)
	//}
	// todo restore
	//// iterate over Generic Systems in the Rack Type
	//for _, srcGenericSystem := range src.GenericSystems {
	//	dstGenericSystemIndex, _ := o.findDeviceIndexAndTypeByName(srcGenericSystem.Name.Value, diags)
	//	if diags.HasError() {
	//		return
	//	}
	//	o.GenericSystems[dstGenericSystemIndex].copyWriteOnlyAttributesFrom(&srcGenericSystem, diags)
	//}
}

func goapstraRackTypeToResourceRackType(rt *goapstra.RackType, diags *diag.Diagnostics) *ResourceRackType {
	var description types.String
	if rt.Description == "" {
		description = types.String{Null: true}
	} else {
		description = types.String{Value: rt.Description}
	}
	return &ResourceRackType{
		Id:                       types.String{Value: string(rt.Id)},
		Name:                     types.String{Value: rt.DisplayName},
		Description:              description,
		FabricConnectivityDesign: types.String{Value: rt.FabricConnectivityDesign.String()},
		LeafSwitches:             goApstraRackTypeToRLeafSwitches(rt, diags),
		// todo restore
		//AccessSwitches:           goApstraRackTypeToRAccessSwitches(rt, diags),
		//GenericSystems:           goApstraRackTypeToRGenericSystems(rt, diags),
	}
}

func goApstraRackTypeToRAccessSwitches(rt *goapstra.RackType, diags *diag.Diagnostics) []RAccessSwitch {
	// return a nil slice rather than zero-length slice to prevent state churn
	if len(rt.AccessSwitches) == 0 {
		return nil
	}

	result := make([]RAccessSwitch, len(rt.AccessSwitches))
	for i, access := range rt.AccessSwitches {
		result[i] = RAccessSwitch{
			Name:               types.String{Value: access.Label},
			Count:              types.Int64{Value: int64(access.InstanceCount)},
			LogicalDeviceId:    types.String{Unknown: true}, // this value cannot be polled from the API
			RedundancyProtocol: stringerToTfString(access.RedundancyProtocol),
			DisplayName:        types.String{Value: access.DisplayName},
			Links:              sliceGoapstraRackLinksToTfRackLinks(access.Links, diags),
			TagLabels:          nil, // copied in later by copyWriteOnlyAttributesFrom()
			TagData:            sliceGoapstraTagDataToSliceTfTagData(access.Tags, diags),
			EsiLagInfo: &EsiLagInfo{
				AccessAccessLinkCount: types.Int64{Value: int64(access.AccessAccessLinkCount)},
				AccessAccessLinkSpeed: types.String{Value: string(access.AccessAccessLinkSpeed)},
			},
			Panels: goApstraPanelsToTfPanels(access.Panels, diags),
		}
	}
	return result
}

func goApstraRackTypeToRLeafSwitches(rt *goapstra.RackType, diags *diag.Diagnostics) []RLeafSwitch {
	// return a nil slice rather than zero-length slice to prevent state churn
	if len(rt.LeafSwitches) == 0 {
		return nil
	}

	result := make([]RLeafSwitch, len(rt.LeafSwitches))
	for i, leaf := range rt.LeafSwitches {
		result[i] = RLeafSwitch{
			Name:               types.String{Value: leaf.Label},
			LogicalDeviceId:    types.String{Unknown: true}, // copied in later by copyWriteOnlyAttributesFrom()
			LinkPerSpineCount:  types.Int64{Value: int64(leaf.LinkPerSpineCount)},
			LinkPerSpineSpeed:  types.String{Value: string(leaf.LinkPerSpineSpeed)},
			RedundancyProtocol: stringerToTfString(leaf.RedundancyProtocol),
			DisplayName:        types.String{Value: leaf.DisplayName},
			Panels:             goApstraPanelsToTfPanels(leaf.Panels, diags),
			//TagLabels: nil, // copied in later by copyWriteOnlyAttributesFrom()
			//TagData:   sliceGoapstraTagDataToSliceTfTagData(leaf.Tags, diags),
		}
		if leaf.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag {
			result[i].MlagInfo = &MlagInfo{
				VlanId:                      types.Int64{Value: int64(leaf.MlagVlanId)},
				LeafLeafLinkCount:           types.Int64{Value: int64(leaf.LeafLeafLinkCount)},
				LeafLeafLinkSpeed:           types.String{Value: string(leaf.LeafLeafLinkSpeed)},
				LeafLeafLinkPortChannelId:   types.Int64{Value: int64(leaf.LeafLeafLinkPortChannelId)},
				LeafLeafL3LinkCount:         types.Int64{Value: int64(leaf.LeafLeafL3LinkCount)},
				LeafLeafL3LinkSpeed:         types.String{Value: string(leaf.LeafLeafL3LinkSpeed)},
				LeafLeafL3LinkPortChannelId: types.Int64{Value: int64(leaf.LeafLeafL3LinkPortChannelId)},
			}
		}
	}
	return result
}

func goApstraRackTypeToRGenericSystems(rt *goapstra.RackType, diags *diag.Diagnostics) []RGenericSystem {
	// return a nil slice rather than zero-length slice to prevent state churn
	if len(rt.GenericSystems) == 0 {
		return nil
	}

	result := make([]RGenericSystem, len(rt.GenericSystems))
	for i, generic := range rt.GenericSystems {
		result[i] = RGenericSystem{
			Name:             types.String{Value: generic.Label},
			Count:            types.Int64{Value: int64(generic.Count)},
			LogicalDeviceId:  types.String{Unknown: true}, // copied in later by copyWriteOnlyAttributesFrom()
			PortChannelIdMin: types.Int64{Value: int64(generic.PortChannelIdMin)},
			PortChannelIdMax: types.Int64{Value: int64(generic.PortChannelIdMax)},
			TagLabels:        nil, // copied in later by copyWriteOnlyAttributesFrom()
			TagData:          sliceGoapstraTagDataToSliceTfTagData(generic.Tags, diags),
			Links:            sliceGoapstraRackLinksToTfRackLinks(generic.Links, diags),
			DisplayName:      types.String{Value: generic.DisplayName},
			Panels:           goApstraPanelsToTfPanels(generic.Panels, diags),
		}
	}
	return result
}

func (o *ResourceRackType) fcd() goapstra.FabricConnectivityDesign {
	switch o.FabricConnectivityDesign.Value {
	case goapstra.FabricConnectivityDesignL3Collapsed.String():
		return goapstra.FabricConnectivityDesignL3Collapsed
	default:
		return goapstra.FabricConnectivityDesignL3Clos
	}
}

// forceValues sets object values which are required by Apstra, but we don't
// want to bother the user about.
//goland:noinspection GoUnusedParameter
func (o *ResourceRackType) forceValues(diags *diag.Diagnostics) {
	//// all access -> leaf links must be LACP active
	//for i, accessSwitch := range o.AccessSwitches {
	//	for j := range accessSwitch.Links {
	//		o.AccessSwitches[i].Links[j].LagMode = types.String{Value: goapstra.RackLinkLagModeActive.String()}
	//	}
	//}
}

func (o *ResourceRackType) goapstraRequest(diags *diag.Diagnostics) *goapstra.RackTypeRequest {
	return &goapstra.RackTypeRequest{
		DisplayName:              o.Name.Value,
		Description:              o.Description.Value,
		FabricConnectivityDesign: o.fcd(),
		LeafSwitches:             o.leafSwitchRequests(diags),
		// todo restore
		//AccessSwitches:           o.accessSwitchRequests(diags),
		//GenericSystems:           o.genericSystemRequests(diags),
	}
}

//goland:noinspection GoUnusedParameter
func (o *ResourceRackType) leafSwitchRequests(diags *diag.Diagnostics) []goapstra.RackElementLeafSwitchRequest {
	result := make([]goapstra.RackElementLeafSwitchRequest, len(o.LeafSwitches))
	for i, leafSwitch := range o.LeafSwitches {
		result[i] = goapstra.RackElementLeafSwitchRequest{
			Label:              leafSwitch.Name.Value,
			LogicalDeviceId:    goapstra.ObjectId(leafSwitch.LogicalDeviceId.Value),
			LinkPerSpineCount:  int(leafSwitch.LinkPerSpineCount.Value),
			LinkPerSpineSpeed:  goapstra.LogicalDevicePortSpeed(leafSwitch.LinkPerSpineSpeed.Value),
			RedundancyProtocol: leafSwitch.redundancyProtocol(),
			Tags:               leafSwitch.TagLabels.toGoapstraTagLabels(),
		}
		if leafSwitch.MlagInfo != nil {
			result[i].MlagVlanId = int(leafSwitch.MlagInfo.VlanId.Value)
			result[i].LeafLeafLinkCount = int(leafSwitch.MlagInfo.LeafLeafLinkCount.Value)
			result[i].LeafLeafLinkSpeed = goapstra.LogicalDevicePortSpeed(leafSwitch.MlagInfo.LeafLeafLinkSpeed.Value)
			result[i].LeafLeafLinkPortChannelId = int(leafSwitch.MlagInfo.LeafLeafLinkPortChannelId.Value)
			result[i].LeafLeafL3LinkCount = int(leafSwitch.MlagInfo.LeafLeafL3LinkCount.Value)
			result[i].LeafLeafL3LinkSpeed = goapstra.LogicalDevicePortSpeed(leafSwitch.MlagInfo.LeafLeafL3LinkSpeed.Value)
			result[i].LeafLeafL3LinkPortChannelId = int(leafSwitch.MlagInfo.LeafLeafL3LinkPortChannelId.Value)
		}
	}
	return result
}

////goland:noinspection GoUnusedParameter
//func (o *ResourceRackType) accessSwitchRequests(diags *diag.Diagnostics) []goapstra.RackElementAccessSwitchRequest {
//	result := make([]goapstra.RackElementAccessSwitchRequest, len(o.AccessSwitches))
//	for i, accessSwitch := range o.AccessSwitches {
//		result[i] = goapstra.RackElementAccessSwitchRequest{
//			Label:                 accessSwitch.Name.Value,
//			InstanceCount:         int(accessSwitch.Count.Value),
//			LogicalDeviceId:       goapstra.ObjectId(accessSwitch.Name.Value),
//			RedundancyProtocol:    accessSwitch.redundancyProtocol(),
//			Links:                 accessSwitch.Links.toGoapstraRackLinkRequests(),
//			Tags:                  accessSwitch.TagLabels.toGoapstraTagLabels(),
//			AccessAccessLinkCount: int(accessSwitch.EsiLagInfo.AccessAccessLinkCount.Value),
//			AccessAccessLinkSpeed: goapstra.LogicalDevicePortSpeed(accessSwitch.EsiLagInfo.AccessAccessLinkSpeed.Value),
//		}
//	}
//	return result
//}
//
////goland:noinspection GoUnusedParameter
//func (o *ResourceRackType) genericSystemRequests(diags *diag.Diagnostics) []goapstra.RackElementGenericSystemRequest {
//	result := make([]goapstra.RackElementGenericSystemRequest, len(o.GenericSystems))
//	for i, gs := range o.GenericSystems {
//		result[i] = goapstra.RackElementGenericSystemRequest{
//			Label:            gs.Name.Value,
//			Count:            int(gs.Count.Value),
//			LogicalDeviceId:  goapstra.ObjectId(gs.LogicalDeviceId.Value),
//			PortChannelIdMin: 0,
//			PortChannelIdMax: 0,
//			Links:            gs.Links.toGoapstraRackLinkRequests(),
//			Tags:             gs.TagLabels.toGoapstraTagLabels(),
//			//AsnDomain:       0, // not exposed in WebUI, so skipping
//			//ManagementLevel: 0, // not exposed in WebUI, so skipping
//			//Loopback:        0, // not exposed in WebUI, so skipping
//		}
//	}
//	return result
//}

// findDeviceIndexAndTypeByName searches the ResourceRackType for an element
// (leaf_switch/access_switch/generic_system), returns the index of the element
// with matching name, and a rackElementType indicating the type of element.
// Returns (-1, -1) if no match.
func (o *ResourceRackType) findDeviceIndexAndTypeByName(name string, diags *diag.Diagnostics) (int, rackElementType) {
	for i, ls := range o.LeafSwitches {
		if ls.Name.Value == name {
			return i, rackElementTypeLeafSwitch
		}
	}
	// todo restore
	//for i, ls := range o.AccessSwitches {
	//	if ls.Name.Value == name {
	//		return i, rackElementTypeAccessSwitch
	//	}
	//}
	// todo restore
	//for i, ls := range o.GenericSystems {
	//	if ls.Name.Value == name {
	//		return i, rackElementTypeGenericSystem
	//	}
	//}
	diags.AddError("rack element not found",
		fmt.Sprintf("rack element named '%s' was not found in the Rack Type definition returned by Apstra",
			name))
	return -1, -1
}

func (o *ResourceRackType) switchIsRedundant(switchLabel string, diags *diag.Diagnostics) bool {
	// todo restore
	//idx, reType := o.findDeviceIndexAndTypeByName(switchLabel, diags)
	//if diags.HasError() {
	//	return false
	//}

	//switch reType {
	////case rackElementTypeLeafSwitch:
	////	return !o.LeafSwitches[idx].RedundancyProtocol.IsNull() // return true when RP is null
	//	// todo restore
	//	//case rackElementTypeAccessSwitch:
	//	//	return !o.AccessSwitches[idx].RedundancyProtocol.IsNull() // return true when RP is null
	//}

	diags.AddError("cannot determine switch redundancy status",
		fmt.Sprintf("rack type '%s' has no switch with label '%s' ", o.Id.Value, switchLabel))
	return false
}

func (o *ResourceRackType) detectDeviceNameCollisions(diags *diag.Diagnostics) {
	// map keyed by string to detect device name collisions within a rack type
	// todo restore
	//rackDeviceNames := make(map[string]struct{}, len(o.LeafSwitches)+len(o.AccessSwitches)+len(o.GenericSystems))
	//rackDeviceNames := make(map[string]struct{}, len(o.LeafSwitches))

	//collisionDetected := func(name string) {
	//	diags.AddError("rack type device name conflict",
	//		fmt.Sprintf("multiple devices use the name '%s'", name))
	//}

	//for _, d := range o.LeafSwitches {
	//	if _, found := rackDeviceNames[d.Name.Value]; found {
	//		collisionDetected(d.Name.Value)
	//		return
	//	}
	//	rackDeviceNames[d.Name.Value] = struct{}{}
	//}
	// todo restore
	//for _, d := range o.AccessSwitches {
	//	if _, found := rackDeviceNames[d.Name.Value]; found {
	//		collisionDetected(d.Name.Value)
	//		return
	//	}
	//	rackDeviceNames[d.Name.Value] = struct{}{}
	//}
	// todo restore
	//for _, d := range o.GenericSystems {
	//	if _, found := rackDeviceNames[d.Name.Value]; found {
	//		collisionDetected(d.Name.Value)
	//		return
	//	}
	//	rackDeviceNames[d.Name.Value] = struct{}{}
	//}
}

func linkLagModeToGoapstraAttachmentType(lm types.String) goapstra.RackLinkAttachmentType {
	switch lm.Value {
	case goapstra.RackLinkLagModeActive.String():
	case goapstra.RackLinkLagModePassive.String():
	case goapstra.RackLinkLagModeStatic.String():
	default:
		return goapstra.RackLinkAttachmentTypeSingle
	}
	return goapstra.RackLinkAttachmentTypeDual
}

func linkLagModeToGoapstraLagMode(lm types.String) goapstra.RackLinkLagMode {
	switch {
	case lm.Value == goapstra.RackLinkLagModeActive.String():
		return goapstra.RackLinkLagModeActive
	case lm.Value == goapstra.RackLinkLagModePassive.String():
		return goapstra.RackLinkLagModePassive
	case lm.Value == goapstra.RackLinkLagModeStatic.String():
		return goapstra.RackLinkLagModeStatic
	default:
		return goapstra.RackLinkLagModeNone
	}
}

func linkSwitchPeerToGoapstraSwitchPeer(sp types.String) goapstra.RackLinkSwitchPeer {
	switch {
	case !sp.Null && !sp.Unknown && sp.Value == goapstra.RackLinkSwitchPeerFirst.String():
		return goapstra.RackLinkSwitchPeerFirst
	case !sp.Null && !sp.Unknown && sp.Value == goapstra.RackLinkSwitchPeerSecond.String():
		return goapstra.RackLinkSwitchPeerSecond
	default:
		return goapstra.RackLinkSwitchPeerNone
	}
}

func stringerToTfString(in fmt.Stringer) types.String {
	switch in.String() {
	case "":
		return types.String{Null: true}
	}
	return types.String{Value: in.String()}
}

func sliceGoapstraRackLinksToTfRackLinks(in []goapstra.RackLink, diags *diag.Diagnostics) rackLinks {
	if len(in) == 0 {
		return nil
	}
	out := make(rackLinks, len(in))
	for i, inLink := range in {
		out[i] = RackLink{
			Name:               types.String{Value: inLink.Label},
			TargetSwitchLabel:  types.String{Value: inLink.TargetSwitchLabel},
			LagMode:            stringerToTfString(inLink.LagMode),
			LinkPerSwitchCount: types.Int64{Value: int64(inLink.LinkPerSwitchCount)},
			Speed:              types.String{Value: string(inLink.LinkSpeed)},
			SwitchPeer:         stringerToTfString(inLink.SwitchPeer),
			TagLabels:          nil, // copied in later by copyWriteOnlyAttributesFrom()
			TagData:            sliceGoapstraTagDataToSliceTfTagData(inLink.Tags, diags),
		}
	}
	return out
}
