package apstra

import (
	"context"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"strings"
)

type resourceRackTypeType struct{}

func (r resourceRackTypeType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	fcpModes := []string{
		goapstra.FabricConnectivityDesignL3Clos.String(),
		goapstra.FabricConnectivityDesignL3Collapsed.String()}
	leafRedundancyProtocols := []string{
		goapstra.LeafRedundancyProtocolEsi.String(),
		//goapstra.LeafRedundancyProtocolMlag.String(), not yet implemented
	}
	gsLinkLagModes := []string{
		goapstra.RackLinkLagModeActive.String(),
		goapstra.RackLinkLagModePassive.String(),
		goapstra.RackLinkLagModeStatic.String(),
	}
	gsLinkSwitchPeers := []string{
		goapstra.RackLinkSwitchPeerNone.String(),
		goapstra.RackLinkSwitchPeerFirst.String(),
		goapstra.RackLinkSwitchPeerSecond.String(),
	}

	fcdRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(fcpModes, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling fabric connectivity design regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	leafRedundancyRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(leafRedundancyProtocols, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling leaf redundancy regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	gsLinkSwitchPeerRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(gsLinkSwitchPeers, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling generic system link switch peer regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	gsLinkLagModeRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(gsLinkLagModes, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling generic system LAG mode regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:       types.StringType,
				Required:   true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"description": {
				Type:     types.StringType,
				Optional: true,
			},
			"fabric_connectivity_design": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
					fcdRegexp,
					fmt.Sprintf("fabric_connectivity_design must be one of: '%s'",
						strings.Join(fcpModes, "', '")))},
			},
			"leaf_switches": {
				Required: true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
					},
					"logical_device_id": {
						Type:     types.StringType,
						Required: true,
					},
					"spine_link_count": {
						Type:     types.Int64Type,
						Required: true,
					},
					"spine_link_speed": {
						Type:     types.StringType,
						Required: true,
					},
					"redundancy_protocol": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
							leafRedundancyRegexp,
							fmt.Sprintf("redundancy_protocol must be one of: '%s'",
								strings.Join(leafRedundancyProtocols, "', '")))},
					},
					"tags": {
						Type:     types.SetType{ElemType: types.StringType},
						Optional: true,
					},
					//"l3_peer_link_count": {
					//	Type:     types.Int64Type,
					//	Optional: true,
					//},
					//"l3_peer_link_speed": {
					//	Type:     types.StringType,
					//	Optional: true,
					//},
					//"l3_peer_link_port_channel_id": {
					//	Type:     types.Int64Type,
					//	Optional: true,
					//},
					//"peer_link_count": {
					//	Type:     types.Int64Type,
					//	Optional: true,
					//},
					//"peer_link_speed": {
					//	Type:     types.StringType,
					//	Optional: true,
					//},
					//"peer_link_port_channel_id": {
					//	Type:     types.Int64Type,
					//	Optional: true,
					//},
					//"mlag_vlan_id": {
					//	Type:     types.Int64Type,
					//	Optional: true,
					//},
				})},
			"generic_systems": {
				Required: true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
					},
					"count": {
						Type:     types.Int64Type,
						Required: true,
					},
					"logical_device_id": {
						Type:     types.StringType,
						Required: true,
					},
					"port_channel_id_min": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"port_channel_id_max": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"tags": {
						Type:     types.SetType{ElemType: types.StringType},
						Optional: true,
					},
					"links": {
						Required: true,
						Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
							"name": {
								Required: true,
								Type:     types.StringType,
							},
							"target_switch_name": {
								Required: true,
								Type:     types.StringType,
							},
							"lag_mode": {
								Optional: true,
								Type:     types.StringType,
								Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
									gsLinkLagModeRegexp,
									fmt.Sprintf("lag_mode must be one of: '%s'",
										strings.Join(gsLinkLagModes, "', '")))},
							},
							"links_per_switch": {
								Required: true,
								Type:     types.Int64Type,
							},
							"speed": {
								Required: true,
								Type:     types.StringType,
							},
							"tags": {
								Type:     types.SetType{ElemType: types.StringType},
								Optional: true,
							},
							// todo: validate incompatible with lag_mode != none
							"switch_peer": {
								Optional: true,
								Computed: true,
								Type:     types.StringType,
								Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
									gsLinkSwitchPeerRegexp,
									fmt.Sprintf("link switch_peer must be one of: '%s'",
										strings.Join(gsLinkSwitchPeers, "', '")))},
							},
						}),
					},
				})},
		},
	}, nil
}

func (r resourceRackTypeType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceRackType{
		p: *(p.(*provider)),
	}, nil
}

type resourceRackType struct {
	p provider
}

func (r resourceRackType) ValidateConfig(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	var cfg ResourceRackType
	req.Config.Get(ctx, &cfg)

	if len(cfg.LeafSwitches) == 0 {
		resp.Diagnostics.AddError(
			"missing required configuration element",
			"at least one 'leaf_switches' element is required")
	}
}

func (r resourceRackType) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceRackType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.p.client.CreateRackType(ctx,
		&goapstra.RackTypeRequest{
			DisplayName:              plan.Name.Value,
			Description:              plan.Description.Value,
			FabricConnectivityDesign: parseFCD(plan.FabricConnectivityDesign),
			LeafSwitches:             parseTfLeafSwitchesToGoapstraLeafSwitchRequests(&plan),
			GenericSystems:           parseTfGenericSystemsToGoapstraGenericSystemsRequests(&plan),
			//AccessSwitches:         parseTfAccessSwitchesToGoapstraAccessSwitchRequests(&plan),
		})
	if err != nil {
		resp.Diagnostics.AddError("error creating rack type", err.Error())
		return
	}
	plan.Id = types.String{Value: string(id)}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRackType) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ResourceRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceRackType) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get current state
	var state ResourceRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan ResourceRackType
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceRackType) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
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

func parseFCD(fcd types.String) goapstra.FabricConnectivityDesign {
	switch {
	case fcd.Value == goapstra.FabricConnectivityDesignL3Collapsed.String():
		return goapstra.FabricConnectivityDesignL3Collapsed
	default:
		return goapstra.FabricConnectivityDesignL3Clos
	}
}

func parseLeafRP(rp types.String) goapstra.LeafRedundancyProtocol {
	switch {
	case rp.Value == goapstra.LeafRedundancyProtocolEsi.String():
		return goapstra.LeafRedundancyProtocolEsi
	case rp.Value == goapstra.LeafRedundancyProtocolMlag.String():
		return goapstra.LeafRedundancyProtocolMlag
	default:
		return goapstra.LeafRedundancyProtocolNone
	}
}

func parseGSLagMode(lm types.String) goapstra.RackLinkLagMode {
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

func parseGSLinkSwitchPeer(sp types.String) goapstra.RackLinkSwitchPeer {
	switch {
	case sp.Value == goapstra.RackLinkSwitchPeerFirst.String():
		return goapstra.RackLinkSwitchPeerFirst
	case sp.Value == goapstra.RackLinkSwitchPeerSecond.String():
		return goapstra.RackLinkSwitchPeerSecond
	default:
		return goapstra.RackLinkSwitchPeerNone
	}
}

func parseTfTagsToGoapstraTagLabel(in []types.String) []goapstra.TagLabel {
	result := make([]goapstra.TagLabel, len(in))
	for i, t := range in {
		result[i] = goapstra.TagLabel(t.Value)
	}
	return result
}

func parseTfLeafSwitchesToGoapstraLeafSwitchRequests(plan *ResourceRackType) []goapstra.RackElementLeafSwitchRequest {
	result := make([]goapstra.RackElementLeafSwitchRequest, len(plan.LeafSwitches))
	for i, leaf := range plan.LeafSwitches {
		result[i] = goapstra.RackElementLeafSwitchRequest{
			Label:              leaf.Name.Value,
			LogicalDeviceId:    goapstra.ObjectId(leaf.LogicalDeviceId.Value),
			LinkPerSpineCount:  int(leaf.LinkPerSpineCount.Value),
			LinkPerSpineSpeed:  goapstra.LogicalDevicePortSpeed(leaf.LinkPerSpineSpeed.Value),
			RedundancyProtocol: parseLeafRP(leaf.RedundancyProtocol),
			Tags:               parseTfTagsToGoapstraTagLabel(leaf.Tags),
			//LeafLeafL3LinkCount:         int(leaf.LeafLeafL3LinkCount.Value),
			//LeafLeafL3LinkPortChannelId: int(leaf.LeafLeafL3LinkPortChannelId.Value),
			//LeafLeafL3LinkSpeed:         goapstra.LogicalDevicePortSpeed(leaf.LeafLeafL3LinkSpeed.Value),
			//LeafLeafLinkCount:           int(leaf.LeafLeafLinkCount.Value),
			//LeafLeafLinkPortChannelId:   int(leaf.LeafLeafLinkPortChannelId.Value),
			//LeafLeafLinkSpeed:           goapstra.LogicalDevicePortSpeed(leaf.LeafLeafLinkSpeed.Value),
			//MlagVlanId:                  int(leaf.MlagVlanId.Value),
		}
	}
	return result
}

func parseTfAccessSwitchesToGoapstraAccessSwitchRequests(tfAccessSwitches []AccessSwitch) []goapstra.RackElementAccessSwitchRequest {
	return nil
	//result := make([]goapstra.RackElementAccessSwitchRequest, len(tfAccessSwitches))
	//for i, leaf := range tfAccessSwitches {
	//	result[i] = goapstra.RackElementAccessSwitchRequest{
	//		InstanceCount:         0,
	//		RedundancyProtocol:    0,
	//		Links:                 nil,
	//		Label:                 "",
	//		Panels:                nil,
	//		DisplayName:           "",
	//		LogicalDeviceId:       "",
	//		Tags:                  nil,
	//		AccessAccessLinkCount: 0,
	//		AccessAccessLinkSpeed: "",
	//	}
	//}
	//return result
}

func parseTfGenericSystemsToGoapstraGenericSystemsRequests(plan *ResourceRackType) []goapstra.RackElementGenericSystemRequest {
	result := make([]goapstra.RackElementGenericSystemRequest, len(plan.GenericSystems))
	for i, gs := range plan.GenericSystems {
		links := make([]goapstra.RackLinkRequest, len(gs.Links))
		for j, link := range gs.Links {
			var attachmentType goapstra.RackLinkAttachmentType
			if link.LagMode.Value == goapstra.RackLinkLagModeNone.String() {
				attachmentType = goapstra.RackLinkAttachmentTypeSingle
			} else {
				plan.GenericSystems[i].Links[j].SwitchPeer = types.String{Value: goapstra.RackLinkSwitchPeerNone.String()}
				attachmentType = goapstra.RackLinkAttachmentTypeDual
			}
			links[j] = goapstra.RackLinkRequest{
				Label:              link.Name.Value,
				TargetSwitchLabel:  link.TargetSwitchLabel.Value,
				LagMode:            parseGSLagMode(link.LagMode),
				LinkPerSwitchCount: int(link.LinkPerSwitchCount.Value),
				LinkSpeed:          goapstra.LogicalDevicePortSpeed(link.Speed.Value),
				Tags:               parseTfTagsToGoapstraTagLabel(link.Tags),
				SwitchPeer:         parseGSLinkSwitchPeer(link.SwitchPeer),
				AttachmentType:     attachmentType,
			}
		}
		result[i] = goapstra.RackElementGenericSystemRequest{
			Label:            gs.Name.Value,
			Count:            int(gs.Count.Value),
			LogicalDeviceId:  goapstra.ObjectId(gs.LogicalDeviceId.Value),
			PortChannelIdMin: 0,
			PortChannelIdMax: 0,
			Tags:             parseTfTagsToGoapstraTagLabel(gs.Tags),
			Links:            links,
			//AsnDomain:        0,
			//ManagementLevel:  0,
			//Loopback:         0,
		}

	}
	return result
}
