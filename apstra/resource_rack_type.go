package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
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
		MarkdownDescription: "This resource creates an Apstra Rack Type.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Object ID for the Rack Type, assigned by Apstra.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
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
						MarkdownDescription: "Links per spine, minimum value '1'.",
						Type:                types.Int64Type,
						Required:            true,
					},
					"spine_link_speed": {
						MarkdownDescription: "Speed of spine-facing links, something like '10G'",
						Type:                types.StringType,
						Required:            true,
					},
					"redundancy_protocol": {
						MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(leafRedundancyProtocols, "', '")),
						Type:                types.StringType,
						Optional:            true,
						Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
							leafRedundancyRegexp,
							fmt.Sprintf("redundancy_protocol must be one of: '%s'",
								strings.Join(leafRedundancyProtocols, "', '")))},
					},
					//"tags": {
					//	Type:     types.SetType{ElemType: types.StringType},
					//	Optional: true,
					//},
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
				Optional:            true,
				MarkdownDescription: "Template for servers and similar which will be created upon rack instantiation.",
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Name for instances of these Generic Systems.",
						Type:                types.StringType,
						Required:            true,
					},
					"count": {
						MarkdownDescription: "Number of generic systems of this type.",
						Type:                types.Int64Type,
						Required:            true,
					},
					"logical_device_id": {
						MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Generic System.",
						Type:                types.StringType,
						Required:            true,
					},
					"port_channel_id_min": {
						MarkdownDescription: "Port Channel ID Min.",
						Type:                types.Int64Type,
						Optional:            true,
					},
					"port_channel_id_max": {
						MarkdownDescription: "Port Channel ID Max.",
						Type:                types.Int64Type,
						Optional:            true,
					},
					//"tags": {
					//	Type:     types.SetType{ElemType: types.StringType},
					//	Optional: true,
					//},
					"links": {
						MarkdownDescription: "Describe links between the generic system and other systems within the rack",
						Required:            true,
						Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
							"name": {
								MarkdownDescription: "Link name",
								Required:            true,
								Type:                types.StringType,
							},
							"target_switch_name": {
								MarkdownDescription: "Name of the Leaf or Access switch to which the Generic System connects.",
								Required:            true,
								Type:                types.StringType,
							},
							"lag_mode": {
								MarkdownDescription: fmt.Sprintf("LAG Mode must be one of: '%s'.", strings.Join(gsLinkLagModes, "', '")),
								Optional:            true,
								Type:                types.StringType,
								Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
									gsLinkLagModeRegexp,
									fmt.Sprintf("lag_mode must be one of: '%s'",
										strings.Join(gsLinkLagModes, "', '")))},
							},
							"links_per_switch": {
								MarkdownDescription: "Default value '1'.",
								Required:            true,
								Type:                types.Int64Type,
							},
							"speed": {
								MarkdownDescription: "Link Speed, something like '10G'",
								Required:            true,
								Type:                types.StringType,
							},
							//"tags": {
							//	Type:     types.SetType{ElemType: types.StringType},
							//	Optional: true,
							//},
							"switch_peer": {
								// todo: validate and err if set and lag_mode != none
								MarkdownDescription: fmt.Sprintf("For non-LAG links to redundant switches, must be one of '%s'.", strings.Join(gsLinkSwitchPeers, "', '")),
								Optional:            true,
								Computed:            true,
								Type:                types.StringType,
								Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
									gsLinkSwitchPeerRegexp,
									fmt.Sprintf("link switch_peer must be one of: '%s'",
										strings.Join(gsLinkSwitchPeers, "', '")))},
							},
						}),
					},
				}),
			},
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
	plan := &ResourceRackType{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate plan
	plan.Validate(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compute required elements
	plan.compute(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare a goapstra.RackTypeRequest
	rtReq := goapstra.RackTypeRequest{
		DisplayName:              plan.Name.Value,
		Description:              plan.Description.Value,
		FabricConnectivityDesign: parseFCD(plan.FabricConnectivityDesign),
		LeafSwitches:             plan.parseTfLeafSwitchesToGoapstraLeafSwitchRequests(&diags),
		GenericSystems:           plan.parseTfGenericSystemsToGoapstraGenericSystemsRequests(&diags),
		//AccessSwitches:         plan.parseTfAccessSwitchesToGoapstraAccessSwitchRequests(&diags),
	}

	// one of the parse functions above may have generated an error
	if diags.HasError() {
		return
	}

	// request the rack type from Apstra
	id, err := r.p.client.CreateRackType(ctx, &rtReq)
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

	newState := goApstraRackTypeToResourceRackType(rt, &resp.Diagnostics)

	// copy read-only elements of old state into new state
	copyReadOnlyAttributes(oldState, newState, &resp.Diagnostics)

	//o, _ := json.Marshal(oldState)
	//n, _ := json.Marshal(newState)
	//resp.Diagnostics.AddWarning("o", string(o))
	//resp.Diagnostics.AddWarning("n", string(n))

	// Set state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceRackType) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
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

	// Validate plan
	plan.Validate(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compute required elements
	plan.compute(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	rtReq := goapstra.RackTypeRequest{
		DisplayName:              plan.Name.Value,
		Description:              plan.Description.Value,
		FabricConnectivityDesign: parseFCD(plan.FabricConnectivityDesign),
		LeafSwitches:             plan.parseTfLeafSwitchesToGoapstraLeafSwitchRequests(&diags),
		GenericSystems:           plan.parseTfGenericSystemsToGoapstraGenericSystemsRequests(&diags),
		//AccessSwitches:         plan.parseTfAccessSwitchesToGoapstraAccessSwitchRequests(&diags),
	}

	// one of the parse functions above may have generated an error
	if diags.HasError() {
		return
	}

	err := r.p.client.UpdateRackType(ctx, goapstra.ObjectId(state.Id.Value), &rtReq)
	if err != nil {
		resp.Diagnostics.AddError("error updating rack type", err.Error())
		return
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
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

// copyReadOnlyAttributes duplicates user input (from the saved state) into the
// new state object instantiated during Read() operations. Currently, those
// elements are:
// - leaf switch logical device IDs (b/c the LDID returned in the rack type
//   object does not relate to the create-time LDID found in the global catalog)
// - generic system logical device IDs (b/c the LDID returned in the rack type
//   object does not relate to the create-time LDID found in the global catalog)
func copyReadOnlyAttributes(oldState *ResourceRackType, fromApstra *ResourceRackType, diags *diag.Diagnostics) {
	// duplicate the logical device ID from the state file into the object returned by goapstra
	for i, oldLeafSwitch := range oldState.LeafSwitches {
		idx := fromApstra.getLeafSwitchIndexByName(oldLeafSwitch.Name.Value)
		if idx >= 0 {
			fromApstra.LeafSwitches[i].LogicalDeviceId = types.String{Value: oldLeafSwitch.LogicalDeviceId.Value}
		}
	}

	// duplicate the logical device ID from the state file into the object returned by goapstra
	for i, oldGenericSystem := range oldState.GenericSystems {
		idx := fromApstra.getGenericSystemIndexByName(oldGenericSystem.Name.Value)
		if idx >= 0 {
			fromApstra.GenericSystems[i].LogicalDeviceId = types.String{Value: oldGenericSystem.LogicalDeviceId.Value}
		}
	}
}

func goapstraLeafSwitchesToTfLeafSwitches(leafs []goapstra.RackElementLeafSwitch, diag *diag.Diagnostics) []LeafSwitch {
	// return a nil slice rather than zero-length slice to prevent state churn
	if len(leafs) == 0 {
		return nil
	}

	result := make([]LeafSwitch, len(leafs))
	for i, leaf := range leafs {
		result[i] = LeafSwitch{
			Name:               types.String{Value: leaf.Label},
			LogicalDeviceId:    types.String{Unknown: true}, // this value cannot be polled from the API
			LinkPerSpineCount:  types.Int64{Value: int64(leaf.LinkPerSpineCount)},
			LinkPerSpineSpeed:  types.String{Value: string(leaf.LinkPerSpineSpeed)},
			RedundancyProtocol: readLeafRedundancyProtocol(leaf.RedundancyProtocol),
			//Tags:             goapstraDesignTagsToTfTagLabels(leaf.Tags),
		}
	}
	return result
}

func readLeafRedundancyProtocol(in goapstra.LeafRedundancyProtocol) types.String {
	if in == goapstra.LeafRedundancyProtocolNone {
		return types.String{Null: true}
	}
	return types.String{Value: in.String()}
}

func goapstraGenericSystemsToTfGenericSystems(genericSystems []goapstra.RackElementGenericSystem, diags *diag.Diagnostics) []GenericSystem {
	// return a nil slice rather than zero-length slice to prevent state churn
	if len(genericSystems) == 0 {
		return nil
	}

	result := make([]GenericSystem, len(genericSystems))
	for i, genericSystem := range genericSystems {
		var portChannelIdMin, portChannelIdMax types.Int64

		if genericSystem.PortChannelIdMin == 0 {
			portChannelIdMin = types.Int64{Null: true}
		} else {
			portChannelIdMin = types.Int64{Value: int64(genericSystem.PortChannelIdMin)}
		}

		if genericSystem.PortChannelIdMax == 0 {
			portChannelIdMax = types.Int64{Null: true}
		} else {
			portChannelIdMax = types.Int64{Value: int64(genericSystem.PortChannelIdMax)}
		}

		result[i] = GenericSystem{
			Name:             types.String{Value: genericSystem.Label},
			Count:            types.Int64{Value: int64(genericSystem.Count)},
			LogicalDeviceId:  types.String{Value: string(genericSystem.LogicalDeviceId)},
			PortChannelIdMin: portChannelIdMin,
			PortChannelIdMax: portChannelIdMax,
			Links:            goApstraGenericSystemRackLinksToTfGSLinks(genericSystem.Links, diags),
			//Tags:           goapstraDesignTagsToTfTagLabels(genericSystem.Tags),
		}
	}
	return result
}

func goApstraGenericSystemRackLinksToTfGSLinks(in []goapstra.RackLink, diags *diag.Diagnostics) []GSLink {
	if len(in) == 0 {
		return nil
	}
	out := make([]GSLink, len(in))
	for i, link := range in {
		// don't blindly store lagMode: no lag means user omitted it.
		var lagMode types.String
		if link.LagMode == goapstra.RackLinkLagModeNone {
			lagMode = types.String{Null: true}
		} else {
			lagMode = types.String{Value: link.LagMode.String()}
		}

		var switchPeer types.String
		if link.SwitchPeer == goapstra.RackLinkSwitchPeerNone {
			switchPeer = types.String{Null: true}
		} else {
			switchPeer = types.String{Value: link.SwitchPeer.String()}
		}

		out[i] = GSLink{
			Name:               types.String{Value: link.Label},
			TargetSwitchLabel:  types.String{Value: link.TargetSwitchLabel},
			LagMode:            lagMode,
			LinkPerSwitchCount: types.Int64{Value: int64(link.LinkPerSwitchCount)},
			Speed:              types.String{Value: string(link.LinkSpeed)},
			SwitchPeer:         switchPeer,
			//Tags:               goapstraDesignTagsToTfTagLabels(link.Tags),
		}
	}
	return out
}

func goApstraRackTypeToResourceRackType(rt *goapstra.RackType, diags *diag.Diagnostics) *ResourceRackType {
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
		LeafSwitches:             goapstraLeafSwitchesToTfLeafSwitches(rt.LeafSwitches, diags),
		GenericSystems:           goapstraGenericSystemsToTfGenericSystems(rt.GenericSystems, diags),
		//AccessSwitches:         goapstraRtAccessSwitchesToTfAccessSwitches(rt.AccessSwitches, diags), // todo
	}
}

func goapstraDesignTagsToTfTagLabels(in []goapstra.DesignTag) []types.String {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.String, len(in))
	for i, tag := range in {
		out[i] = types.String{Value: string(tag.Label)}
	}
	return out
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
	case !sp.Null && !sp.Unknown && sp.Value == goapstra.RackLinkSwitchPeerFirst.String():
		return goapstra.RackLinkSwitchPeerFirst
	case !sp.Null && !sp.Unknown && sp.Value == goapstra.RackLinkSwitchPeerSecond.String():
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

func (o *ResourceRackType) parseTfLeafSwitchesToGoapstraLeafSwitchRequests(diags *diag.Diagnostics) []goapstra.RackElementLeafSwitchRequest {
	result := make([]goapstra.RackElementLeafSwitchRequest, len(o.LeafSwitches))
	var i int
	for _, leaf := range o.LeafSwitches {
		result[i] = goapstra.RackElementLeafSwitchRequest{
			Label:              leaf.Name.Value,
			LogicalDeviceId:    goapstra.ObjectId(leaf.LogicalDeviceId.Value),
			LinkPerSpineCount:  int(leaf.LinkPerSpineCount.Value),
			LinkPerSpineSpeed:  goapstra.LogicalDevicePortSpeed(leaf.LinkPerSpineSpeed.Value),
			RedundancyProtocol: parseLeafRP(leaf.RedundancyProtocol),
			//Tags:                        parseTfTagsToGoapstraTagLabel(leaf.Tags),
			//LeafLeafL3LinkCount:         int(leaf.LeafLeafL3LinkCount.Value),
			//LeafLeafL3LinkPortChannelId: int(leaf.LeafLeafL3LinkPortChannelId.Value),
			//LeafLeafL3LinkSpeed:         goapstra.LogicalDevicePortSpeed(leaf.LeafLeafL3LinkSpeed.Value),
			//LeafLeafLinkCount:           int(leaf.LeafLeafLinkCount.Value),
			//LeafLeafLinkPortChannelId:   int(leaf.LeafLeafLinkPortChannelId.Value),
			//LeafLeafLinkSpeed:           goapstra.LogicalDevicePortSpeed(leaf.LeafLeafLinkSpeed.Value),
			//MlagVlanId:                  int(leaf.MlagVlanId.Value),
		}
		i++
	}
	return result
}

func parseTfGSLinksToGoapstraRackLinkRequests(plan *ResourceRackType, gsIdx int, diags *diag.Diagnostics) []goapstra.RackLinkRequest {
	links := make([]goapstra.RackLinkRequest, len(plan.GenericSystems[gsIdx].Links))
	for i, link := range plan.GenericSystems[gsIdx].Links {
		var attachmentType goapstra.RackLinkAttachmentType
		if link.LagMode.Value == goapstra.RackLinkLagModeNone.String() {
			attachmentType = goapstra.RackLinkAttachmentTypeSingle
		} else {
			attachmentType = goapstra.RackLinkAttachmentTypeDual
		}
		links[i] = goapstra.RackLinkRequest{
			Label:              link.Name.Value,
			TargetSwitchLabel:  link.TargetSwitchLabel.Value,
			LagMode:            parseGSLagMode(link.LagMode),
			LinkPerSwitchCount: int(link.LinkPerSwitchCount.Value),
			LinkSpeed:          goapstra.LogicalDevicePortSpeed(link.Speed.Value),
			SwitchPeer:         parseGSLinkSwitchPeer(link.SwitchPeer),
			AttachmentType:     attachmentType,
			//Tags:             parseTfTagsToGoapstraTagLabel(link.Tags),
		}
	}
	return links
}

func (o *ResourceRackType) parseTfGenericSystemsToGoapstraGenericSystemsRequests(diags *diag.Diagnostics) []goapstra.RackElementGenericSystemRequest {
	result := make([]goapstra.RackElementGenericSystemRequest, len(o.GenericSystems))
	for i, gs := range o.GenericSystems {
		result[i] = goapstra.RackElementGenericSystemRequest{
			Label:            gs.Name.Value,
			Count:            int(gs.Count.Value),
			LogicalDeviceId:  goapstra.ObjectId(gs.LogicalDeviceId.Value),
			PortChannelIdMin: 0,
			PortChannelIdMax: 0,
			Links:            parseTfGSLinksToGoapstraRackLinkRequests(o, i, diags),
			//Tags:            parseTfTagsToGoapstraTagLabel(gs.Tags),
			//AsnDomain:       0,
			//ManagementLevel: 0,
			//Loopback:        0,
		}
	}
	return result
}

// todo: real validator?
func (o *ResourceRackType) Validate(diags *diag.Diagnostics) {
	for _, gs := range o.GenericSystems {
		for _, link := range gs.Links {
			if !link.LagMode.IsNull() && link.SwitchPeer.Value != "" {
				diags.AddError("incompatible generic system link config",
					"'switch_peer' cannot be set concurrently with 'lag_mode'")
			}
		}
	}
}

// compute fills in missing elements in a plan (*ResourceRackType), including:
// - generic system links to redundant switches have SwitchPeer set null (a link
//   will be made to each MLAG/ESI-LAG member)
// - something else we'll need later, I assume?
func (o *ResourceRackType) compute(diags *diag.Diagnostics) {
	for gsName, gs := range o.GenericSystems {
		for linkNum, link := range gs.Links {
			if !link.LagMode.IsNull() {
				// This link has LAG enabled. SwitchPeer, which selects between MLAG/ESI-LAG
				// switch domain members must be null.
				o.GenericSystems[gsName].Links[linkNum].SwitchPeer = types.String{Null: true}
			} else {
				// This link has LAG disabled....
				switchLagEnabled := o.switchIsRedundant(link.TargetSwitchLabel.Value, diags)
				if diags.HasError() {
					return
				}

				switch {
				case !switchLagEnabled:
					// This link has LAG disabled and the target switch is
					// non-redundant. SwitchPeer must be null because there are
					// no MLAG/ESI-LAG domain members to identify.
					o.GenericSystems[gsName].Links[linkNum].SwitchPeer = types.String{Null: true}
				case switchLagEnabled && link.SwitchPeer.IsUnknown():
					// This link has LAG disabled, the target switch is
					// redundant, and SwitchPeer is unset. Without LAG, we use
					// SwitchPeer to select a single (the first) MLAG/ESI-LAG
					// domain member.
					o.GenericSystems[gsName].Links[linkNum].SwitchPeer = types.String{Value: goapstra.RackLinkSwitchPeerFirst.String()}
				}
			}
		}
	}
}

func (o *ResourceRackType) getLeafSwitchIndexByName(name string) int {
	for i, ls := range o.LeafSwitches {
		if ls.Name.Value == name {
			return i
		}
	}
	return -1
}

func (o *ResourceRackType) getGenericSystemIndexByName(name string) int {
	for i, ls := range o.GenericSystems {
		if ls.Name.Value == name {
			return i
		}
	}
	return -1
}

func (o *ResourceRackType) switchIsRedundant(switchLabel string, diags *diag.Diagnostics) bool {
	leafIdx := o.getLeafSwitchIndexByName(switchLabel)
	switchIsLeaf := leafIdx >= 0

	//accessIdx := r.getAccessSwitchIndexByName(switchLabel) // todo: required for access switch support
	//switchIsAccess := accessIdx >= 0                       // todo: required for access switch support

	//if switchIsLeaf && switchIsAccess {                                                             // todo: required for access switch support
	//	diags.AddError("switch label is not unique",                                                  // todo: required for access switch support
	//		fmt.Sprintf("rack type '%s' has both a leaf switch and an access switch with label '%s'", // todo: required for access switch support
	//			r.Id.Value, switchLabel))                                                             // todo: required for access switch support
	//}                                                                                               // todo: required for access switch support

	switch {
	case switchIsLeaf && o.LeafSwitches[leafIdx].RedundancyProtocol.IsNull():
		return false
	case switchIsLeaf && !o.LeafSwitches[leafIdx].RedundancyProtocol.IsNull():
		return true
		//case linkTargetIsAccess && r.AccessSwitches[accessIdx].RedundancyProtocol.IsNull():  // todo: required for access switch support
		//	return false                                                                       // todo: required for access switch support
		//case linkTargetIsAccess && !r.AccessSwitches[accessIdx].RedundancyProtocol.IsNull(): // todo: required for access switch support
		//	return true                                                                        // todo: required for access switch support
	}

	diags.AddError("no such switch",
		fmt.Sprintf("rack type '%s' has no switch with label '%s' ", o.Id.Value, switchLabel))
	return false
}
