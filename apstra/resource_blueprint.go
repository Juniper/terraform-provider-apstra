package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type resourceBlueprintType struct{}

func (r resourceBlueprintType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				//PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"template_id": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"leaf_asn_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"leaf_ip_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"link_ip_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"spine_asn_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"spine_ip_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"switches": {
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"interface_map": {
						Type:          types.StringType,
						Optional:      true,
						Computed:      true,
						PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
					},
					"device_key": {
						Type:     types.StringType,
						Required: true,
					},
					"device_profile": {
						Type:     types.StringType,
						Computed: true,
					},
					"system_node_id": {
						Type:          types.StringType,
						Computed:      true,
						PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
					},
				}),
				Optional: true,
			},
		},
	}, nil
}

func (r resourceBlueprintType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceBlueprint{
		p: *(p.(*apstraProvider)),
	}, nil
}

type resourceBlueprint struct {
	p apstraProvider
}

func (r resourceBlueprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceBlueprint
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ensure ASN pools from the plan exist on Apstra
	missing, err := asnPoolsExist(ctx, append(plan.LeafAsns, plan.SpineAsns...), r.p.client)
	if err != nil {
		resp.Diagnostics.AddError("error fetching available ASN pools", err.Error())
		return
	}
	if missing != "" {
		resp.Diagnostics.AddError("cannot assign ASN pool",
			fmt.Sprintf("requested pool '%s' does not exist", missing))
		return
	}

	// ensure IP4 pools from the plan exist on Apstra
	missing, err = ip4PoolsExist(ctx, append(append(plan.LeafIp4s, plan.SpineIp4s...), plan.LinkIp4s...), r.p.client)
	if err != nil {
		resp.Diagnostics.AddError("error fetching available IP pools", err.Error())
		return
	}
	if missing != "" {
		resp.Diagnostics.AddError("cannot assign IP pool",
			fmt.Sprintf("requested pool '%s' does not exist", missing))
		return
	}

	// ensure switches (by device key) exist on Apstra
	asi, err := r.p.client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
	if err != nil {
		resp.Diagnostics.AddError("get managed system info", err.Error())
		return
	}
	deviceKeyToSystemInfo := make(map[string]*goapstra.ManagedSystemInfo) // map-ify the Apstra output
	for _, si := range asi {
		deviceKeyToSystemInfo[si.DeviceKey] = &si
	}
	// check each planned switch exists in Apstra, and save the aos_hcl_model (device profile)
	for switchLabel, switchFromPlan := range plan.Switches {
		if si, found := deviceKeyToSystemInfo[switchFromPlan.DeviceKey.Value]; !found {
			resp.Diagnostics.AddError("switch not found",
				fmt.Sprintf("no switch with device_key '%s' exists on Apstra", switchFromPlan.DeviceKey.Value))
			return
		} else {
			switchFromPlan.DeviceProfile = types.String{Value: si.Facts.AosHclModel}
			plan.Switches[switchLabel] = switchFromPlan
		}
	}

	// create blueprint
	blueprintId, err := r.p.client.CreateBlueprintFromTemplate(ctx, &goapstra.CreateBlueprintFromTemplate{
		RefDesign:  goapstra.RefDesignDatacenter,
		Label:      plan.Name.Value,
		TemplateId: goapstra.ObjectId(plan.TemplateId.Value),
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating Blueprint", err.Error())
		return
	}
	plan.Id = types.String{Value: string(blueprintId)}

	// create a client specific to the reference design
	refDesignClient, err := r.p.client.NewTwoStageL3ClosClient(ctx, blueprintId)

	// warn about switches discovered in the graph db, and which do not appear in the tf config
	err = warnAboutSwitchesMissingFromPlan(ctx, r.p.client, blueprintId, plan.Switches, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("error while inventorying switches after blueprint creation", err.Error())
		return
	}

	// assign leaf ASN pool
	err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeAsnPool,
		Name:    goapstra.ResourceGroupNameLeafAsn,
		PoolIds: tfStringSliceToSliceObjectId(plan.LeafAsns),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign leaf IP4 pool
	err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeIp4Pool,
		Name:    goapstra.ResourceGroupNameLeafIps,
		PoolIds: tfStringSliceToSliceObjectId(plan.LeafIp4s),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign link IP4 pool
	err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeIp4Pool,
		Name:    goapstra.ResourceGroupNameLinkIps,
		PoolIds: tfStringSliceToSliceObjectId(plan.LinkIp4s),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign spine ASN pool
	err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeAsnPool,
		Name:    goapstra.ResourceGroupNameSpineAsn,
		PoolIds: tfStringSliceToSliceObjectId(plan.SpineAsns),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign spine IP4 pool
	err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeIp4Pool,
		Name:    goapstra.ResourceGroupNameSpineIps,
		PoolIds: tfStringSliceToSliceObjectId(plan.SpineIp4s),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// structure we'll use when assigning interface maps to switches
	ifmapAssignments := make(goapstra.SystemIdToInterfaceMapAssignment)

	// assign details of each configured switch (don't add elements to the plan.Switches map)
	//	- DeviceKey : required user input
	//	- InterfaceMap : optional user input - if only one option, we'll auto-assign
	//	- DeviceProfile : a.k.a. aos_hcl_model - determined from InterfaceMap, represents physical device/model
	//	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a spine/leaf/etc...
	for switchLabel, switchPlan := range plan.Switches {
		// fetch the switch graph db node ID and candidate interface maps
		systemNodeId, ifmapCandidates, err := getSystemNodeIdAndIfmapCandidates(ctx, r.p.client, blueprintId, switchLabel)
		if err != nil {
			resp.Diagnostics.AddWarning("error fetching interface map candidates", err.Error())
			continue
		}

		// save the SystemNodeId (1:1 relationship with switchLabel in graph db)
		switchPlan.SystemNodeId = types.String{Value: systemNodeId}

		// validate/choose interface map, build ifmap assignment structure
		if !switchPlan.InterfaceMap.Null && !switchPlan.InterfaceMap.Unknown && !(switchPlan.InterfaceMap.Value == "") {
			// user gave us an interface map label they'd like to use
			ifmapNodeId := ifmapCandidateFromCandidates(switchPlan.InterfaceMap.Value, ifmapCandidates)
			if ifmapNodeId != nil {
				ifmapAssignments[systemNodeId] = ifmapNodeId.id
				switchPlan.DeviceProfile = types.String{Value: ifmapNodeId.deviceProfileId}
			} else {
				resp.Diagnostics.AddWarning(
					"invalid interface map",
					fmt.Sprintf("interface map '%s' not found among candidates for node '%s'",
						switchPlan.InterfaceMap.Value, switchLabel))
			}
		} else {
			// user didn't give us an interface map label; try to find a default
			switch len(ifmapCandidates) {
			case 0: // no candidates!
				resp.Diagnostics.AddWarning(
					"interface map not specified, and no candidates found",
					fmt.Sprintf("no candidate interface maps found for node '%s'", switchLabel))
			case 1: // exact match; we can work with this
				ifmapAssignments[systemNodeId] = ifmapCandidates[0].id
				switchPlan.InterfaceMap = types.String{Value: ifmapCandidates[0].label}
				switchPlan.DeviceProfile = types.String{Value: ifmapCandidates[0].deviceProfileId}
			default: // multiple match!
				sb := strings.Builder{}
				sb.WriteString(fmt.Sprintf("'%s'", ifmapCandidates[0].label))
				for _, candidate := range ifmapCandidates[1:] {
					sb.WriteString(fmt.Sprintf(", '%s'", candidate.label))
				}
				resp.Diagnostics.AddWarning(
					"cannot assign interface map",
					fmt.Sprintf("node '%s' has %d interface map candidates. Please choose one of ['%s']",
						switchLabel, len(ifmapCandidates), sb.String()))
			}
		}

		plan.Switches[switchLabel] = switchPlan
	}

	// assign previously-selected interface maps
	err = refDesignClient.SetInterfaceMapAssignments(ctx, ifmapAssignments)
	if err != nil {
		if err != nil {
			resp.Diagnostics.AddError("error assigning interface maps", err.Error())
			return
		}
	}

	// having assigned interface maps, link physical assets to graph db 'switch' nodes
	var patch struct {
		SystemId string `json:"system_id"`
	}
	for _, switchPlan := range plan.Switches {
		patch.SystemId = switchPlan.DeviceKey.Value
		err = r.p.client.PatchNode(ctx, blueprintId, goapstra.ObjectId(switchPlan.SystemNodeId.Value), &patch, nil)
		if err != nil {
			resp.Diagnostics.AddWarning("failed to assign switch device", err.Error())
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceBlueprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ResourceBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// some interesting details are in blueprintStatus
	blueprintStatus, err := r.p.client.GetBlueprintStatus(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error fetching blueprint", err.Error())
		return
	}

	// refDesignClient is for the 'datacenter' ref design API
	refDesignClient, err := r.p.client.NewTwoStageL3ClosClient(ctx, blueprintStatus.Id)
	if err != nil {
		resp.Diagnostics.AddError("error getting ref design client", err.Error())
		return
	}

	// get Leaf ASNs
	leafAsns, err := refDesignClient.GetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeAsnPool,
		Name: goapstra.ResourceGroupNameLeafAsn,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}

	// get Leaf IP pools
	leafIps, err := refDesignClient.GetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeIp4Pool,
		Name: goapstra.ResourceGroupNameLeafIps,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}

	// get fabric IP pools
	linkIps, err := refDesignClient.GetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeIp4Pool,
		Name: goapstra.ResourceGroupNameLinkIps,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}

	// get spine ASN pools
	spineAsns, err := refDesignClient.GetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeAsnPool,
		Name: goapstra.ResourceGroupNameSpineAsn,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}

	// get spine IP pools
	spineIps, err := refDesignClient.GetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeIp4Pool,
		Name: goapstra.ResourceGroupNameSpineIps,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}

	// get switch info
	for switchLabel, stateSwitch := range state.Switches {
		// assign details of each known switch (don't add elements to the state.Switches map)
		//	- DeviceKey : required user input
		//	- InterfaceMap : optional user input - if only one option, we'll auto-assign
		//	- DeviceProfile : a.k.a. aos_hcl_model - determined from InterfaceMap, represents physical device/model
		//	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a spine/leaf/etc...
		systemInfo, err := getSystemNodeInfo(ctx, r.p.client, blueprintStatus.Id, switchLabel)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error while reading info for system node '%s'", switchLabel),
				err.Error())
		}
		stateSwitch.SystemNodeId = types.String{Value: systemInfo.id}
		stateSwitch.DeviceKey = types.String{Value: systemInfo.systemId}
		interfaceMap, err := getNodeInterfaceMap(ctx, r.p.client, blueprintStatus.Id, switchLabel)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error while reading interface map for node '%s'", switchLabel),
				err.Error())
		}
		stateSwitch.InterfaceMap = types.String{Value: interfaceMap.label}
		stateSwitch.DeviceProfile = types.String{Value: interfaceMap.deviceProfileId}
		state.Switches[switchLabel] = stateSwitch
	}

	state.Name = types.String{Value: blueprintStatus.Label}
	state.LeafAsns = resourceGroupAllocationToTfStringSlice(leafAsns)
	state.LeafIp4s = resourceGroupAllocationToTfStringSlice(leafIps)
	state.LinkIp4s = resourceGroupAllocationToTfStringSlice(linkIps)
	state.SpineAsns = resourceGroupAllocationToTfStringSlice(spineAsns)
	state.SpineIp4s = resourceGroupAllocationToTfStringSlice(spineIps)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceBlueprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve state
	var state ResourceBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve plan
	var plan ResourceBlueprint
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ensure planned switches (by device key) exist on Apstra
	asi, err := r.p.client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
	if err != nil {
		resp.Diagnostics.AddError("get managed system info", err.Error())
		return
	}
	deviceKeyToSystemInfo := make(map[string]*goapstra.ManagedSystemInfo) // map-ify the Apstra output
	for _, si := range asi {
		deviceKeyToSystemInfo[si.DeviceKey] = &si
	}
	// check each planned switch exists in Apstra, and save the aos_hcl_model (device profile)
	for switchLabel, switchFromPlan := range plan.Switches {
		if si, found := deviceKeyToSystemInfo[switchFromPlan.DeviceKey.Value]; !found {
			resp.Diagnostics.AddError("switch not found",
				fmt.Sprintf("no switch with device_key '%s' exists on Apstra", switchFromPlan.DeviceKey.Value))
			return
		} else {
			switchFromPlan.DeviceProfile = types.String{Value: si.Facts.AosHclModel}
			plan.Switches[switchLabel] = switchFromPlan
		}
	}

	refDesignClient, err := r.p.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error generating reference design client", err.Error())
		return
	}

	// name change?
	if state.Name.Value != plan.Name.Value {
		type metadataNode struct {
			Label string            `json:"label,omitempty"`
			Id    goapstra.ObjectId `json:"id,omitempty"`
		}
		response := &struct {
			Nodes map[string]metadataNode `json:"nodes"`
		}{}
		err = r.p.client.GetNodes(ctx, goapstra.ObjectId(state.Id.Value), goapstra.NodeTypeMetadata, response)
		if err != nil {
			resp.Diagnostics.AddError("error querying blueprint nodes", err.Error())
			return
		}
		if len(response.Nodes) != 1 {
			resp.Diagnostics.AddError(fmt.Sprintf("wrong number of %s nodes", goapstra.NodeTypeMetadata.String()),
				fmt.Sprintf("expecting 1 got %d nodes", len(response.Nodes)))
			return
		}
		var nodeId goapstra.ObjectId
		for _, v := range response.Nodes {
			nodeId = v.Id
		}
		err = r.p.client.PatchNode(ctx, goapstra.ObjectId(state.Id.Value), nodeId, &metadataNode{Label: plan.Name.Value}, nil)
		if err != nil {
			resp.Diagnostics.AddError("error setting blueprint name", err.Error())
			return
		}
		state.Name = types.String{Value: plan.Name.Value}
	}

	if !setsOfStringsMatch(plan.LeafAsns, state.LeafAsns) {
		err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeAsnPool,
			Name:    goapstra.ResourceGroupNameLeafIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.LeafAsns),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
		return
	}

	if !setsOfStringsMatch(plan.LeafIp4s, state.LeafIp4s) {
		err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeIp4Pool,
			Name:    goapstra.ResourceGroupNameLeafIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.LeafIp4s),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
		return
	}

	if !setsOfStringsMatch(plan.LinkIp4s, state.LinkIp4s) {
		err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeIp4Pool,
			Name:    goapstra.ResourceGroupNameLinkIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.LinkIp4s),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
		return
	}

	if !setsOfStringsMatch(plan.SpineAsns, state.SpineAsns) {
		err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeAsnPool,
			Name:    goapstra.ResourceGroupNameSpineIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.SpineAsns),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
		return
	}

	if !setsOfStringsMatch(plan.SpineIp4s, state.SpineIp4s) {
		err = refDesignClient.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeIp4Pool,
			Name:    goapstra.ResourceGroupNameSpineIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.SpineIp4s),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
		return
	}

	// combine switch labels from plan and state into a single set (map of empty struct)
	combinedSwitchLabels := make(map[string]struct{})
	for stateSwitchLabel := range state.Switches {
		combinedSwitchLabels[stateSwitchLabel] = struct{}{}
	}
	for planSwitchLabel := range plan.Switches {
		combinedSwitchLabels[planSwitchLabel] = struct{}{}
	}

	// structure we'll use when assigning interface maps to switches
	ifmapReassignments := make(goapstra.SystemIdToInterfaceMapAssignment)

	// loop over all switches: plan and/or state
	for switchLabel := range combinedSwitchLabels {
		// compare details of each switch
		//	- DeviceKey : required user input - changeable
		//	- InterfaceMap : optional user input - changeable
		//	- DeviceProfile : a.k.a. aos_hcl_model - changeable
		//	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a spine/leaf/etc...

		// fetch the switch graph db node ID and candidate interface maps
		systemNodeId, ifmapCandidates, err := getSystemNodeIdAndIfmapCandidates(ctx, r.p.client, goapstra.ObjectId(state.Id.Value), switchLabel)
		if err != nil {
			resp.Diagnostics.AddWarning("error fetching interface map candidates", err.Error())
			continue
		}
		var foundInPlan, foundInState bool
		var planSwitch, stateSwitch Switch
		planSwitch, foundInPlan = plan.Switches[switchLabel]
		stateSwitch, foundInState = state.Switches[switchLabel]
		switch {
		case foundInPlan && foundInState: // the normal case: switch exists in plan and state
			if planSwitch.SystemNodeId.Value != stateSwitch.SystemNodeId.Value {
				resp.Diagnostics.AddError(
					fmt.Sprintf("node graph entry for %s changed", switchLabel),
					fmt.Sprintf("change: '%s'->'%s' this isn't supposed to happen",
						planSwitch.SystemNodeId.Value, stateSwitch.SystemNodeId.Value))
				return
			}
			if (planSwitch.DeviceKey.Value != stateSwitch.DeviceKey.Value) || // device change?
				(planSwitch.InterfaceMap.Value != stateSwitch.InterfaceMap.Value) {
				// clear existing system id from switch node
				var patch struct {
					SystemId interface{} `json:"system_id"`
				}
				patch.SystemId = nil
				err = r.p.client.PatchNode(ctx, goapstra.ObjectId(plan.Id.Value), goapstra.ObjectId(planSwitch.SystemNodeId.Value), &patch, nil)
				if err != nil {
					resp.Diagnostics.AddWarning("failed to revoke switch device", err.Error())
				}

				// proceed as in Create()
				// validate/choose interface map, build ifmap assignment structure
				if !planSwitch.InterfaceMap.Null && !planSwitch.InterfaceMap.Unknown && !(planSwitch.InterfaceMap.Value == "") {
					// user gave us an interface map label they'd like to use
					ifmapNodeId := ifmapCandidateFromCandidates(planSwitch.InterfaceMap.Value, ifmapCandidates)
					if ifmapNodeId != nil {
						ifmapReassignments[systemNodeId] = ifmapNodeId.id
						planSwitch.DeviceProfile = types.String{Value: ifmapNodeId.deviceProfileId}
					} else {
						resp.Diagnostics.AddWarning(
							"invalid interface map",
							fmt.Sprintf("interface map '%s' not found among candidates for node '%s'",
								planSwitch.InterfaceMap.Value, switchLabel))
					}
				} else {
					// user didn't give us an interface map label; try to find a default
					switch len(ifmapCandidates) {
					case 0: // no candidates!
						resp.Diagnostics.AddWarning(
							"interface map not specified, and no candidates found",
							fmt.Sprintf("no candidate interface maps found for node '%s'", switchLabel))
					case 1: // exact match; we can work with this
						ifmapReassignments[systemNodeId] = ifmapCandidates[0].id
						planSwitch.InterfaceMap = types.String{Value: ifmapCandidates[0].label}
						planSwitch.DeviceProfile = types.String{Value: ifmapCandidates[0].deviceProfileId}
					default: // multiple match!
						sb := strings.Builder{}
						sb.WriteString(fmt.Sprintf("'%s'", ifmapCandidates[0].label))
						for _, candidate := range ifmapCandidates[1:] {
							sb.WriteString(fmt.Sprintf(", '%s'", candidate.label))
						}
						resp.Diagnostics.AddWarning(
							"cannot assign interface map",
							fmt.Sprintf("node '%s' has %d interface map candidates. Please choose one of ['%s']",
								switchLabel, len(ifmapCandidates), sb.String()))
					}
				}
			}
			state.Switches[switchLabel] = planSwitch

		case foundInPlan && !foundInState: // new switch
			// save the SystemNodeId (1:1 relationship with switchLabel in graph db)
			planSwitch.SystemNodeId = types.String{Value: systemNodeId}

			// validate/choose interface map, build ifmap assignment structure
			if !planSwitch.InterfaceMap.Null && !planSwitch.InterfaceMap.Unknown && !(planSwitch.InterfaceMap.Value == "") {
				// user gave us an interface map label they'd like to use
				ifmapNodeId := ifmapCandidateFromCandidates(planSwitch.InterfaceMap.Value, ifmapCandidates)
				if ifmapNodeId != nil {
					ifmapReassignments[systemNodeId] = ifmapNodeId.id
					planSwitch.DeviceProfile = types.String{Value: ifmapNodeId.deviceProfileId}
				} else {
					resp.Diagnostics.AddWarning(
						"invalid interface map",
						fmt.Sprintf("interface map '%s' not found among candidates for node '%s'",
							planSwitch.InterfaceMap.Value, switchLabel))
				}
			} else {
				// user didn't give us an interface map label; try to find a default
				switch len(ifmapCandidates) {
				case 0: // no candidates!
					resp.Diagnostics.AddWarning(
						"interface map not specified, and no candidates found",
						fmt.Sprintf("no candidate interface maps found for node '%s'", switchLabel))
				case 1: // exact match; we can work with this
					ifmapReassignments[systemNodeId] = ifmapCandidates[0].id
					planSwitch.InterfaceMap = types.String{Value: ifmapCandidates[0].label}
					planSwitch.DeviceProfile = types.String{Value: ifmapCandidates[0].deviceProfileId}
				default: // multiple match!
					sb := strings.Builder{}
					sb.WriteString(fmt.Sprintf("'%s'", ifmapCandidates[0].label))
					for _, candidate := range ifmapCandidates[1:] {
						sb.WriteString(fmt.Sprintf(", '%s'", candidate.label))
					}
					resp.Diagnostics.AddWarning(
						"cannot assign interface map",
						fmt.Sprintf("node '%s' has %d interface map candidates. Please choose one of ['%s']",
							switchLabel, len(ifmapCandidates), sb.String()))
				}
			}

			state.Switches[switchLabel] = stateSwitch

		case !foundInPlan && foundInState: // deleted switch
			resp.Diagnostics.AddWarning("request to delete switch not yet supported",
				fmt.Sprintf("cannot delete removed switch '%s'", switchLabel))
		}
	}

	// assign previously-selected interface maps
	err = refDesignClient.SetInterfaceMapAssignments(ctx, ifmapReassignments)
	if err != nil {
		if err != nil {
			resp.Diagnostics.AddError("error assigning interface maps", err.Error())
			return
		}
	}

	// having assigned interface maps, link physical assets to graph db 'switch' nodes
	var patch struct {
		SystemId string `json:"system_id"`
	}
	for _, switchPlan := range plan.Switches {
		patch.SystemId = switchPlan.DeviceKey.Value
		err = r.p.client.PatchNode(ctx, goapstra.ObjectId(state.Id.Value), goapstra.ObjectId(switchPlan.SystemNodeId.Value), &patch, nil)
		if err != nil {
			resp.Diagnostics.AddWarning("failed to assign switch device", err.Error())
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (r resourceBlueprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.client.DeleteBlueprint(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error deleting blueprint", err.Error())
		return
	}
}

func resourceGroupAllocationToTfStringSlice(in *goapstra.ResourceGroupAllocation) []types.String {
	var result []types.String
	for _, pool := range in.PoolIds {
		result = append(result, types.String{Value: string(pool)})
	}
	return result
}

func tfStringSliceToSliceObjectId(in []types.String) []goapstra.ObjectId {
	var result []goapstra.ObjectId
	for _, s := range in {
		result = append(result, goapstra.ObjectId(s.Value))
	}
	return result
}

func asnPoolsExist(ctx context.Context, in []types.String, client *goapstra.Client) (string, error) {
	poolsPerApi, err := client.ListAsnPoolIds(ctx)
	if err != nil {
		return "", fmt.Errorf("error listing available resource pool IDs")
	}

testPool:
	for _, testPool := range in {
		for _, apiPool := range poolsPerApi {
			if goapstra.ObjectId(testPool.Value) == apiPool {
				continue testPool // this one's good, check the next testPool
			}
		}
		return testPool.Value, nil // we looked at every apiPool, none matched testPool
	}
	return "", nil
}

func ip4PoolsExist(ctx context.Context, in []types.String, client *goapstra.Client) (string, error) {
	poolsPerApi, err := client.ListIp4PoolIds(ctx)
	if err != nil {
		return "", fmt.Errorf("error listing available resource pool IDs")
	}

testPool:
	for _, testPool := range in {
		for _, apiPool := range poolsPerApi {
			if goapstra.ObjectId(testPool.Value) == apiPool {
				continue testPool // this one's good, check the next testPool
			}
		}
		return testPool.Value, nil // we looked at every apiPool, none matched testPool
	}
	return "", nil
}

func setsOfStringsMatch(a []types.String, b []types.String) bool {
	if len(a) != len(b) {
		return false
	}

loopA:
	for _, ta := range a {
		for _, tb := range b {
			if ta.Null == tb.Null && ta.Unknown == tb.Unknown && ta.Value == tb.Value {
				continue loopA
			}
		}
		return false
	}
	return true
}

//type switchLabelToInterfaceMap map[string]struct {
//	label string
//	id    string
//}

// getSwitchLabelId queries the graph db for 'switch' type systems, returns
// map[string]string (map[label]id)
func getSwitchLabelId(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (map[string]string, error) {
	var switchQr struct {
		Count int `json:"count"`
		Items []struct {
			System struct {
				Label string `json:"label"`
				Id    string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"name", goapstra.QEStringVal("n_system")},
			{"system_type", goapstra.QEStringVal("switch")},
		}).
		Do(&switchQr)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, switchQr.Count)
	for _, item := range switchQr.Items {
		result[item.System.Label] = item.System.Id
	}

	return result, nil
}

//type switchLabelToCandidateInterfaceMaps map[string][]struct {
//	Id    string
//	Label string
//}
//
//func (o *switchLabelToCandidateInterfaceMaps) string() (string, error) {
//	data, err := json.Marshal(o)
//	if err != nil {
//		return "", err
//	}
//	return string(data), nil
//}

//// getSwitchCandidateInterfaceMaps queries the graph db for
//// 'switch' type systems and their candidate interface maps.
//// It returns switchLabelToCandidateInterfaceMaps.
//func getSwitchCandidateInterfaceMaps(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (switchLabelToCandidateInterfaceMaps, error) {
//	var candidateInterfaceMapsQR struct {
//		Items []struct {
//			System struct {
//				Label string `json:"label"`
//			} `json:"n_system"`
//			InterfaceMap struct {
//				Id    string `json:"id"`
//				Label string `json:"label"`
//			} `json:"n_interface_map"`
//		} `json:"items"`
//	}
//	err := client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("system")},
//			{"name", goapstra.QEStringVal("n_system")},
//			{"system_type", goapstra.QEStringVal("switch")},
//		}).
//		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("logical_device")},
//		}).
//		In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("interface_map")},
//			{"name", goapstra.QEStringVal("n_interface_map")},
//		}).
//		Do(&candidateInterfaceMapsQR)
//	if err != nil {
//		return nil, err
//	}
//
//	result := make(switchLabelToCandidateInterfaceMaps)
//
//	for _, item := range candidateInterfaceMapsQR.Items {
//		mapEntry := result[item.System.Label]
//		mapEntry = append(mapEntry, struct {
//			Id    string
//			Label string
//		}{Id: item.InterfaceMap.Id, Label: item.InterfaceMap.Label})
//		result[item.System.Label] = mapEntry
//	}
//
//	return result, nil
//}

func warnAboutSwitchesMissingFromPlan(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, switches map[string]Switch, diag *diag.Diagnostics) error {
	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, client, bpId)
	if err != nil {
		return err
	}
	var missing []string
	for switchLabel := range switchLabelToGraphDbId {
		if _, found := switches[switchLabel]; !found {
			missing = append(missing, switchLabel)
		}
	}
	if len(missing) != 0 {
		diag.AddWarning("switches with no configuration",
			fmt.Sprintf("please add the following to %s.switches: ['%s']", resourceBlueprintName, strings.Join(missing, "', '")))
	}
	return nil
}

type ifmapInfo struct {
	id              string
	label           string
	deviceProfileId string
}

// getSystemNodeIdAndIfmapCandidates takes the 'label' field representing a
// graph db node with "type='system', returns the node id and a []ifmapInfo
// representing candidate interface maps for that system.
func getSystemNodeIdAndIfmapCandidates(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string) (string, []ifmapInfo, error) {
	var candidateInterfaceMapsQR struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
			InterfaceMap struct {
				Id              string `json:"id"`
				Label           string `json:"label"`
				DeviceProfileId string `json:"device_profile_id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(label)},
			{"name", goapstra.QEStringVal("n_system")},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("logical_device")},
		}).
		In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Do(&candidateInterfaceMapsQR)
	if err != nil {
		return "", nil, err
	}

	var systemNodeId string
	var candidates []ifmapInfo
	for _, item := range candidateInterfaceMapsQR.Items {
		if item.System.Id == "" {
			return "", nil, fmt.Errorf("graph db search for \"type='system', label='%s'\" found match with empty 'id' field", label)
		}
		if systemNodeId != "" && systemNodeId != item.System.Id {
			return "", nil,
				fmt.Errorf("graph db search for \"type='system', label='%s'\" found nodes with different 'id' fields: '%s' and '%s'",
					label, systemNodeId, item.System.Id)
		}
		if systemNodeId == "" {
			systemNodeId = item.System.Id
		}
		candidates = append(candidates, ifmapInfo{
			label:           item.InterfaceMap.Label,
			id:              item.InterfaceMap.Id,
			deviceProfileId: item.InterfaceMap.DeviceProfileId,
		})
	}

	return systemNodeId, candidates, nil
}

// ifmapCandidateFromCandidates finds an interface map (by label) within a
// []ifmapInfo, returns pointer to it, nil if not found.
func ifmapCandidateFromCandidates(label string, candidates []ifmapInfo) *ifmapInfo {
	for _, candidate := range candidates {
		if label == candidate.label {
			return &candidate
		}
	}
	return nil
}

func getNodeInterfaceMap(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string) (*ifmapInfo, error) {
	var interfaceMapQR struct {
		Items []struct {
			InterfaceMap struct {
				Id              string `json:"id"`
				Label           string `json:"label"`
				DeviceProfileId string `json:"device_profile_id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(label)},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Do(&interfaceMapQR)
	if err != nil {
		return nil, err
	}
	if len(interfaceMapQR.Items) != 1 {
		return nil, fmt.Errorf("expected exactly one interface map, got %d", len(interfaceMapQR.Items))
	}
	return &ifmapInfo{
		id:              interfaceMapQR.Items[0].InterfaceMap.Id,
		label:           interfaceMapQR.Items[0].InterfaceMap.Label,
		deviceProfileId: interfaceMapQR.Items[0].InterfaceMap.DeviceProfileId,
	}, nil
}

type systemNodeInfo struct {
	id       string
	label    string
	systemId string
}

func getSystemNodeInfo(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string) (*systemNodeInfo, error) {
	var systemQR struct {
		Items []struct {
			System struct {
				Id       string `json:"id"`
				Label    string `json:"label"`
				SystemID string `json:"system_id"`
			} `json:"n_system"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(label)},
			{"name", goapstra.QEStringVal("n_system")},
		}).Do(&systemQR)
	if err != nil {
		return nil, err
	}
	if len(systemQR.Items) != 1 {
		return nil, fmt.Errorf("expected exactly one system node, got %d", len(systemQR.Items))
	}
	return &systemNodeInfo{
		id:       systemQR.Items[0].System.Id,
		label:    systemQR.Items[0].System.Label,
		systemId: systemQR.Items[0].System.SystemID,
	}, nil
}
