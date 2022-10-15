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
					//"panels": {
					//	MarkdownDescription: "Details physical layout of interfaces on the device.",
					//	Computed:            true,
					//	Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					//		"rows": {
					//			MarkdownDescription: "Physical vertical dimension of the panel.",
					//			Computed:            true,
					//			Type:                types.Int64Type,
					//		},
					//		"columns": {
					//			MarkdownDescription: "Physical horizontal dimension of the panel.",
					//			Computed:            true,
					//			Type:                types.Int64Type,
					//		},
					//		"port_groups": {
					//			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
					//			Computed:            true,
					//			Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					//				"port_count": {
					//					MarkdownDescription: "Number of ports in the group.",
					//					Computed:            true,
					//					Type:                types.Int64Type,
					//				},
					//				"port_speed_gbps": {
					//					MarkdownDescription: "Port speed in Gbps.",
					//					Computed:            true,
					//					Type:                types.Int64Type,
					//				},
					//				"port_roles": {
					//					MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
					//					Computed:            true,
					//					Type:                types.SetType{ElemType: types.StringType},
					//				},
					//			}),
					//		},
					//	}),
					//},
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
			//"access_switches": {
			//	MarkdownDescription: "Details of Access Switches in this Rack Type.",
			//	Computed:            true,
			//	Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//		"name": {
			//			MarkdownDescription: "Indicates the role of the switch within the rack, also used for targeting in-rack links.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"display_name": {
			//			MarkdownDescription: "Name copied from the Logical Device upon which this Leaf Switch was modeled.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"count": {
			//			MarkdownDescription: "Count of Access Switches of this type.",
			//			Computed:            true,
			//			Type:                types.Int64Type,
			//		},
			//		"redundancy_protocol": {
			//			MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"esi_lag_info": {
			//			MarkdownDescription: "Interconnect information for Access Switches in ESI-LAG redundancy mode.",
			//			Computed:            true,
			//			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			//				"l3_link_count": {
			//					MarkdownDescription: "Count of L3 links to ESI peer.",
			//					Computed:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"l3_link_speed": {
			//					MarkdownDescription: "Speed of L3 links to ESI peer.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//			}),
			//		},
			//		"links": {
			//			MarkdownDescription: "Details links from this Access Switch to other switches in this Rack Type.",
			//			Computed:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"name": {
			//					MarkdownDescription: "Name of this link.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"target_switch_name": {
			//					MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"lag_mode": {
			//					MarkdownDescription: "LAG negotiation mode of the Link.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"links_per_switch": {
			//					MarkdownDescription: "Number of Links to each switch.",
			//					Computed:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"speed": {
			//					MarkdownDescription: "Speed of this Link.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"switch_peer": {
			//					MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"tags": {
			//					MarkdownDescription: "Details any tags applied to the Link",
			//					Computed:            true,
			//					Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//						"label": {
			//							MarkdownDescription: "Tag label (name) field.",
			//							Computed:            true,
			//							Type:                types.StringType,
			//						},
			//						"description": {
			//							MarkdownDescription: "Tag description field.",
			//							Computed:            true,
			//							Type:                types.StringType,
			//						},
			//					}),
			//				},
			//			}),
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
			//		"tags": {
			//			MarkdownDescription: "Details any tags applied to the Access Switch",
			//			Computed:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"label": {
			//					MarkdownDescription: "Tag label (name) field.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"description": {
			//					MarkdownDescription: "Tag description field.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//			}),
			//		},
			//	}),
			//},
			//"generic_systems": {
			//	MarkdownDescription: "Details Generic Systems found in the Rack Type.",
			//	Computed:            true,
			//	Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//		"name": {
			//			MarkdownDescription: "Indicates the role of the generic system within the rack.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"display_name": {
			//			MarkdownDescription: "Name copied from the Logical Device upon which this Generic System was modeled.",
			//			Computed:            true,
			//			Type:                types.StringType,
			//		},
			//		"count": {
			//			MarkdownDescription: "Number of Generic Systems of this type.",
			//			Computed:            true,
			//			Type:                types.Int64Type,
			//		},
			//		"port_channel_id_min": {
			//			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			//			Computed:            true,
			//			Type:                types.Int64Type,
			//		},
			//		"port_channel_id_max": {
			//			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			//			Computed:            true,
			//			Type:                types.Int64Type,
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
			//		"tags": {
			//			MarkdownDescription: "Details any tags applied to the Generic System",
			//			Computed:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"label": {
			//					MarkdownDescription: "Tag label (name) field.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"description": {
			//					MarkdownDescription: "Tag description field.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//			}),
			//		},
			//		"links": {
			//			MarkdownDescription: "Details links from this Generic System to switches in this Rack Type.",
			//			Computed:            true,
			//			Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//				"name": {
			//					MarkdownDescription: "Name of this link.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"target_switch_name": {
			//					MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"lag_mode": {
			//					MarkdownDescription: "LAG negotiation mode of the Link.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"links_per_switch": {
			//					MarkdownDescription: "Number of Links to each switch.",
			//					Computed:            true,
			//					Type:                types.Int64Type,
			//				},
			//				"speed": {
			//					MarkdownDescription: "Speed of this Link.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"switch_peer": {
			//					MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
			//					Computed:            true,
			//					Type:                types.StringType,
			//				},
			//				"tags": {
			//					MarkdownDescription: "Details any tags applied to the Link",
			//					Computed:            true,
			//					Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			//						"label": {
			//							MarkdownDescription: "Tag label (name) field.",
			//							Computed:            true,
			//							Type:                types.StringType,
			//						},
			//						"description": {
			//							MarkdownDescription: "Tag description field.",
			//							Computed:            true,
			//							Type:                types.StringType,
			//						},
			//					}),
			//				},
			//			}),
			//		},
			//	}),
			//},
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

	newState := parseRackType(rt, &resp.Diagnostics)
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

	for _, leaf := range rt.Data.LeafSwitches {
		if leaf.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag && leaf.MlagInfo == nil {
			diags.AddError("leaf switch MLAG Info missing",
				fmt.Sprintf("rack type '%s', leaf switch '%s' has '%s', but EsiLagInfo is nil",
					rt.Id, leaf.Label, leaf.RedundancyProtocol.String()))
		}
		if leaf.LogicalDevice == nil {
			diags.AddError("leaf switch logical device info missing",
				fmt.Sprintf("rack type '%s', leaf switch '%s' logical device is nil",
					rt.Id, leaf.Label))
		}
	}

	for _, access := range rt.Data.AccessSwitches {
		if access.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi && access.EsiLagInfo == nil {
			diags.AddError("access switch ESI LAG Info missing",
				fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
					rt.Id, access.Label, access.RedundancyProtocol.String()))
		}
		if access.LogicalDevice == nil {
			diags.AddError("access switch logical device info missing",
				fmt.Sprintf("rack type '%s', access switch '%s' logical device is nil",
					rt.Id, access.Label))
		}
	}

	for _, generic := range rt.Data.GenericSystems {
		if generic.LogicalDevice == nil {
			diags.AddError("generic system logical device info missing",
				fmt.Sprintf("rack type '%s', generic system '%s' logical device is nil",
					rt.Id, generic.Label))
		}
	}
}

func parseRackType(rt *goapstra.RackType, diags *diag.Diagnostics) *dRackType {
	result := &dRackType{
		Id:                       types.String{Value: string(rt.Id)},
		Name:                     types.String{Value: rt.Data.DisplayName},
		Description:              types.String{Value: rt.Data.Description},
		FabricConnectivityDesign: types.String{Value: rt.Data.FabricConnectivityDesign.String()},
		LeafSwitches:             parseRackTypeLeafSwitches(rt.Data.LeafSwitches, diags),
	}
	//o.AccessSwitches =           goApstraRackTypeToDSAccessSwitches(rt, diags)
	//o.GenericSystems =           goApstraRackTypeToDSGenericSystems(rt, diags)
	return result
}

func leafAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"display_name":        types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"tags": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: tagAttrTypes()}},
		"mlag_info": types.ObjectType{
			AttrTypes: mlagInfoAttrTypes()},
	}
}

func newLeafSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: leafAttrTypes()},
	}
}

func newTagSetFromSliceDesignTagData(tags []goapstra.DesignTagData) types.Set {
	result := newTagSet(len(tags))
	for i, tag := range tags {
		result.Elems[i] = types.Object{
			AttrTypes: tagAttrTypes(),
			Attrs: map[string]attr.Value{
				"label":       types.String{Value: tag.Label},
				"description": types.String{Value: tag.Description},
			},
		}
	}
	return result
}

func newMlagInfoObjFromLeafMlagInfo(in *goapstra.LeafMlagInfo) types.Object {
	if in == nil || (in.LeafLeafLinkCount == 0 && in.LeafLeafL3LinkCount == 0) {
		return types.Object{
			Null:      true,
			AttrTypes: mlagInfoAttrTypes(),
		}
	}

	return types.Object{
		AttrTypes: mlagInfoAttrTypes(),
		Attrs: map[string]attr.Value{
			"mlag_keepalive_vlan":          types.Int64{Value: int64(in.MlagVlanId)},
			"peer_links":                   types.Int64{Value: int64(in.LeafLeafLinkCount)},
			"peer_link_speed":              types.String{Value: string(in.LeafLeafLinkSpeed)},
			"peer_link_port_channel_id":    types.Int64{Value: int64(in.LeafLeafLinkPortChannelId)},
			"l3_peer_links":                types.Int64{Value: int64(in.LeafLeafL3LinkCount)},
			"l3_peer_link_speed":           types.String{Value: string(in.LeafLeafL3LinkSpeed)},
			"l3_peer_link_port_channel_id": types.Int64{Value: int64(in.LeafLeafL3LinkPortChannelId)},
		},
	}
}

func newLeafObjFromRackElementLeafSwitch(rels *goapstra.RackElementLeafSwitch) types.Object {
	var spineLinkSpeed types.String
	if rels.LinkPerSpineCount == 0 || rels.LinkPerSpineSpeed == "" {
		spineLinkSpeed = types.String{Null: true}
	} else {
		spineLinkSpeed = types.String{Value: string(rels.LinkPerSpineSpeed)}
	}

	var redundancyProtocol types.String
	if rels.RedundancyProtocol == goapstra.LeafRedundancyProtocolNone {
		redundancyProtocol = types.String{Null: true}
	} else {
		redundancyProtocol = types.String{Value: rels.RedundancyProtocol.String()}
	}

	return types.Object{
		AttrTypes: leafAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: rels.Label},
			"display_name":        types.String{Value: rels.LogicalDevice.DisplayName},
			"spine_link_count":    types.Int64{Value: int64(rels.LinkPerSpineCount)},
			"spine_link_speed":    spineLinkSpeed,
			"redundancy_protocol": redundancyProtocol,
			"tags":                newTagSetFromSliceDesignTagData(rels.Tags),
			"mlag_info":           newMlagInfoObjFromLeafMlagInfo(rels.MlagInfo),
		},
	}
}

func mlagInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mlag_keepalive_vlan":          types.Int64Type,
		"peer_links":                   types.Int64Type,
		"peer_link_speed":              types.StringType,
		"peer_link_port_channel_id":    types.Int64Type,
		"l3_peer_links":                types.Int64Type,
		"l3_peer_link_speed":           types.StringType,
		"l3_peer_link_port_channel_id": types.Int64Type,
	}
}

func tagAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"label":       types.StringType,
		"description": types.StringType,
	}
}

func newTagSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: tagAttrTypes()},
	}
}

func parseRackTypeLeafSwitches(in []goapstra.RackElementLeafSwitch, diags *diag.Diagnostics) types.Set {
	result := newLeafSet(len(in))
	for i, ls := range in {
		result.Elems[i] = newLeafObjFromRackElementLeafSwitch(&ls)
	}
	return result
}

func goApstraRackTypeToDSLeafSwitches(rt *goapstra.RackType, diags *diag.Diagnostics) []DSLeafSwitch {
	leafs := make([]DSLeafSwitch, len(rt.Data.LeafSwitches))
	for i, leaf := range rt.Data.LeafSwitches {
		leafs[i] = DSLeafSwitch{
			Name:               types.String{Value: leaf.Label},
			LinkPerSpineCount:  types.Int64{Value: int64(leaf.LinkPerSpineCount)},
			LinkPerSpineSpeed:  types.String{Value: string(leaf.LinkPerSpineSpeed)},
			RedundancyProtocol: types.String{Value: leaf.RedundancyProtocol.String()},
			DisplayName:        types.String{Value: leaf.LogicalDevice.DisplayName},
			TagData:            sliceGoapstraTagDataToSliceTypesObject(leaf.Tags, diags),
			Panels:             goApstraPanelsToTfPanels(leaf.LogicalDevice.Panels, diags),
		}
		if leaf.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag {
			leafs[i].MlagInfo = &MlagInfo{
				VlanId:                      types.Int64{Value: int64(leaf.MlagInfo.MlagVlanId)},
				LeafLeafLinkCount:           types.Int64{Value: int64(leaf.MlagInfo.LeafLeafLinkCount)},
				LeafLeafLinkSpeed:           types.String{Value: string(leaf.MlagInfo.LeafLeafLinkSpeed)},
				LeafLeafLinkPortChannelId:   types.Int64{Value: int64(leaf.MlagInfo.LeafLeafLinkPortChannelId)},
				LeafLeafL3LinkCount:         types.Int64{Value: int64(leaf.MlagInfo.LeafLeafL3LinkCount)},
				LeafLeafL3LinkSpeed:         types.String{Value: string(leaf.MlagInfo.LeafLeafL3LinkSpeed)},
				LeafLeafL3LinkPortChannelId: types.Int64{Value: int64(leaf.MlagInfo.LeafLeafL3LinkPortChannelId)},
			}
		}
	}
	return leafs
}

func goApstraRackTypeToDSAccessSwitches(rt *goapstra.RackType, diags *diag.Diagnostics) []DSAccessSwitch {
	accessSwitches := make([]DSAccessSwitch, len(rt.Data.LeafSwitches))
	for i, accessSwitch := range rt.Data.AccessSwitches {
		accessSwitches[i] = DSAccessSwitch{
			Name:               types.String{Value: accessSwitch.Label},
			DisplayName:        types.String{Value: accessSwitch.LogicalDevice.DisplayName},
			Count:              types.Int64{Value: int64(accessSwitch.InstanceCount)},
			RedundancyProtocol: types.String{Value: accessSwitch.RedundancyProtocol.String()},
			Links:              goApstraLinksToTfLinks(accessSwitch.Links, diags),
			Panels:             goApstraPanelsToTfPanels(accessSwitch.LogicalDevice.Panels, diags),
			Tags:               sliceGoapstraTagDataToSliceTypesObject(accessSwitch.Tags, diags),
		}
		if accessSwitch.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi {
			if accessSwitch.EsiLagInfo == nil {
				diags.AddError("access switch ESI LAG Info missing",
					fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
						rt.Id, accessSwitch.Label, accessSwitch.RedundancyProtocol.String()))
			}
			accessSwitches[i].EsiLagInfo = &EsiLagInfo{
				AccessAccessLinkCount: types.Int64{Value: int64(accessSwitch.EsiLagInfo.AccessAccessLinkCount)},
				AccessAccessLinkSpeed: types.String{Value: string(accessSwitch.EsiLagInfo.AccessAccessLinkSpeed)},
			}
		}
	}
	return accessSwitches
}

func goApstraRackTypeToDSGenericSystems(rt *goapstra.RackType, diags *diag.Diagnostics) []DSGenericSystem {
	genericSystems := make([]DSGenericSystem, len(rt.Data.GenericSystems))
	for i, genericSystem := range rt.Data.GenericSystems {
		genericSystems[i] = DSGenericSystem{
			Name:             types.String{Value: genericSystem.Label},
			DisplayName:      types.String{Value: genericSystem.LogicalDevice.DisplayName},
			Count:            types.Int64{Value: int64(genericSystem.Count)},
			PortChannelIdMin: types.Int64{Value: int64(genericSystem.PortChannelIdMin)},
			PortChannelIdMax: types.Int64{Value: int64(genericSystem.PortChannelIdMax)},
			Tags:             sliceGoapstraTagDataToSliceTypesObject(genericSystem.Tags, diags),
			Panels:           goApstraPanelsToTfPanels(genericSystem.LogicalDevice.Panels, diags),
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
			TagData:            sliceGoapstraTagDataToSliceTypesObject(link.Tags, diags),
		}
	}
	return out
}

func goApstraPanelsToTfPanels(in []goapstra.LogicalDevicePanel, diags *diag.Diagnostics) []LogicalDevicePanel {
	if len(in) == 0 {
		return nil
	}
	out := make([]LogicalDevicePanel, len(in))
	for i, panel := range in {
		_ = panel // todo delete
		out[i] = LogicalDevicePanel{
			// todo restore
			Rows: types.Int64{Value: int64(panel.PanelLayout.RowCount)},
			//Columns: types.Int64{Value: int64(panel.PanelLayout.ColumnCount)},
			//PortGroups: make([]logicalDevicePortGroup, len(panel.PortGroups)),
		}
		//diags.Append(tfsdk.ValueFrom(context.Background(), panel.PanelLayout.RowCount, types.Int64Type, out[i].Rows)...)
		//diags.Append(tfsdk.ValueFrom(context.Background(), panel.PanelLayout.ColumnCount, types.Int64Type, out[i].Columns)...)
		// todo restore
		//for j, pg := range panel.PortGroups {
		//out[i].PortGroups[j] = logicalDevicePortGroup{
		//	Count: types.Int64{Value: int64(pg.Count)},
		//	Speed: types.Int64{Value: pg.Speed.BitsPerSecond()},
		//	Roles: sliceStringToSliceTfString(pg.Roles.Strings()),
		//}
		//}
	}
	return out
}

func sliceGoapstraTagDataToSliceTypesObject(in []goapstra.DesignTagData, diags *diag.Diagnostics) []types.Object {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.Object, len(in))
	for i, tag := range in {
		out[i] = types.Object{
			Attrs: map[string]attr.Value{
				"label":       types.String{Value: string(tag.Label)},
				"description": types.String{Value: string(tag.Description)},
			},
			AttrTypes: map[string]attr.Type{
				"label":       types.StringType,
				"description": types.StringType,
			},
		}
	}
	return out
}

type dRackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Set    `tfsdk:"leaf_switches"`
	//AccessSwitches           []DSAccessSwitch  `tfsdk:"access_switches"`
	//GenericSystems           []DSGenericSystem `tfsdk:"generic_systems"`
}
