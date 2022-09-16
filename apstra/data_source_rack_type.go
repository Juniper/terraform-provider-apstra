package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceRackTypeType struct{}

func (r dataSourceRackTypeType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
					"display_name": {
						MarkdownDescription: "Name copied from the Logical Device upon which this Leaf Switch was modeled.",
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
					"panels": {
						MarkdownDescription: "Details physical layout of interfaces on the device.",
						Computed:            true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"rows": {
								MarkdownDescription: "Physical vertical dimension of the panel.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"columns": {
								MarkdownDescription: "Physical horizontal dimension of the panel.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"port_groups": {
								MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
								Computed:            true,
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"port_count": {
										MarkdownDescription: "Number of ports in the group.",
										Computed:            true,
										Type:                types.Int64Type,
									},
									"port_speed_gbps": {
										MarkdownDescription: "Port speed in Gbps.",
										Computed:            true,
										Type:                types.Int64Type,
									},
									"port_roles": {
										MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
										Computed:            true,
										Type:                types.SetType{ElemType: types.StringType},
									},
								}),
							},
						}),
					},
					"tags": {
						MarkdownDescription: "Details any tags applied to the Leaf Switch",
						Computed:            true,
						Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
							"label": {
								MarkdownDescription: "Tag label (name) field.",
								Computed:            true,
								Type:                types.StringType,
							},
							"description": {
								MarkdownDescription: "Tag description field.",
								Computed:            true,
								Type:                types.StringType,
							},
						}),
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
							"peer_links": {
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
							"l3_peer_links": {
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
					"display_name": {
						MarkdownDescription: "Name copied from the Logical Device upon which this Leaf Switch was modeled.",
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
					"links": {
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
							"tags": {
								MarkdownDescription: "Details any tags applied to the Link",
								Computed:            true,
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"label": {
										MarkdownDescription: "Tag label (name) field.",
										Computed:            true,
										Type:                types.StringType,
									},
									"description": {
										MarkdownDescription: "Tag description field.",
										Computed:            true,
										Type:                types.StringType,
									},
								}),
							},
						}),
					},
					"panels": {
						MarkdownDescription: "Details physical layout of interfaces on the device.",
						Computed:            true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"rows": {
								MarkdownDescription: "Physical vertical dimension of the panel.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"columns": {
								MarkdownDescription: "Physical horizontal dimension of the panel.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"port_groups": {
								MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
								Computed:            true,
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"port_count": {
										MarkdownDescription: "Number of ports in the group.",
										Computed:            true,
										Type:                types.Int64Type,
									},
									"port_speed_gbps": {
										MarkdownDescription: "Port speed in Gbps.",
										Computed:            true,
										Type:                types.Int64Type,
									},
									"port_roles": {
										MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
										Computed:            true,
										Type:                types.SetType{ElemType: types.StringType},
									},
								}),
							},
						}),
					},
					"tags": {
						MarkdownDescription: "Details any tags applied to the Access Switch",
						Computed:            true,
						Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
							"label": {
								MarkdownDescription: "Tag label (name) field.",
								Computed:            true,
								Type:                types.StringType,
							},
							"description": {
								MarkdownDescription: "Tag description field.",
								Computed:            true,
								Type:                types.StringType,
							},
						}),
					},
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
					"display_name": {
						MarkdownDescription: "Name copied from the Logical Device upon which this Generic System was modeled.",
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
					"panels": {
						MarkdownDescription: "Details physical layout of interfaces on the device.",
						Computed:            true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"rows": {
								MarkdownDescription: "Physical vertical dimension of the panel.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"columns": {
								MarkdownDescription: "Physical horizontal dimension of the panel.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"port_groups": {
								MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
								Computed:            true,
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"port_count": {
										MarkdownDescription: "Number of ports in the group.",
										Computed:            true,
										Type:                types.Int64Type,
									},
									"port_speed_gbps": {
										MarkdownDescription: "Port speed in Gbps.",
										Computed:            true,
										Type:                types.Int64Type,
									},
									"port_roles": {
										MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
										Computed:            true,
										Type:                types.SetType{ElemType: types.StringType},
									},
								}),
							},
						}),
					},
					"tags": {
						MarkdownDescription: "Details any tags applied to the Generic System",
						Computed:            true,
						Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
							"label": {
								MarkdownDescription: "Tag label (name) field.",
								Computed:            true,
								Type:                types.StringType,
							},
							"description": {
								MarkdownDescription: "Tag description field.",
								Computed:            true,
								Type:                types.StringType,
							},
						}),
					},
					"links": {
						MarkdownDescription: "Details links from this Generic System to switches in this Rack Type.",
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
							"tags": {
								MarkdownDescription: "Details any tags applied to the Link",
								Computed:            true,
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"label": {
										MarkdownDescription: "Tag label (name) field.",
										Computed:            true,
										Type:                types.StringType,
									},
									"description": {
										MarkdownDescription: "Tag description field.",
										Computed:            true,
										Type:                types.StringType,
									},
								}),
							},
						}),
					},
				}),
			},
		},
	}, nil
}

func (r dataSourceRackTypeType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceRackType{
		p: *(p.(*apstraProvider)),
	}, nil
}

type dataSourceRackType struct {
	p apstraProvider
}

func (r dataSourceRackType) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataRackType
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var rt *goapstra.RackType
	if config.Name.Null == false {
		rt, err = r.p.client.GetRackTypeByName(ctx, config.Name.Value)
	}
	if config.Id.Null == false {
		rt, err = r.p.client.GetRackType(ctx, goapstra.ObjectId(config.Id.Value))
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Tag", err.Error())
		return
	}

	newState := goApstraRackTypeToDataSourceRackType(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func goApstraRackTypeToDataSourceRackType(rt *goapstra.RackType, diags *diag.Diagnostics) *DataRackType {
	return &DataRackType{
		Id:                       types.String{Value: string(rt.Id)},
		Name:                     types.String{Value: rt.DisplayName},
		Description:              types.String{Value: rt.Description},
		FabricConnectivityDesign: types.String{Value: rt.FabricConnectivityDesign.String()},
		LeafSwitches:             goApstraRackTypeToDSLeafSwitches(rt, diags),
		AccessSwitches:           goApstraRackTypeToDSAccessSwitches(rt, diags),
		GenericSystems:           goApstraRackTypeToDSGenericSystems(rt, diags),
	}
}

func goApstraRackTypeToDSLeafSwitches(rt *goapstra.RackType, diags *diag.Diagnostics) []DSLeafSwitch {
	leafs := make([]DSLeafSwitch, len(rt.LeafSwitches))
	for i, leaf := range rt.LeafSwitches {
		leafs[i] = DSLeafSwitch{
			Name:               types.String{Value: leaf.Label},
			LinkPerSpineCount:  types.Int64{Value: int64(leaf.LinkPerSpineCount)},
			LinkPerSpineSpeed:  types.String{Value: string(leaf.LinkPerSpineSpeed)},
			RedundancyProtocol: types.String{Value: leaf.RedundancyProtocol.String()},
			DisplayName:        types.String{Value: leaf.DisplayName},
			TagData:            sliceGoapstraTagDataToSliceTfTagData(leaf.Tags, diags),
			Panels:             goApstraPanelsToTfPanels(leaf.Panels, diags),
			MlagInfo: &MlagInfo{
				VlanId:                      types.Int64{Value: int64(leaf.MlagVlanId)},
				LeafLeafLinkCount:           types.Int64{Value: int64(leaf.LeafLeafLinkCount)},
				LeafLeafLinkSpeed:           types.String{Value: string(leaf.LeafLeafLinkSpeed)},
				LeafLeafLinkPortChannelId:   types.Int64{Value: int64(leaf.LeafLeafLinkPortChannelId)},
				LeafLeafL3LinkCount:         types.Int64{Value: int64(leaf.LeafLeafL3LinkCount)},
				LeafLeafL3LinkSpeed:         types.String{Value: string(leaf.LeafLeafL3LinkSpeed)},
				LeafLeafL3LinkPortChannelId: types.Int64{Value: int64(leaf.LeafLeafL3LinkPortChannelId)},
			},
		}
	}
	return leafs
}

func goApstraRackTypeToDSAccessSwitches(rt *goapstra.RackType, diags *diag.Diagnostics) []DSAccessSwitch {
	accessSwitches := make([]DSAccessSwitch, len(rt.LeafSwitches))
	for i, accessSwitch := range rt.AccessSwitches {
		accessSwitches[i] = DSAccessSwitch{
			Name:               types.String{Value: accessSwitch.Label},
			DisplayName:        types.String{Value: accessSwitch.DisplayName},
			Count:              types.Int64{Value: int64(accessSwitch.InstanceCount)},
			RedundancyProtocol: types.String{Value: accessSwitch.RedundancyProtocol.String()},
			Links:              goApstraLinksToTfLinks(accessSwitch.Links, diags),
			Panels:             goApstraPanelsToTfPanels(accessSwitch.Panels, diags),
			Tags:               sliceGoapstraTagDataToSliceTfTagData(accessSwitch.Tags, diags),
			EsiLagInfo: &EsiLagInfo{
				AccessAccessLinkCount: types.Int64{Value: int64(accessSwitch.AccessAccessLinkCount)},
				AccessAccessLinkSpeed: types.String{Value: string(accessSwitch.AccessAccessLinkSpeed)},
			},
		}
	}
	return accessSwitches
}

func goApstraRackTypeToDSGenericSystems(rt *goapstra.RackType, diags *diag.Diagnostics) []DSGenericSystem {
	genericSystems := make([]DSGenericSystem, len(rt.GenericSystems))
	for i, genericSystem := range rt.GenericSystems {
		genericSystems[i] = DSGenericSystem{
			Name:             types.String{Value: genericSystem.Label},
			DisplayName:      types.String{Value: genericSystem.DisplayName},
			Count:            types.Int64{Value: int64(genericSystem.Count)},
			PortChannelIdMin: types.Int64{Value: int64(genericSystem.PortChannelIdMin)},
			PortChannelIdMax: types.Int64{Value: int64(genericSystem.PortChannelIdMax)},
			Tags:             sliceGoapstraTagDataToSliceTfTagData(genericSystem.Tags, diags),
			Panels:           goApstraPanelsToTfPanels(genericSystem.Panels, diags),
			Links:            goApstraLinksToTfLinks(genericSystem.Links, diags),
		}
	}
	return genericSystems
}

func goApstraLinksToTfLinks(in []goapstra.RackLink, diags *diag.Diagnostics) []RackLink {
	out := make([]RackLink, len(in))
	for i, link := range in {
		out[i] = RackLink{
			Name:               types.String{Value: link.Label},
			TargetSwitchLabel:  types.String{Value: link.TargetSwitchLabel},
			LagMode:            types.String{Value: link.LagMode.String()},
			LinkPerSwitchCount: types.Int64{Value: int64(link.LinkPerSwitchCount)},
			Speed:              types.String{Value: string(link.LinkSpeed)},
			SwitchPeer:         types.String{Value: link.SwitchPeer.String()},
			TagData:            sliceGoapstraTagDataToSliceTfTagData(link.Tags, diags),
		}
	}
	return out
}

func goApstraPanelsToTfPanels(in []goapstra.LogicalDevicePanel, diags *diag.Diagnostics) []LogicalDevicePanel {
	out := make([]LogicalDevicePanel, len(in))
	for i, panel := range in {
		out[i] = LogicalDevicePanel{
			Rows:       types.Int64{Value: int64(panel.PanelLayout.RowCount)},
			Columns:    types.Int64{Value: int64(panel.PanelLayout.ColumnCount)},
			PortGroups: make([]LogicalDevicePortGroup, len(panel.PortGroups)),
		}
		for j, pg := range panel.PortGroups {
			out[i].PortGroups[j] = LogicalDevicePortGroup{
				Count: types.Int64{Value: int64(pg.Count)},
				Speed: types.Int64{Value: pg.Speed.BitsPerSecond()},
				Roles: sliceStringToSliceTfString(pg.Roles.Strings()),
			}
		}
	}
	return out
}

func sliceGoapstraTagDataToSliceTfTagData(in []goapstra.DesignTag, diags *diag.Diagnostics) []TagData {
	out := make([]TagData, len(in))
	for i, tag := range in {
		out[i] = TagData{
			Label:       types.String{Value: string(tag.Label)},
			Description: types.String{Value: tag.Description},
		}
	}
	return out
}
