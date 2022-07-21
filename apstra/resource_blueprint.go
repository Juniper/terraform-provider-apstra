package apstra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				//PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			"template_id": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
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
						PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
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
						Type:     types.StringType,
						Computed: true,
					},
				}),
				Optional: true,
			},
		},
	}, nil
}

func (r resourceBlueprintType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceBlueprint{
		p: *(p.(*provider)),
	}, nil
}

type resourceBlueprint struct {
	p provider
}

func (r resourceBlueprint) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
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
	blueprintId, err := r.p.client.CreateBlueprintFromTemplate(ctx, &goapstra.CreateBluePrintFromTemplate{
		RefDesign:  goapstra.RefDesignDatacenter,
		Label:      plan.Name.Value,
		TemplateId: plan.TemplateId.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating Blueprint", err.Error())
		return
	}
	plan.Id = types.String{Value: string(blueprintId)}

	// create a client specific to the reference design
	refDesignClient, err := r.p.client.NewTwoStageL3ClosClient(ctx, blueprintId)

	// warn about switches discovered in the graph db, but which are not in the tf config
	switchesNotInPlan, err := switchLabelsMissingFromPlan(ctx, r.p.client, blueprintId, plan)
	if err != nil {
		resp.Diagnostics.AddError("error while inventorying switches after blueprint creation", err.Error())
		return
	}
	if len(missing) != 0 {
		resp.Diagnostics.AddWarning(
			"switches missing from blueprint configuration",
			fmt.Sprintf("missing config elements for switches: ['%s']", strings.Join(switchesNotInPlan, "', '")))
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

	// determine interface map assignments
	ifMapAssignments, err := generateSwitchToInterfaceMapAssignments(ctx, r.p.client, blueprintId, resp.Diagnostics, &plan)
	if err != nil {
		resp.Diagnostics.AddError("error generating interface map assignments", err.Error())
		return
	}

	// assign interface maps
	err = refDesignClient.SetInterfaceMapAssignments(ctx, ifMapAssignments)
	if err != nil {
		if err != nil {
			resp.Diagnostics.AddError("error assigning interface maps", err.Error())
			return
		}
	}

	// loop over system -> interface_map assignments
	switchLabelToInterfaceMap, err := getSwitchAssignedInterfaceMap(ctx, r.p.client, blueprintId)
	if err != nil {
		resp.Diagnostics.AddError("error reading interface map assignments", err.Error())
		return
	}
	for switchLabel, ifMap := range switchLabelToInterfaceMap {
		// update assignment for switches found in TF plan
		if switchInfo, found := plan.Switches[switchLabel]; found {
			switchInfo.InterfaceMap = types.String{Value: ifMap.label}
			plan.Switches[switchLabel] = switchInfo
		}
	}

	// update graph db node id for all switches in plan
	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, r.p.client, goapstra.ObjectId(plan.Id.Value))
	if err != nil {
		resp.Diagnostics.AddWarning("unable to query blueprint graph db", err.Error())
	}
	for switchLabel, graphDbId := range switchLabelToGraphDbId {
		if switchPlan, found := plan.Switches[switchLabel]; found {
			switchPlan.SystemNodeId = types.String{Value: graphDbId}
			plan.Switches[switchLabel] = switchPlan
		}
	}

	// assign switch (by device_key) to switch node/role for each switch in plan
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

func (r resourceBlueprint) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
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

	// get all system -> interface_map assignments
	switchLabelToInterfaceMap, err := getSwitchAssignedInterfaceMap(ctx, r.p.client, blueprintStatus.Id)
	if err != nil {
		resp.Diagnostics.AddError("error reading interface map assignments", err.Error())
		return
	}
	// loop over system -> interface_map assignments
	for switchLabel, ifMap := range switchLabelToInterfaceMap {
		// update assignment for switches found in TF state
		if switchInfo, found := state.Switches[switchLabel]; found {
			switchInfo.InterfaceMap = types.String{Value: ifMap.label}
			state.Switches[switchLabel] = switchInfo
		}
	}

	// update graph db node id for all switches in blueprint
	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, r.p.client, blueprintStatus.Id)
	if err != nil {
		resp.Diagnostics.AddWarning("unable to query blueprint graph db", err.Error())
	}
	for switchLabel, graphDbId := range switchLabelToGraphDbId {
		switchState := state.Switches[switchLabel]
		switchState.SystemNodeId = types.String{Value: graphDbId}
		state.Switches[switchLabel] = switchState
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
func (r resourceBlueprint) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
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

	refDesignClient, err := r.p.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error generating reference design client", err.Error())
		return
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

	//// prepare three lists (maps of empty struct) of switches
	//switchLabelsOnlyInPlan := make(map[string]struct{})
	//switchLabelsOnlyInState := make(map[string]struct{})
	//switchLabelsInPlanAndState := make(map[string]struct{})
	//for switchLabel := range plan.Switches {
	//	if _, found := state.Switches[switchLabel]; found {
	//		switchLabelsInPlanAndState[switchLabel] = struct{}{}
	//	} else {
	//		switchLabelsOnlyInPlan[switchLabel] = struct{}{}
	//	}
	//}
	//for switchLabel := range state.Switches {
	//	if _, found := plan.Switches[switchLabel]; found {
	//		switchLabelsInPlanAndState[switchLabel] = struct{}{}
	//	} else {
	//		switchLabelsOnlyInState[switchLabel] = struct{}{}
	//	}
	//}
	//
	//// warn about switches which don't exist in the blueprint
	//if len(switchLabelsOnlyInPlan) != 0 {
	//	var sloip []string
	//	for i := range switchLabelsOnlyInPlan {
	//		sloip = append(sloip, i)
	//	}
	//	resp.Diagnostics.AddWarning(
	//		"configured switches not found in blueprint",
	//		fmt.Sprintf("have config details for these switches, but they're not part of the blueprint: ['%s']",
	//			strings.Join(sloip, "', '")))
	//}
	//
	//// switches only in state aren't a problem b/c we commit the plan to the
	//// state file below. Plan evaporates. Leaving that part in for now anyway.
	//
	//// switches in plan and state: compare, make edits as needed.
	//for switchLabel := range switchLabelsInPlanAndState {
	//	if plan.Switches[switchLabel].DeviceKey != state.Switches[switchLabel].DeviceKey {
	//	}
	//	if plan.Switches[switchLabel].DeviceProfile != state.Switches[switchLabel].DeviceProfile {
	//	}
	//	if plan.Switches[switchLabel].InterfaceMap == state.Switches[switchLabel].InterfaceMap {
	//		// todo delete from newIfMapAssignements b/c they match
	//	}
	//	if plan.Switches[switchLabel].SystemNodeId != state.Switches[switchLabel].SystemNodeId {
	//	}
	//}

	// node graph detail we'll need:

	//currentIfMapAssignments, err := getSwitchAssignedInterfaceMap(ctx, r.p.client, goapstra.ObjectId(state.Id.Value))
	//if err != nil {
	//	resp.Diagnostics.AddError("error determining current interface mappings", err.Error())
	//}

	// find the full set of interface map assignments
	ifMapAssignments, err := generateSwitchToInterfaceMapAssignments(ctx, r.p.client, goapstra.ObjectId(state.Id.Value), resp.Diagnostics, &plan)
	if err != nil {
		resp.Diagnostics.AddError("error determining new interface mappings", err.Error())
	}

	// assign interface maps
	err = refDesignClient.SetInterfaceMapAssignments(ctx, ifMapAssignments)
	if err != nil {
		if err != nil {
			resp.Diagnostics.AddError("error assigning interface maps", err.Error())
			return
		}
	}

	// loop over system -> interface_map assignments found via API, update the plan (soon-to-be state)
	switchLabelToInterfaceMap, err := getSwitchAssignedInterfaceMap(ctx, r.p.client, goapstra.ObjectId(plan.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error reading interface map assignments", err.Error())
		return
	}
	for switchLabel, ifMap := range switchLabelToInterfaceMap {
		// update assignment for switches found in TF plan
		if switchInfo, found := plan.Switches[switchLabel]; found {
			switchInfo.InterfaceMap = types.String{Value: ifMap.label}
			plan.Switches[switchLabel] = switchInfo
		}
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

	// update graph db node id for all switches in plan
	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, r.p.client, goapstra.ObjectId(plan.Id.Value))
	if err != nil {
		resp.Diagnostics.AddWarning("unable to query blueprint graph db", err.Error())
	}
	for switchLabel, graphDbId := range switchLabelToGraphDbId {
		if switchPlan, found := plan.Switches[switchLabel]; found {
			switchPlan.SystemNodeId = types.String{Value: graphDbId}
			plan.Switches[switchLabel] = switchPlan
		}
	}

	// assign switch (by device_key) to switch node/role
	// todo - yeah do that

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (r resourceBlueprint) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
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

type switchLabelToInterfaceMap map[string]struct {
	label string
	id    string
}

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

// getSwitchAssignedInterfaceMap queries the graph db for 'switch' type systems
// with an assigned interface map, returns switchLabelToInterfaceMap
func getSwitchAssignedInterfaceMap(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (switchLabelToInterfaceMap, error) {
	var assignedInterfaceMapQR struct {
		Items []struct {
			System struct {
				Label string `json:"label"`
			} `json:"n_system"`
			InterfaceMap struct {
				Id    string `json:"id"`
				Label string `json:"label"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"name", goapstra.QEStringVal("n_system")},
			{"system_type", goapstra.QEStringVal("switch")},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Do(&assignedInterfaceMapQR)
	if err != nil {
		return nil, err
	}

	response := make(switchLabelToInterfaceMap)
	for _, i := range assignedInterfaceMapQR.Items {
		response[i.System.Label] = struct {
			label string
			id    string
		}{label: i.InterfaceMap.Label, id: i.InterfaceMap.Id}
	}

	return response, err
}

type switchLabelToCandidateInterfaceMaps map[string][]struct {
	Id    string
	Label string
}

func (o *switchLabelToCandidateInterfaceMaps) string() (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// getSwitchCandidateInterfaceMaps queries the graph db for
// 'switch' type systems and their candidate interface maps.
// It returns switchLabelToCandidateInterfaceMaps.
func getSwitchCandidateInterfaceMaps(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (switchLabelToCandidateInterfaceMaps, error) {
	var candidateInterfaceMapsQR struct {
		Items []struct {
			System struct {
				Label string `json:"label"`
			} `json:"n_system"`
			InterfaceMap struct {
				Id    string `json:"id"`
				Label string `json:"label"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"name", goapstra.QEStringVal("n_system")},
			{"system_type", goapstra.QEStringVal("switch")},
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
		return nil, err
	}

	result := make(switchLabelToCandidateInterfaceMaps)

	for _, item := range candidateInterfaceMapsQR.Items {
		mapEntry := result[item.System.Label]
		mapEntry = append(mapEntry, struct {
			Id    string
			Label string
		}{Id: item.InterfaceMap.Id, Label: item.InterfaceMap.Label})
		result[item.System.Label] = mapEntry
	}

	return result, nil
}

// planSwitchesToSimpleStruct takes the map[string]Switch from the TF plan,
// returns a simple map containing only valid/populated records
func planSwitchesToSimpleStruct(switchPlan map[string]Switch) map[string]struct {
	ifMap     string
	deviceKey string
} {
	result := make(map[string]struct {
		ifMap     string
		deviceKey string
	})
	for switchLabel, switchInfo := range switchPlan {
		populated := false
		i := struct {
			ifMap     string
			deviceKey string
		}{}
		if !switchInfo.InterfaceMap.IsNull() && !switchInfo.InterfaceMap.IsUnknown() {
			i.ifMap = switchInfo.InterfaceMap.Value
			populated = true
		}
		if !switchInfo.DeviceKey.IsNull() && !switchInfo.DeviceKey.IsUnknown() {
			i.deviceKey = switchInfo.DeviceKey.Value
			populated = true
		}
		if populated {
			result[switchLabel] = i
		}
	}
	return result
}

// generateSwitchToInterfaceMapAssignments takes the 'switches' map from the
// terraform plan and returns goapstra.SystemIdToInterfaceMapAssignment
// representing all switches in the blueprint. Note that the map is keyed by
// node graph system ID, not platform API "managed systems" system ID <sigh>.
func generateSwitchToInterfaceMapAssignments(ctx context.Context, client *goapstra.Client, blueprint goapstra.ObjectId, diag diag.Diagnostics, plan *ResourceBlueprint) (goapstra.SystemIdToInterfaceMapAssignment, error) {
	switchInfoByLabel := make(map[string]struct {
		switchId           string
		ifMapLabelFromPlan string // the label
		ifMapCandidate     map[string]string
	})

	// all map[label]id for all switches
	switchLabelToId, err := getSwitchLabelId(ctx, client, blueprint)
	if err != nil {
		return nil, err
	}

	// all planned info
	switchLabelToPlan := planSwitchesToSimpleStruct(plan.Switches)

	// all ifMap Candidates
	switchLabelToCandidates, err := getSwitchCandidateInterfaceMaps(ctx, client, blueprint)
	if err != nil {
		return nil, err
	}

	// build the easy-to-consume switchInfoByLabel structure
	for switchLabel, switchId := range switchLabelToId {
		ifMapCandidateLabelToId := make(map[string]string)
		for _, candidate := range switchLabelToCandidates[switchLabel] {
			ifMapCandidateLabelToId[candidate.Label] = candidate.Id
		}
		switchInfoByLabel[switchLabel] = struct {
			switchId           string
			ifMapLabelFromPlan string
			ifMapCandidate     map[string]string
		}{switchId: switchId, ifMapLabelFromPlan: switchLabelToPlan[switchLabel].ifMap, ifMapCandidate: ifMapCandidateLabelToId}
	}

	result := make(goapstra.SystemIdToInterfaceMapAssignment)
	var unassigned []string
	for switchLabel, switchId := range switchLabelToId {
		ifMapFromPlan := switchInfoByLabel[switchLabel].ifMapLabelFromPlan
		if ifMapFromPlan != "" {
			// user supplied an ifMapLabel
			if ifMapId, found := switchInfoByLabel[switchLabel].ifMapCandidate[ifMapFromPlan]; !found {
				unassigned = append(unassigned, switchLabel)
				diag.AddWarning(
					"invalid interface map not assigned to system",
					fmt.Sprintf("interface map '%s' not found among candidates for '%s'", ifMapFromPlan, switchLabel))
			} else {
				// this right here is our objective (1 of 2)
				result[switchId] = ifMapId
			}
		} else {
			// user didn't supply an ifMapLabel
			switch len(switchInfoByLabel[switchLabel].ifMapCandidate) {
			case 0:
				unassigned = append(unassigned, switchLabel)
				diag.AddWarning(
					"cannot assign interface map",
					fmt.Sprintf("system '%s' has no candidate interface maps", switchLabel))
			case 1:
				// this loop only runs once
				for _, ifMapId := range switchInfoByLabel[switchLabel].ifMapCandidate {
					// this right here is our objective (2 of 2)
					result[switchId] = ifMapId
				}
			default:
				unassigned = append(unassigned, switchLabel)
				diag.AddWarning(
					"cannot assign interface map",
					fmt.Sprintf("cowardly refusing to choose between %d interface maps for '%s'",
						len(switchInfoByLabel[switchLabel].ifMapCandidate), switchLabel))
			}
		}
	}
	if len(unassigned) != 0 {
		diag.AddWarning(
			"switches with no interface maps assigned",
			fmt.Sprintf("no interface maps assigned: '%s'", strings.Join(unassigned, "', '")))
	}
	return result, nil
}

func switchLabelsMissingFromPlan(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, plan ResourceBlueprint) ([]string, error) {
	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, client, bpId)
	if err != nil {
		return nil, err
	}
	var result []string
	for switchLabel, _ := range switchLabelToGraphDbId {
		if _, found := plan.Switches[switchLabel]; !found {
			result = append(result, switchLabel)
		}
	}
	return result, nil
}
