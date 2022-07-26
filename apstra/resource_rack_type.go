package apstra

import (
	"context"
	"errors"
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
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					//"name": {
					//	Type:     types.StringType,
					//	Required: true,
					//},
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
				Optional: true,
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"count": {
						Type:     types.Int64Type,
						Required: true,
					},
					"logical_device_id": {
						Type:     types.StringType,
						Required: true,
						//PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
					},
					"port_channel_id_min": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"port_channel_id_max": {
						Type:     types.Int64Type,
						Optional: true,
					},
					//"tags": {
					//	Type:     types.SetType{ElemType: types.StringType},
					//	Optional: true,
					//},
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
							//"tags": {
							//	Type:     types.SetType{ElemType: types.StringType},
							//	Optional: true,
							//},
							// todo: validate err if set and lag_mode != none
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
	plan.Compute(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	rtReq := goapstra.RackTypeRequest{
		DisplayName:              plan.Name.Value,
		Description:              plan.Description.Value,
		FabricConnectivityDesign: parseFCD(plan.FabricConnectivityDesign),
		LeafSwitches:             parseTfLeafSwitchesToGoapstraLeafSwitchRequests(plan, &diags),
		GenericSystems:           parseTfGenericSystemsToGoapstraGenericSystemsRequests(plan, &diags),
		//AccessSwitches:         parseTfAccessSwitchesToGoapstraAccessSwitchRequests(&plan),
	}

	// one of the parse functions above may have generated an error
	if diags.HasError() {
		return
	}

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

func copyReadOnlyAttributes(savedState *ResourceRackType, fromApstra *ResourceRackType, diags *diag.Diagnostics) {
	// enhance data read from API with LogicalDeviceId found in the state file
	for name, oldLeafSwitch := range savedState.LeafSwitches {
		if leafSwitchPerAPI, found := fromApstra.LeafSwitches[name]; found {
			leafSwitchPerAPI.LogicalDeviceId = types.String{Value: oldLeafSwitch.LogicalDeviceId.Value}
			fromApstra.LeafSwitches[name] = leafSwitchPerAPI
		}
	}

	// enhance data read from API with LogicalDeviceId found in the state file
	for name, oldGenericSystem := range savedState.GenericSystems {
		if genericSystemPerAPI, found := fromApstra.GenericSystems[name]; found {
			genericSystemPerAPI.LogicalDeviceId = types.String{Value: oldGenericSystem.LogicalDeviceId.Value}
			fromApstra.GenericSystems[name] = genericSystemPerAPI
		}
	}
}

func goapstraRtLeafSwitchesToTfLeafSwitches(leafs []goapstra.RackElementLeafSwitch, diag *diag.Diagnostics) map[string]LeafSwitch {
	result := make(map[string]LeafSwitch, len(leafs))
	for _, leaf := range leafs {
		result[leaf.Label] = LeafSwitch{
			LogicalDeviceId:    types.String{Unknown: true}, // this value cannot be polled from the API
			LinkPerSpineCount:  types.Int64{Value: int64(leaf.LinkPerSpineCount)},
			LinkPerSpineSpeed:  types.String{Value: string(leaf.LinkPerSpineSpeed)},
			RedundancyProtocol: readLeafRedundancyProtocol(leaf.RedundancyProtocol),
			//Tags:               goapstraDesignTagsToTfTagLabels(leaf.Tags),
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

func goapstraRtGenericSystemsToTfGenericSystems(genericSystems []goapstra.RackElementGenericSystem, diags *diag.Diagnostics) map[string]GenericSystem {
	if len(genericSystems) == 0 {
		return nil
	}
	result := make(map[string]GenericSystem, len(genericSystems))
	for _, genericSystem := range genericSystems {
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

		result[genericSystem.Label] = GenericSystem{
			Count:            types.Int64{Value: int64(genericSystem.Count)},
			LogicalDeviceId:  types.String{Value: string(genericSystem.LogicalDeviceId)},
			PortChannelIdMin: portChannelIdMin,
			PortChannelIdMax: portChannelIdMax,
			//Tags:             goapstraDesignTagsToTfTagLabels(genericSystem.Tags),
			Links: goApstraGenericSystemRackLinksToTfGSLinks(genericSystem.Links, diags),
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
		LeafSwitches:             goapstraRtLeafSwitchesToTfLeafSwitches(rt.LeafSwitches, diags),
		GenericSystems:           goapstraRtGenericSystemsToTfGenericSystems(rt.GenericSystems, diags),
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

func parseTfLeafSwitchesToGoapstraLeafSwitchRequests(plan *ResourceRackType, diags *diag.Diagnostics) []goapstra.RackElementLeafSwitchRequest {
	result := make([]goapstra.RackElementLeafSwitchRequest, len(plan.LeafSwitches))
	var i int
	for name, leaf := range plan.LeafSwitches {
		result[i] = goapstra.RackElementLeafSwitchRequest{
			Label:              name,
			LogicalDeviceId:    goapstra.ObjectId(leaf.LogicalDeviceId.Value),
			LinkPerSpineCount:  int(leaf.LinkPerSpineCount.Value),
			LinkPerSpineSpeed:  goapstra.LogicalDevicePortSpeed(leaf.LinkPerSpineSpeed.Value),
			RedundancyProtocol: parseLeafRP(leaf.RedundancyProtocol),
			//Tags:               parseTfTagsToGoapstraTagLabel(leaf.Tags),
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

func parseTfGSLinksToGoapstraRackLinkRequests(plan *ResourceRackType, gsName string, diags *diag.Diagnostics) []goapstra.RackLinkRequest {
	links := make([]goapstra.RackLinkRequest, len(plan.GenericSystems[gsName].Links))
	for i, link := range plan.GenericSystems[gsName].Links {
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

func parseTfGenericSystemsToGoapstraGenericSystemsRequests(plan *ResourceRackType, diags *diag.Diagnostics) []goapstra.RackElementGenericSystemRequest {
	result := make([]goapstra.RackElementGenericSystemRequest, len(plan.GenericSystems))
	var i int
	for gsName, gs := range plan.GenericSystems {
		result[i] = goapstra.RackElementGenericSystemRequest{
			Label:            gsName,
			Count:            int(gs.Count.Value),
			LogicalDeviceId:  goapstra.ObjectId(gs.LogicalDeviceId.Value),
			PortChannelIdMin: 0,
			PortChannelIdMax: 0,
			//Tags:             parseTfTagsToGoapstraTagLabel(gs.Tags),
			//Links: links,
			Links: parseTfGSLinksToGoapstraRackLinkRequests(plan, gsName, diags),
			//AsnDomain:        0,
			//ManagementLevel:  0,
			//Loopback:         0,
		}
		i++
	}
	return result
}

// todo: real validator?
func (r ResourceRackType) Validate(diags *diag.Diagnostics) {
	for _, gs := range r.GenericSystems {
		for _, link := range gs.Links {
			if !link.LagMode.IsNull() && !link.SwitchPeer.IsNull() {
				diags.AddError("incompatible generic system link config",
					"'switch_peer' cannot be set concurrently with 'lag_mode'")
			}
		}
	}
}

func (r *ResourceRackType) Compute(diags *diag.Diagnostics) {
	for gsName, gs := range r.GenericSystems {
		for linkNum, link := range gs.Links {
			if !link.LagMode.IsNull() {
				// with LAG enabled, switch_peer must be empty
				r.GenericSystems[gsName].Links[linkNum].SwitchPeer = types.String{Null: true}
			} else {
				// LAG is disabled.

				// set booleans to indicate link target switch type
				_, linkTargetLeaf := r.LeafSwitches[link.TargetSwitchLabel.Value]
				//_, linkTargetAccess := r.AccessSwitches[link.TargetSwitchLabel.Value]

				// flag whether the target switch (whichever link/access) has mlag/esi capability
				var targetSwitchRedundant bool
				switch {
				case linkTargetLeaf:
					if !r.LeafSwitches[link.TargetSwitchLabel.Value].RedundancyProtocol.IsNull() {
						targetSwitchRedundant = true
					}
				//case link_to_access:
				//	if !r.AccessSwitches[link.TargetSwitchLabel.Value].RedundancyProtocol.IsNull() {
				//		targetSwitchRedundant = true
				//	}
				default:
					diags.AddError("no such switch",
						fmt.Sprintf("generic system '%s' link %d calls for unknown switch '%s'", gsName, linkNum, link.TargetSwitchLabel.Value))
					return
				}

				// where switches come in pairs (redundant esi/mlag capability),
				// non-lag links must specify which (first/second) switch they
				// intend to target
				if targetSwitchRedundant && link.SwitchPeer.IsUnknown() {
					r.GenericSystems[gsName].Links[linkNum].SwitchPeer = types.String{Value: goapstra.RackLinkSwitchPeerFirst.String()}
				} else {
					r.GenericSystems[gsName].Links[linkNum].SwitchPeer = types.String{Null: true}
				}
			}

		}
	}
}
