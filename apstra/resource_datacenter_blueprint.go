package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterBlueprint{}
var _ resource.ResourceWithValidateConfig = &resourceDatacenterBlueprint{}
var _ versionValidator = &resourceDatacenterBlueprint{}

type resourceDatacenterBlueprint struct {
	client           *goapstra.Client
	minClientVersion *version.Version
	maxClientVersion *version.Version
}

func (o *resourceDatacenterBlueprint) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *resourceDatacenterBlueprint) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceDatacenterBlueprint) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource instantiates a Datacenter Blueprint from a template.",
		Attributes:          blueprint.Blueprint{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterBlueprint) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Cannot proceed without a client
	if o.client == nil {
		return
	}

	var config blueprint.Blueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the min/max API versions required by the client. These elements set within 'o'
	// do not persist after ValidateConfig exits even though 'o' is a pointer receiver.
	o.minClientVersion, o.maxClientVersion = config.MinMaxApiVersions(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if o.client == nil {
		// Bail here because we can't validate config's API version needs if the client doesn't exist.
		// This method should be called again (after the provider's Configure() method) with a non-nil
		// client pointer.
		return
	}

	// validate version compatibility between the API server and the configuration's min/max needs.
	o.checkVersion(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//// ensure ASN pools from the plan exist on Apstra
	//var asnPools []attr.Value
	//asnPools = append(asnPools, config.SpineAsnPoolIds.Elements()...)
	//asnPools = append(asnPools, config.LeafAsnPoolIds.Elements()...)
	//asnPools = append(asnPools, config.AccessAsnPoolIds.Elements()...)
	//missing := findMissingAsnPools(ctx, asnPools, o.client, &resp.Diagnostics)
	//if len(missing) > 0 {
	//	resp.Diagnostics.AddError("cannot assign ASN pool",
	//		fmt.Sprintf("requested pool id(s) %s not found", missing))
	//}
	//
	//// ensure Ip4 pools from the plan exist on Apstra
	//var ipv4Pools []attr.Value
	//ipv4Pools = append(ipv4Pools, config.SpineIp4PoolIds.Elements()...)  // Spine loopback
	//ipv4Pools = append(ipv4Pools, config.LeafIp4PoolIds.Elements()...)   // leaf loopback
	//ipv4Pools = append(ipv4Pools, config.AccessIp4PoolIds.Elements()...) // access loopback
	//ipv4Pools = append(ipv4Pools, config.SpineLeafPoolIp4.Elements()...) // Spine fabric
	//ipv4Pools = append(ipv4Pools, config.LeafLeafPoolIp4.Elements()...)  // leaf-only fabric
	//ipv4Pools = append(ipv4Pools, config.LeafMlagPeerIp4.Elements()...)  // leaf peer link
	//ipv4Pools = append(ipv4Pools, config.AccessEsiPeerIp4.Elements()...) // access peer link
	//ipv4Pools = append(ipv4Pools, config.VtepIps.Elements()...)          // vtep
	//missing = findMissingIpv4Pools(ctx, ipv4Pools, o.client, &resp.Diagnostics)
	//if len(missing) > 0 {
	//	resp.Diagnostics.AddError("cannot assign IPv4 pool",
	//		fmt.Sprintf("requested pool id(s) %s not found", missing))
	//}
	//
	//// ensure Ip6 pools from the plan exist on Apstra
	//var ip6Pools []attr.Value
	//ip6Pools = append(ip6Pools, config.SpineLeafPoolIp6.Elements()...) // Spine fabric
	//missing = findMissingIpv6Pools(ctx, ip6Pools, o.client, &resp.Diagnostics)
	//if len(missing) > 0 {
	//	resp.Diagnostics.AddError("cannot assign IPv6 pool",
	//		fmt.Sprintf("requested pool(s) %s not found", missing))
	//}
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// populate device profile IDs, detect errors along the way
	//config.populateDeviceProfileIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// populate interface map IDs, detect errors along the way
	//config.populateInterfaceMapIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
}

func (o *resourceDatacenterBlueprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//// compute the device profile of each switch the user told us about (use device key)
	//plan.populateDeviceProfileIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateBlueprintFromTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Rack Based Blueprint", err.Error())
	}

	apiData, err := o.client.GetBlueprintStatus(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint after creation", err.Error())
	}

	// Create new state object
	var state blueprint.Blueprint
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.FabricAddressing = plan.FabricAddressing // blindly copy because resource.RequiresReplace()
	state.TemplateId = plan.TemplateId             // blindly copy because resource.RequiresReplace()

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	//plan.allocateResources(ctx, o.client, blueprintId, &resp.Diagnostics)
	//if diags.HasError() {
	//	// todo set state?
	//	return
	//}
	//
	//// warn the user about any omitted resource group allocations
	//warnMissingResourceGroupAllocations(ctx, Blueprint, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// compute the Blueprint "system" node IDs (switches)
	//plan.populateSystemNodeIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// compute the interface map IDs
	//plan.populateInterfaceMapIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// set interface map assignments (selects hardware model, but not specific instance)
	//plan.assignInterfaceMaps(ctx, Blueprint, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// warn about switches discovered in the graph db, and which do not appear in the tf config
	//plan.warnSwitchConfigVsBlueprint(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// assign switches (managed devices) to Blueprint system nodes
	//plan.assignManagedDevices(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	////// structure we'll use when assigning interface maps to switches
	////ifmapAssignments := make(goapstra.SystemIdToInterfaceMapAssignment)
	////
	//// assign details of each configured switch (don't add elements to the plan.Switches map)
	////	- DeviceKey : required user input
	////	- InterfaceMap : optional user input - if only one option, we'll auto-assign
	////	- DeviceProfile : a.k.a. aos_hcl_model - determined from InterfaceMap, represents physical device/model
	////	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a Spine/leaf/etc...
	////for switchLabel, switchPlan := range plan.Switches {
	////	// fetch the switch graph db node ID and candidate interface maps
	////	systemNodeId, ifmapCandidates, err := getSystemNodeIdAndIfmapCandidates(ctx, r.p.client, blueprintId, switchLabel)
	////	if err != nil {
	////		resp.Diagnostics.AddWarning("error fetching interface map candidates", err.Error())
	////		continue
	////	}
	////
	////	// save the SystemNodeId (1:1 relationship with switchLabel in graph db)
	////	switchPlan.SystemNodeId = types.String{Value: systemNodeId}
	////
	////	// validate/choose interface map, build ifmap assignment structure
	////	if !switchPlan.InterfaceMap.Null && !switchPlan.InterfaceMap.Unknown && !(switchPlan.InterfaceMap.Value == "") {
	////		// user gave us an interface map label they'd like to use
	////		ifmapNodeId := ifmapCandidateFromCandidates(switchPlan.InterfaceMap.Value, ifmapCandidates)
	////		if ifmapNodeId != nil {
	////			ifmapAssignments[systemNodeId] = ifmapNodeId.id
	////			switchPlan.DeviceProfile = types.String{Value: ifmapNodeId.deviceProfileId}
	////		} else {
	////			resp.Diagnostics.AddWarning(
	////				"invalid interface map",
	////				fmt.Sprintf("interface map '%s' not found among candidates for node '%s'",
	////					switchPlan.InterfaceMap.Value, switchLabel))
	////		}
	////	} else {
	////		// user didn't give us an interface map label; try to find a default
	////		switch len(ifmapCandidates) {
	////		case 0: // no candidates!
	////			resp.Diagnostics.AddWarning(
	////				"interface map not specified, and no candidates found",
	////				fmt.Sprintf("no candidate interface maps found for node '%s'", switchLabel))
	////		case 1: // exact match; we can work with this
	////			ifmapAssignments[systemNodeId] = ifmapCandidates[0].id
	////			switchPlan.InterfaceMap = types.String{Value: ifmapCandidates[0].label}
	////			switchPlan.DeviceProfile = types.String{Value: ifmapCandidates[0].deviceProfileId}
	////		default: // multiple match!
	////			sb := strings.Builder{}
	////			sb.WriteString(fmt.Sprintf("'%s'", ifmapCandidates[0].label))
	////			for _, candidate := range ifmapCandidates[1:] {
	////				sb.WriteString(fmt.Sprintf(", '%s'", candidate.label))
	////			}
	////			resp.Diagnostics.AddWarning(
	////				"cannot assign interface map",
	////				fmt.Sprintf("node '%s' has %d interface map candidates. Please choose one of ['%s']",
	////					switchLabel, len(ifmapCandidates), sb.String()))
	////		}
	////	}
	////
	////	plan.Switches[switchLabel] = switchPlan
	////}
	//
	////// assign previously-selected interface maps
	////err = refDesignClient.SetInterfaceMapAssignments(ctx, ifmapAssignments)
	////if err != nil {
	////	if err != nil {
	////		resp.Diagnostics.AddError("error assigning interface maps", err.Error())
	////		return
	////	}
	////}
	//
	////// having assigned interface maps, link physical assets to graph db 'switch' nodes
	////var patch struct {
	////	SystemId string `json:"system_id"`
	////}
	////for _, switchPlan := range plan.Switches {
	////	patch.SystemId = switchPlan.DeviceKey.Value
	////	err = r.p.client.PatchNode(ctx, blueprintId, goapstra.ObjectId(switchPlan.SystemNodeId.Value), &patch, nil)
	////	if err != nil {
	////		resp.Diagnostics.AddWarning("failed to assign switch device", err.Error())
	////	}
	////}
	//
	//diags = resp.State.Set(ctx, &plan)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
}

func (o *resourceDatacenterBlueprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state blueprint.Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// some interesting details are in blueprintStatus
	apiData, err := o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error fetching Blueprint",
			fmt.Sprintf("Could not read %q - %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	//// create a client specific to the reference design
	//Blueprint, err := o.client.NewTwoStageL3ClosClient(ctx, blueprintStatus.Id)
	//if err != nil {
	//	resp.Diagnostics.AddError("error getting Blueprint client", err.Error())
	//	return
	//}

	// create new state object with some obvious values
	var newState blueprint.Blueprint
	newState.LoadApiData(ctx, apiData, &resp.Diagnostics)
	newState.FabricAddressing = state.FabricAddressing // blindly copy because resource.RequiresReplace()
	newState.TemplateId = state.TemplateId             // blindly copy because resource.RequiresReplace()

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)

	//// collect resource pool values into new state object
	//for _, tag := range listOfResourceGroupAllocationTags() {
	//	newState.readPoolAllocationFromApstraIntoElementByTfsdkTag(ctx, tag, Blueprint, &resp.Diagnostics)
	//}
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// read switch info from Apstra, then delete any switches unknown to the state file
	//newState.readSwitchesFromGraphDb(ctx, o.client, &resp.Diagnostics)
	//for sl := range newState.Switches {
	//	if _, ok := state.Switches[sl]; !ok {
	//		delete(newState.Switches, sl)
	//	}
	//}
	//
}

// Update resource
func (o *resourceDatacenterBlueprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve plan
	var plan blueprint.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// name change is the only possible update method (other attributes trigger replace)
	plan.SetName(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//// reset resource group allocations as needed
	//for _, tag := range listOfResourceGroupAllocationTags() {
	//	plan.updateResourcePoolAllocationByTfsdkTag(ctx, tag, Blueprint, &state, &resp.Diagnostics)
	//	if resp.Diagnostics.HasError() {
	//		return
	//	}
	//}
	//
	//// compute the device profile of each switch the user told us about (use device key)
	//plan.populateDeviceProfileIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// compute the Blueprint "system" node IDs for the planned switches
	//plan.populateSystemNodeIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// compute the interface map IDs for the planned switches
	//plan.populateInterfaceMapIds(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// prepare two lists. changed switches appear on
	//// both lists, will be wiped out then added as new
	//switchesToDeleteOrChange := make(map[string]struct{}) // these switches will be completely wiped out
	//switchesToAddOrChange := make(map[string]struct{})    // these switches will be added as new
	//
	//// accumulate list elements from plan
	//for switchLabel := range plan.Switches {
	//	if _, found := state.Switches[switchLabel]; !found {
	//		switchesToAddOrChange[switchLabel] = struct{}{}
	//	}
	//}
	//
	//// accumulate list elements from state
	//for switchLabel := range state.Switches {
	//	if _, found := plan.Switches[switchLabel]; !found {
	//		switchesToDeleteOrChange[switchLabel] = struct{}{}
	//		continue
	//	}
	//	if !state.Switches[switchLabel].Equal(plan.Switches[switchLabel]) {
	//		switchesToAddOrChange[switchLabel] = struct{}{}
	//		switchesToDeleteOrChange[switchLabel] = struct{}{}
	//	}
	//}
	//
	//// wipe out device allocation for delete/change switches
	//for label := range switchesToDeleteOrChange {
	//	nodeId := state.Switches[label].Attrs["system_node_id"].(types.String)
	//	state.releaseManagedDevice(ctx, nodeId.Value, o.client, &resp.Diagnostics)
	//	if resp.Diagnostics.HasError() {
	//		return
	//	}
	//}
	//
	//// wipe out node->interface map assignments for delete/change switches
	//assignments := make(goapstra.SystemIdToInterfaceMapAssignment, len(switchesToDeleteOrChange))
	//for label := range switchesToDeleteOrChange {
	//	nodeId := state.Switches[label].Attrs["system_node_id"].(types.String)
	//	assignments[nodeId.Value] = nil
	//}
	//err = Blueprint.SetInterfaceMapAssignments(ctx, assignments)
	//if err != nil {
	//	resp.Diagnostics.AddError("error clearing interface map assignment", err.Error())
	//}
	//
	//// create interface_map assignments for add/change switches
	//for label := range switchesToAddOrChange {
	//	nodeId := plan.Switches[label].Attrs["system_node_id"].(types.String)
	//	interfaceMapId := plan.Switches[label].Attrs["interface_map_id"].(types.String)
	//	assignments[nodeId.Value] = interfaceMapId.Value
	//}
	//err = Blueprint.SetInterfaceMapAssignments(ctx, assignments)
	//if err != nil {
	//	resp.Diagnostics.AddError("error setting interface map assignment", err.Error())
	//}
	//
	//// assert device allocation for entire plan
	//plan.assignManagedDevices(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}

	apiData, err := o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint after creation", err.Error())
	}

	// Create new state object
	var state blueprint.Blueprint
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.FabricAddressing = plan.FabricAddressing // blindly copy because resource.RequiresReplace()
	state.TemplateId = plan.TemplateId             // blindly copy because resource.RequiresReplace()

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete resource
func (o *resourceDatacenterBlueprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state blueprint.Blueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteBlueprint(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Blueprint", err.Error())
		return
	}
}

// getSwitchLabelId queries the graph db for 'switch' type systems, returns
// map[string]string (map[label]id)
func getSwitchLabelId(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (map[string]string, error) {
	var switchQr struct {
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

	result := make(map[string]string, len(switchQr.Items))
	for _, item := range switchQr.Items {
		result[item.System.Label] = item.System.Id
	}

	return result, nil
}

//// populateInterfaceMapIds attempts to populate (and validate when known) the
//// interface_map_id for each switch using either the design API's global catalog
//// (Blueprint not created yet) or graphDB elements (Blueprint exists).
//func (o *Blueprint) populateInterfaceMapIds(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	if o.Id.IsNull() {
//		o.populateInterfaceMapIdsFromGC(ctx, client, diags)
//	} else {
//		o.populateInterfaceMapIdsFromBP(ctx, client, diags)
//	}
//}

//func (o *Blueprint) populateInterfaceMapIdsFromGC(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	for switchLabel, plannedSwitch := range o.Switches {
//		devProfile := plannedSwitch.Attrs["device_profile_id"].(types.String)
//		ifMapId := plannedSwitch.Attrs["interface_map_id"].(types.String)
//
//		// sanity check
//		if devProfile.IsUnknown() {
//			diags.AddError(errProviderBug,
//				fmt.Sprintf("attempt to populateInterfaceMapIdsFromGC for switch '%s' while device profile is unknown",
//					switchLabel))
//			return
//		}
//
//		if !ifMapId.IsNull() {
//			assertInterfaceMapSupportsDeviceProfile(ctx, client, ifMapId.Value, devProfile.Value, diags)
//			if diags.HasError() {
//				return
//			}
//			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Value: ifMapId.Value}
//			continue
//		}
//
//		// todo: try to populate interface map ID
//		//  until we have a way to parse the template to learn logical device types, this is impossible
//		o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Null: true}
//	}
//}

//func (o *Blueprint) populateInterfaceMapIdsFromBP(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	// structure for receiving results of Blueprint query
//	var candidateInterfaceMapsQR struct {
//		Items []struct {
//			InterfaceMap struct {
//				Id string `json:"id"`
//			} `json:"n_interface_map"`
//		} `json:"items"`
//	}
//
//	for switchLabel, plannedSwitch := range o.Switches {
//		devProfile := plannedSwitch.Attrs["device_profile_id"].(types.String)
//		ifMapId := plannedSwitch.Attrs["interface_map_id"].(types.String)
//
//		// sanity check
//		if devProfile.IsUnknown() {
//			diags.AddError(errProviderBug,
//				fmt.Sprintf("attempt to populateInterfaceMapIdsFromBP for switch '%s' while device profile is unknown",
//					switchLabel))
//			return
//		}
//
//		query := client.NewQuery(goapstra.ObjectId(o.Id.Value)).
//			SetContext(ctx).
//			Node([]goapstra.QEEAttribute{
//				{"type", goapstra.QEStringVal("system")},
//				{"label", goapstra.QEStringVal(switchLabel)},
//				//{"name", goapstra.QEStringVal("n_system")},
//			}).
//			Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
//			Node([]goapstra.QEEAttribute{
//				{"type", goapstra.QEStringVal("logical_device")},
//			}).
//			In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}})
//
//		if ifMapId.IsUnknown() {
//			query = query.Node([]goapstra.QEEAttribute{
//				{"type", goapstra.QEStringVal("interface_map")},
//				{"name", goapstra.QEStringVal("n_interface_map")},
//			})
//		} else {
//			query = query.Node([]goapstra.QEEAttribute{
//				{"type", goapstra.QEStringVal("interface_map")},
//				{"name", goapstra.QEStringVal("n_interface_map")},
//				{"id", goapstra.QEStringVal(ifMapId.Value)},
//			})
//		}
//
//		err := query.Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("device_profile")}}).
//			Node([]goapstra.QEEAttribute{
//				{"type", goapstra.QEStringVal("device_profile")},
//				{"device_profile_id", goapstra.QEStringVal(devProfile.Value)},
//			}).
//			Do(&candidateInterfaceMapsQR)
//		if err != nil {
//			diags.AddError("error running interface map query", err.Error())
//			return
//		}
//
//		switch len(candidateInterfaceMapsQR.Items) {
//		case 0:
//			diags.AddAttributeError(path.Root("switches").AtMapKey(switchLabel),
//				"unable to assign interface_map",
//				fmt.Sprintf("no interface_map links system '%s' to device profile '%s'",
//					switchLabel, devProfile.Value))
//		case 1:
//			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Value: candidateInterfaceMapsQR.Items[0].InterfaceMap.Id}
//		default:
//		}
//	}
//}

//func (o *Blueprint) populateSystemNodeIds(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	var candidateSystemsQR struct {
//		Items []struct {
//			System struct {
//				Id string `json:"id"`
//			} `json:"n_system"`
//		} `json:"items"`
//	}
//
//	for switchLabel := range o.Switches {
//		err := client.NewQuery(goapstra.ObjectId(o.Id.Value)).
//			SetContext(ctx).
//			Node([]goapstra.QEEAttribute{
//				{"type", goapstra.QEStringVal("system")},
//				{"label", goapstra.QEStringVal(switchLabel)},
//				{"name", goapstra.QEStringVal("n_system")},
//			}).
//			Do(&candidateSystemsQR)
//		if err != nil {
//			diags.AddError("error querying for bp system node", err.Error())
//		}
//
//		switch len(candidateSystemsQR.Items) {
//		case 0:
//			diags.AddError("switch node not found in Blueprint",
//				fmt.Sprintf("switch/system node with label '%s' not found in Blueprint", switchLabel))
//			return
//		case 1:
//			// no error case
//		default:
//			diags.AddError("multiple switches found in Blueprint",
//				fmt.Sprintf("switch/system node with label '%s': %d matches found in Blueprint",
//					switchLabel, len(candidateSystemsQR.Items)))
//			return
//		}
//
//		o.Switches[switchLabel].Attrs["system_node_id"] = types.String{Value: candidateSystemsQR.Items[0].System.Id}
//	}
//}
//
//func (o *Blueprint) assignInterfaceMaps(ctx context.Context, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	assignments := make(goapstra.SystemIdToInterfaceMapAssignment, len(o.Switches))
//	for k, v := range o.Switches {
//		switch {
//		case v.Attrs["system_node_id"].IsUnknown():
//			diags.AddError(errProviderBug, fmt.Sprintf("switch '%s' system_id is unknown", k))
//		case v.Attrs["system_node_id"].IsNull():
//			diags.AddError(errProviderBug, fmt.Sprintf("switch '%s' system_id is null", k))
//		case v.Attrs["interface_map_id"].IsUnknown():
//			diags.AddError(errProviderBug, fmt.Sprintf("switch '%s' interface_map_id is unknown", k))
//		case v.Attrs["interface_map_id"].IsNull():
//			assignments[v.Attrs["system_node_id"].(types.String).Value] = nil
//		default:
//			assignments[v.Attrs["system_node_id"].(types.String).Value] = v.Attrs["interface_map_id"].(types.String).Value
//		}
//
//		err := client.SetInterfaceMapAssignments(ctx, assignments)
//		if err != nil {
//			diags.AddError("error assigning interface maps", err.Error())
//		}
//	}
//
//	err := client.SetInterfaceMapAssignments(ctx, assignments)
//	if err != nil {
//		diags.AddError("error assigning interface maps", err.Error())
//	}
//}
//
//func (o *Blueprint) assignManagedDevices(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	//// having assigned interface maps, link physical assets to graph db 'switch' nodes
//	var patch struct {
//		SystemId string `json:"system_id"`
//	}
//	bpId := goapstra.ObjectId(o.Id.Value)
//	for _, plannedSwitch := range o.Switches {
//		patch.SystemId = plannedSwitch.Attrs["device_key"].(types.String).Value
//		nodeId := goapstra.ObjectId(plannedSwitch.Attrs["system_node_id"].(types.String).Value)
//		err := client.PatchNode(ctx, bpId, nodeId, &patch, nil)
//		if err != nil {
//			diags.AddWarning(fmt.Sprintf("failed to assign switch device for node '%s'", nodeId), err.Error())
//		}
//	}
//}
//
//func (o *Blueprint) releaseManagedDevice(ctx context.Context, nodeId string, client *goapstra.client, diags *diag.Diagnostics) {
//	var patch struct {
//		_ interface{} `json:"system_id"`
//	}
//	err := client.PatchNode(ctx, goapstra.ObjectId(o.Id.Value), goapstra.ObjectId(nodeId), &patch, nil)
//	if err != nil {
//		diags.AddWarning(fmt.Sprintf("failed to assign switch device for node '%s'", nodeId), err.Error())
//	}
//}
//
//func assertInterfaceMapSupportsDeviceProfile(ctx context.Context, client *goapstra.client, ifMapId string, devProfileId string, diags *diag.Diagnostics) {
//	ifMap, err := client.GetInterfaceMapDigest(ctx, goapstra.ObjectId(ifMapId))
//	if err != nil {
//		var ace goapstra.ApstraClientErr
//		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
//			diags.AddError("interface map not found",
//				fmt.Sprintf("interfacem map with id '%s' not found", ifMapId))
//		}
//		diags.AddError(fmt.Sprintf("error fetching interface map '%s'", ifMapId), err.Error())
//	}
//	if string(ifMap.DeviceProfile.Id) != devProfileId {
//		diags.AddError(
//			errInvalidConfig,
//			fmt.Sprintf("interface map '%s' works with device profile '%s', not '%s'",
//				ifMapId, ifMap.DeviceProfile.Id, devProfileId))
//	}
//}
//
////func assertInterfaceMapExists(ctx context.Context, client *goapstra.client, id string, diags *diag.Diagnostics) {
////	_, err := client.GetInterfaceMapDigest(ctx, goapstra.ObjectId(id))
////	if err != nil {
////		var ace goapstra.ApstraClientErr
////		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
////			diags.AddError("interface map not found",
////				fmt.Sprintf("interfacem map with id '%s' not found", id))
////		}
////		diags.AddError(fmt.Sprintf("error fetching interface map '%s'", id), err.Error())
////	}
////}
//
//// populateSwitchNodeAndInterfaceMapIds
//func (o *Blueprint) populateSwitchNodeAndInterfaceMapIds(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	if o.Id.IsUnknown() {
//		diags.AddError(errProviderBug, "attempt to populateSwitchNodeAndInterfaceMapIds while Blueprint ID is unknown")
//		return
//	}
//
//	for switchLabel, plannedSwitch := range o.Switches {
//		nodeId, ldId, err := getSystemNodeIdAndLogicalDeviceId(ctx, client, goapstra.ObjectId(o.Id.Value), switchLabel)
//		if err != nil {
//			diags.AddAttributeError(
//				path.Root("switches").AtMapKey(switchLabel),
//				fmt.Sprintf("error fetching node ID for switch '%s'", switchLabel),
//				err.Error())
//			return
//		}
//		// save the system node ID
//		o.Switches[switchLabel].Attrs["system_node_id"] = types.String{Value: nodeId}
//
//		// device profile should be known at this time
//		if plannedSwitch.Attrs["device_profile_id"].IsUnknown() {
//			diags.AddAttributeWarning(
//				path.Root("switches").AtMapKey(switchLabel),
//				"device profile unknown",
//				fmt.Sprintf("device profile for '%s' unknown - this is probably a bug", switchLabel))
//			continue
//		}
//
//		// fetch interface map candidates
//		deviceProfile := plannedSwitch.Attrs["device_profile_id"].(types.String).Value
//		imaps, err := client.GetInterfaceMapDigestsLogicalDeviceAndDeviceProfile(ctx, goapstra.ObjectId(ldId), goapstra.ObjectId(deviceProfile))
//		if err != nil {
//			diags.AddAttributeError(
//				path.Root("switches").AtMapKey(switchLabel),
//				fmt.Sprintf("error fetching interface map digests for switch '%s'", switchLabel),
//				err.Error())
//		}
//		switch len(imaps) {
//		case 0:
//			diags.AddAttributeError(
//				path.Root("switches").AtMapKey(switchLabel),
//				fmt.Sprintf("switch '%s': could not find interface map", switchLabel),
//				fmt.Sprintf("no interface maps link logical device '%s' to device profile '%s'",
//					ldId, switchLabel))
//		case 1:
//			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Value: string(imaps[0].Id)}
//		default:
//			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Null: true}
//			imapIds := make([]string, len(imaps))
//			for i, imap := range imaps {
//				imapIds[i] = string(imap.Id)
//			}
//			diags.AddAttributeWarning(
//				path.Root("switches").AtMapKey(switchLabel),
//				fmt.Sprintf("switch '%s': multiple interface map candidates", switchLabel),
//				fmt.Sprintf("please configure 'interface_map_id' using one of the valid candidate IDs: '%s'",
//					strings.Join(imapIds, "', '")))
//		}
//	}
//}
//
//func (o *Blueprint) warnSwitchConfigVsBlueprint(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, client, goapstra.ObjectId(o.Id.Value))
//	if err != nil {
//		diags.AddError("error getting Blueprint switch inventory", err.Error())
//		return
//	}
//
//	var missing []string
//	for switchLabel := range switchLabelToGraphDbId {
//		if _, found := o.Switches[switchLabel]; !found {
//			missing = append(missing, switchLabel)
//		}
//	}
//
//	var extra []string
//	for switchLabel := range o.Switches {
//		if _, found := switchLabelToGraphDbId[switchLabel]; !found {
//			extra = append(extra, switchLabel)
//		}
//	}
//
//	// warn about missing switches
//	if len(missing) != 0 {
//		diags.AddAttributeWarning(
//			path.Root("switches"),
//			"switch missing from plan",
//			fmt.Sprintf("Blueprint expects the following switches: ['%s']",
//				strings.Join(missing, "', '")))
//	}
//
//	// warn about extraneous switches mentioned in config
//	if len(extra) != 0 {
//		diags.AddAttributeWarning(
//			path.Root("switches"),
//			"extraneous switches found in configuration",
//			fmt.Sprintf("please remove switches not needed by Blueprint: '%s'",
//				strings.Join(extra, "', '")))
//	}
//}

//type ifmapInfo struct {
//	id              string
//	label           string
//	deviceProfileId string
//}
//
//// getSystemNodeIdAndLogicalDeviceId takes the 'label' field representing a
//// graph db node with "type='system', returns its node id and the linked
//// logical device Id
//func getSystemNodeIdAndLogicalDeviceId(ctx context.Context, client *goapstra.client, bpId goapstra.ObjectId, label string) (string, string, error) {
//	var systemAndLogicalDeviceQR struct {
//		Items []struct {
//			System struct {
//				Id string `json:"id"`
//			} `json:"n_system"`
//			LogicalDevice struct {
//				Id string `json:"id"`
//			} `json:"n_logical_device"`
//		} `json:"items"`
//	}
//
//	err := client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("system")},
//			{"label", goapstra.QEStringVal(label)},
//			{"name", goapstra.QEStringVal("n_system")},
//		}).
//		Out([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("logical_device")},
//		}).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("logical_device")},
//			{"name", goapstra.QEStringVal("n_logical_device")},
//		}).
//		Do(&systemAndLogicalDeviceQR)
//	if err != nil {
//		return "", "", err
//	}
//
//	switch len(systemAndLogicalDeviceQR.Items) {
//	case 0:
//		return "", "", fmt.Errorf("query result for 'system' node, with label '%s' empty", label)
//	case 1:
//		// expected behavior - no error
//	default:
//		return "", "", fmt.Errorf("query result for 'system' node, with label '%s' returned %d results",
//			label, len(systemAndLogicalDeviceQR.Items))
//	}
//
//	return systemAndLogicalDeviceQR.Items[0].System.Id, systemAndLogicalDeviceQR.Items[0].LogicalDevice.Id, nil
//}
//
//// getSystemNodeIdAndIfmapCandidates takes the 'label' field representing a
//// graph db node with "type='system', returns the node id and a []ifmapInfo
//// representing candidate interface maps for that system.
//func getSystemNodeIdAndIfmapCandidates(ctx context.Context, client *goapstra.client, bpId goapstra.ObjectId, label string) (string, []ifmapInfo, error) {
//	var candidateInterfaceMapsQR struct {
//		Items []struct {
//			System struct {
//				Id string `json:"id"`
//			} `json:"n_system"`
//			InterfaceMap struct {
//				Id              string `json:"id"`
//				Label           string `json:"label"`
//				DeviceProfileId string `json:"device_profile_id"`
//			} `json:"n_interface_map"`
//		} `json:"items"`
//	}
//	err := client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("system")},
//			{"label", goapstra.QEStringVal(label)},
//			{"name", goapstra.QEStringVal("n_system")},
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
//		return "", nil, err
//	}
//
//	var systemNodeId string
//	var candidates []ifmapInfo
//	for _, item := range candidateInterfaceMapsQR.Items {
//		if item.System.Id == "" {
//			return "", nil, fmt.Errorf("graph db search for \"type='system', label='%s'\" found match with empty 'id' field", label)
//		}
//		if systemNodeId != "" && systemNodeId != item.System.Id {
//			return "", nil,
//				fmt.Errorf("graph db search for \"type='system', label='%s'\" found nodes with different 'id' fields: '%s' and '%s'",
//					label, systemNodeId, item.System.Id)
//		}
//		if systemNodeId == "" {
//			systemNodeId = item.System.Id
//		}
//		candidates = append(candidates, ifmapInfo{
//			label:           item.InterfaceMap.Label,
//			id:              item.InterfaceMap.Id,
//			deviceProfileId: item.InterfaceMap.DeviceProfileId,
//		})
//	}
//
//	return systemNodeId, candidates, nil
//}
//
//// ifmapCandidateFromCandidates finds an interface map (by label) within a
//// []ifmapInfo, returns pointer to it, nil if not found.
//func ifmapCandidateFromCandidates(label string, candidates []ifmapInfo) *ifmapInfo {
//	for _, candidate := range candidates {
//		if label == candidate.label {
//			return &candidate
//		}
//	}
//	return nil
//}
//
//func getNodeInterfaceMap(ctx context.Context, client *goapstra.client, bpId goapstra.ObjectId, label string) (*ifmapInfo, error) {
//	var interfaceMapQR struct {
//		Items []struct {
//			InterfaceMap struct {
//				Id              string `json:"id"`
//				Label           string `json:"label"`
//				DeviceProfileId string `json:"device_profile_id"`
//			} `json:"n_interface_map"`
//		} `json:"items"`
//	}
//	err := client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("system")},
//			{"label", goapstra.QEStringVal(label)},
//		}).
//		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("interface_map")},
//			{"name", goapstra.QEStringVal("n_interface_map")},
//		}).
//		Do(&interfaceMapQR)
//	if err != nil {
//		return nil, err
//	}
//	if len(interfaceMapQR.Items) != 1 {
//		return nil, fmt.Errorf("expected exactly one interface map, got %d", len(interfaceMapQR.Items))
//	}
//	return &ifmapInfo{
//		id:              interfaceMapQR.Items[0].InterfaceMap.Id,
//		label:           interfaceMapQR.Items[0].InterfaceMap.Label,
//		deviceProfileId: interfaceMapQR.Items[0].InterfaceMap.DeviceProfileId,
//	}, nil
//}
//
//func getSwitchFromGraphDb(ctx context.Context, client *goapstra.client, bpId goapstra.ObjectId, label string, diags *diag.Diagnostics) *types.Object {
//	// query for system node on its own
//	var systemOnlyQR struct {
//		Items []struct {
//			System struct {
//				NodeId   string `json:"id"`
//				Label    string `json:"label"`
//				SystemID string `json:"system_id"`
//			} `json:"n_system"`
//		} `json:"items"`
//	}
//	err := client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("system")},
//			{"label", goapstra.QEStringVal(label)},
//			{"name", goapstra.QEStringVal("n_system")},
//		}).
//		Do(&systemOnlyQR)
//	if err != nil {
//		diags.AddError("error querying Blueprint node", err.Error())
//		return nil
//	}
//	if len(systemOnlyQR.Items) != 1 {
//		diags.AddError("error querying Blueprint node",
//			fmt.Sprintf("expected exactly one system node, got %d", len(systemOnlyQR.Items)))
//		return nil
//	}
//
//	// query for system node with interface map
//	type interfaceMapQrItem struct {
//		InterfaceMap struct {
//			Id              string `json:"id"`
//			DeviceProfileId string `json:"device_profile_id"`
//		} `json:"n_interface_map"`
//	}
//	var interfaceMapQR struct {
//		Items []interfaceMapQrItem `json:"items"`
//	}
//	err = client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"id", goapstra.QEStringVal(systemOnlyQR.Items[0].System.NodeId)},
//		}).
//		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("interface_map")},
//			{"name", goapstra.QEStringVal("n_interface_map")},
//		}).
//		Do(&interfaceMapQR)
//	if err != nil {
//		diags.AddError("error querying Blueprint node", err.Error())
//		return nil
//	}
//	switch len(interfaceMapQR.Items) {
//	case 0: // slam an empty interfaceMapQrItem{} in there so we have something to read
//		interfaceMapQR.Items = append(interfaceMapQR.Items, interfaceMapQrItem{})
//	case 1: // this is the expected case - no special handling
//	default: // this should never happen
//		diags.AddError(
//			errProviderBug,
//			fmt.Sprintf("node '%s' linked to more than one (%d) interface maps",
//				systemOnlyQR.Items[0].System.NodeId, len(interfaceMapQR.Items)))
//	}
//
//	result := newSwitchObject()
//	result.Attrs["device_key"] = types.String{Value: systemOnlyQR.Items[0].System.SystemID}
//	result.Attrs["system_node_id"] = types.String{Value: systemOnlyQR.Items[0].System.NodeId}
//	result.Attrs["interface_map_id"] = types.String{Value: interfaceMapQR.Items[0].InterfaceMap.Id}
//	result.Attrs["device_profile_id"] = types.String{Value: interfaceMapQR.Items[0].InterfaceMap.DeviceProfileId}
//
//	// flag any result elements with empty value as null
//	for k, _ := range switchElementSchema() {
//		if result.Attrs[k].(types.String).Value == "" {
//			result.Attrs[k] = types.String{Null: true}
//		}
//	}
//
//	return &result
//}
//
//// resourceTypeNameFromResourceGroupName guesses a resource type name
//// (asn/ip/ipv6/possibly others) based on the resource group name. Both type and
//// name are required to uniquely identify a resource group allocation, but so
//// far (fingers crossed) the group names (e.g. "leaf_asns") supply enough of a
//// clue to determine the resource type ("asn"). Using this lookup function saves
//// functions which interact with resource groups from the hassle of keeping
//// track of resource type.
//func resourceTypeNameFromResourceGroupName(in goapstra.ResourceGroupName, diags *diag.Diagnostics) goapstra.ResourceType {
//	switch in {
//	case goapstra.ResourceGroupNameSuperspineAsn:
//		return goapstra.ResourceTypeAsnPool
//	case goapstra.ResourceGroupNameSpineAsn:
//		return goapstra.ResourceTypeAsnPool
//	case goapstra.ResourceGroupNameLeafAsn:
//		return goapstra.ResourceTypeAsnPool
//	case goapstra.ResourceGroupNameAccessAsn:
//		return goapstra.ResourceTypeAsnPool
//
//	case goapstra.ResourceGroupNameSuperspineIp4:
//		return goapstra.ResourceTypeIp4Pool
//	case goapstra.ResourceGroupNameSpineIp4:
//		return goapstra.ResourceTypeIp4Pool
//	case goapstra.ResourceGroupNameLeafIp4:
//		return goapstra.ResourceTypeIp4Pool
//	case goapstra.ResourceGroupNameAccessIp4:
//		return goapstra.ResourceTypeIp4Pool
//
//	case goapstra.ResourceGroupNameSuperspineSpineIp4:
//		return goapstra.ResourceTypeIp4Pool
//	case goapstra.ResourceGroupNameSpineLeafIp4:
//		return goapstra.ResourceTypeIp4Pool
//	case goapstra.ResourceGroupNameLeafLeafIp4:
//		return goapstra.ResourceTypeIp4Pool
//
//	case goapstra.ResourceGroupNameMlagDomainSviSubnets:
//		return goapstra.ResourceTypeIp4Pool
//	case goapstra.ResourceGroupNameAccessAccessIps:
//		return goapstra.ResourceTypeIp4Pool
//	case goapstra.ResourceGroupNameVtepIps:
//		return goapstra.ResourceTypeIp4Pool
//
//	case goapstra.ResourceGroupNameSuperspineSpineIp6:
//		return goapstra.ResourceTypeIp6Pool
//	case goapstra.ResourceGroupNameSpineLeafIp6:
//		return goapstra.ResourceTypeIp6Pool
//	}
//	diags.AddError(errProviderBug, fmt.Sprintf("unable to determine type of resource group '%s'", in))
//	return goapstra.ResourceTypeUnknown
//}
//
//// setResourcePoolElementByTfsdkTag sets value (types.Set) into the named field
//// of the Blueprint object by tfsdk tag
//func (o *Blueprint) setResourcePoolElementByTfsdkTag(fieldName string, value types.Set, diags *diag.Diagnostics) {
//	v := reflect.ValueOf(o).Elem()
//	findTfsdkName := func(t reflect.StructTag) string {
//		if tfsdkTag, ok := t.Lookup("tfsdk"); ok {
//			return tfsdkTag
//		}
//		diags.AddError(errProviderBug, fmt.Sprintf("attempt to lookupg nonexistent tfsdk tag '%s'", fieldName))
//		return ""
//	}
//	fieldNames := map[string]int{}
//	for i := 0; i < v.NumField(); i++ {
//		typeField := v.Type().Field(i)
//		tag := typeField.Tag
//		tname := findTfsdkName(tag)
//		fieldNames[tname] = i
//	}
//
//	fieldNum, ok := fieldNames[fieldName]
//	if !ok {
//		diags.AddError(errProviderBug, fmt.Sprintf("field '%s' does not exist within the provided item", fieldName))
//	}
//	fieldVal := v.Field(fieldNum)
//	fieldVal.Set(reflect.ValueOf(value))
//}

//// readPoolAllocationFromApstraIntoElementByTfsdkTag retrieves a pool
//// allocation from apstra and sets the appropriate element in the
//// Blueprint structure
//func (o *Blueprint) readPoolAllocationFromApstraIntoElementByTfsdkTag(ctx context.Context, tag string, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	rgn := tfsdkTagToRgn(tag, diags)
//	rg := &goapstra.ResourceGroup{
//		Type: resourceTypeNameFromResourceGroupName(rgn, diags),
//		Name: rgn,
//	}
//	if diags.HasError() {
//		return
//	}
//	rga, err := client.GetResourceAllocation(ctx, rg)
//	if err != nil {
//		var ace goapstra.ApstraClientErr
//		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
//			// apstra doesn't know about this resource allocation
//			o.setResourcePoolElementByTfsdkTag(tag, types.Set{Null: true, ElemType: types.StringType}, diags)
//			return
//		}
//		diags.AddError("error getting resource group allocation", err.Error())
//		return
//	}
//
//	poolIds := types.Set{
//		ElemType: types.StringType,
//		Elems:    make([]attr.Value, len(rga.PoolIds)),
//	}
//
//	for i, poolId := range rga.PoolIds {
//		poolIds.Elems[i] = types.String{Value: string(poolId)}
//	}
//
//	o.setResourcePoolElementByTfsdkTag(tag, poolIds, diags)
//}
//
//func (o *Blueprint) updateResourcePoolAllocationByTfsdkTag(ctx context.Context, tag string, client *goapstra.TwoStageL3ClosClient, state *Blueprint, diags *diag.Diagnostics) {
//	planPool := o.extractResourcePoolElementByTfsdkTag(tag, diags)
//	statePool := state.extractResourcePoolElementByTfsdkTag(tag, diags)
//	if diags.HasError() {
//		return
//	}
//
//	if setOfAttrValuesMatch(planPool, statePool) {
//		// no change; set plan = state
//		o.setResourcePoolElementByTfsdkTag(tag, statePool, diags)
//	} else {
//		// edit needed
//		o.setApstraPoolAllocationByTfsdkTag(ctx, tag, client, diags) // use plan to update apstra
//		planPool.Unknown = false                                     // mark planed object as !Unknown
//		o.setResourcePoolElementByTfsdkTag(tag, planPool, diags)     // update plan with planned object
//	}
//}
//
//func (o *Blueprint) readSwitchesFromGraphDb(ctx context.Context, client *goapstra.client, diags *diag.Diagnostics) {
//	// get the list of switch roles (spine1, leaf2...) from the Blueprint
//	switchLabels := listSwitches(ctx, client, goapstra.ObjectId(o.Id.Value), diags)
//	if diags.HasError() {
//		return
//	}
//
//	// collect switch info into newState
//	o.Switches = make(map[string]types.Object, len(switchLabels))
//	for _, switchLabel := range switchLabels {
//		sfgdb := getSwitchFromGraphDb(ctx, client, goapstra.ObjectId(o.Id.Value), switchLabel, diags)
//		if diags.HasError() {
//			return
//		}
//		o.Switches[switchLabel] = *sfgdb
//	}
//}
//
//func parseFabricAddressingPolicy(in types.String, diags *diag.Diagnostics) goapstra.AddressingScheme {
//	if in.IsNull() {
//		return defaultFabricAddressingPolicy
//	}
//	switch in.Value {
//	case goapstra.AddressingSchemeIp4.String():
//		return goapstra.AddressingSchemeIp4
//	case goapstra.AddressingSchemeIp46.String():
//		return goapstra.AddressingSchemeIp46
//	case goapstra.AddressingSchemeIp6.String():
//		return goapstra.AddressingSchemeIp6
//	}
//	diags.AddWarning(errProviderBug, fmt.Sprintf("cannot handle '%s' when parsing fabric addressing policy", in))
//	return -1
//}

//// tfsdkTagToRgn is a simple lookup of tfsdk tag to goapstra.ResourceGroupName.
//// Any lookup misses are a provider bug.
//func tfsdkTagToRgn(tag string, diags *diag.Diagnostics) goapstra.ResourceGroupName {
//	switch tag {
//	case "superspine_asn_pool_ids":
//		return goapstra.ResourceGroupNameSuperspineAsn
//	case "spine_asn_pool_ids":
//		return goapstra.ResourceGroupNameSpineAsn
//	case "leaf_asn_pool_ids":
//		return goapstra.ResourceGroupNameLeafAsn
//	case "access_asn_pool_ids":
//		return goapstra.ResourceGroupNameAccessAsn
//	case "superspine_loopback_pool_ids":
//		return goapstra.ResourceGroupNameSuperspineIp4
//	case "spine_loopback_pool_ids":
//		return goapstra.ResourceGroupNameSpineIp4
//	case "leaf_loopback_pool_ids":
//		return goapstra.ResourceGroupNameLeafIp4
//	case "access_loopback_pool_ids":
//		return goapstra.ResourceGroupNameAccessIp4
//	case "superspine_spine_ip4_pool_ids":
//		return goapstra.ResourceGroupNameSuperspineSpineIp4
//	case "spine_leaf_ip4_pool_ids":
//		return goapstra.ResourceGroupNameSpineLeafIp4
//	case "leaf_leaf_ip4_pool_ids":
//		return goapstra.ResourceGroupNameLeafLeafIp4
//	case "leaf_mlag_peer_link_ip4_pool_ids":
//		return goapstra.ResourceGroupNameMlagDomainSviSubnets
//	case "access_esi_peer_link_ip4_pool_ids":
//		return goapstra.ResourceGroupNameAccessAccessIps
//	case "vtep_ip4_pool_ids":
//		return goapstra.ResourceGroupNameVtepIps
//	case "superspine_spine_ip6_pool_ids":
//		return goapstra.ResourceGroupNameSuperspineSpineIp6
//	case "spine_leaf_ip6_pool_ids":
//		return goapstra.ResourceGroupNameSpineLeafIp6
//	}
//	diags.AddError(errProviderBug, fmt.Sprintf("tfsdk tag '%s' unknown", tag))
//	return goapstra.ResourceGroupNameUnknown
//}
//
//// rgnToTfsdkTag is a simple lookup from goapstraResourceGroupName to the tfsdk
//// tag used to represent it.
//// Any lookup misses are a provider bug.
//func rgnToTfsdkTag(rgn goapstra.ResourceGroupName, diags *diag.Diagnostics) string {
//	switch rgn {
//	case goapstra.ResourceGroupNameSuperspineAsn:
//		return "superspine_asn_pool_ids"
//	case goapstra.ResourceGroupNameSpineAsn:
//		return "spine_asn_pool_ids"
//	case goapstra.ResourceGroupNameLeafAsn:
//		return "leaf_asn_pool_ids"
//	case goapstra.ResourceGroupNameAccessAsn:
//		return "access_asn_pool_ids"
//	case goapstra.ResourceGroupNameSuperspineIp4:
//		return "superspine_loopback_pool_ids"
//	case goapstra.ResourceGroupNameSpineIp4:
//		return "spine_loopback_pool_ids"
//	case goapstra.ResourceGroupNameLeafIp4:
//		return "leaf_loopback_pool_ids"
//	case goapstra.ResourceGroupNameAccessIp4:
//		return "access_loopback_pool_ids"
//	case goapstra.ResourceGroupNameSuperspineSpineIp4:
//		return "superspine_spine_ip4_pool_ids"
//	case goapstra.ResourceGroupNameSpineLeafIp4:
//		return "spine_leaf_ip4_pool_ids"
//	case goapstra.ResourceGroupNameLeafLeafIp4:
//		return "leaf_leaf_ip4_pool_ids"
//	case goapstra.ResourceGroupNameMlagDomainSviSubnets:
//		return "leaf_mlag_peer_link_ip4_pool_ids"
//	case goapstra.ResourceGroupNameAccessAccessIps:
//		return "access_esi_peer_link_ip4_pool_ids"
//	case goapstra.ResourceGroupNameVtepIps:
//		return "vtep_ip4_pool_ids"
//	case goapstra.ResourceGroupNameSuperspineSpineIp6:
//		return "superspine_spine_ip6_pool_ids"
//	case goapstra.ResourceGroupNameSpineLeafIp6:
//		return "spine_leaf_ip6_pool_ids"
//	}
//	diags.AddError(errProviderBug, fmt.Sprintf("resource group name '%s' unknown", rgn.String()))
//	return ""
//}
//
//func warnMissingResourceGroupAllocations(ctx context.Context, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	allocations, err := client.GetResourceAllocations(ctx)
//	if err != nil {
//		diags.AddError("error fetching resource group allocations", err.Error())
//		return
//	}
//
//	var missing []string
//	for _, allocation := range allocations {
//		if allocation.IsEmpty() {
//			missing = append(missing, fmt.Sprintf("%q", rgnToTfsdkTag(allocation.ResourceGroup.Name, diags)))
//		}
//	}
//	if len(missing) != 0 {
//		diags.AddWarning(warnMissingResourceSummary, fmt.Sprintf(warnMissingResourceDetail, strings.Join(missing, ", ")))
//	}
//}
//
//// listOfResourceGroupAllocationTags returns the full list of tfsdk tags
//// representing potential resource group allocations for a "datacenter"
//// Blueprint. This could probably be rewritten as a "reflect" operation against
//// an Blueprint which extracts tags ending in "_pool_ids".
//func listOfResourceGroupAllocationTags() []string {
//	return []string{
//		"superspine_asn_pool_ids",
//		"spine_asn_pool_ids",
//		"leaf_asn_pool_ids",
//		"access_asn_pool_ids",
//		"superspine_loopback_pool_ids",
//		"spine_loopback_pool_ids",
//		"leaf_loopback_pool_ids",
//		"access_loopback_pool_ids",
//		"superspine_spine_ip4_pool_ids",
//		"spine_leaf_ip4_pool_ids",
//		"leaf_leaf_ip4_pool_ids",
//		"leaf_mlag_peer_link_ip4_pool_ids",
//		"access_esi_peer_link_ip4_pool_ids",
//		"vtep_ip4_pool_ids",
//		"superspine_spine_ip6_pool_ids",
//		"spine_leaf_ip6_pool_ids",
//	}
//}
//
//func newSwitchObject() types.Object {
//	return types.Object{
//		AttrTypes: switchElementSchema(),
//		Attrs:     make(map[string]attr.Value, len(switchElementSchema())),
//	}
//}
//
//func switchElementSchema() map[string]attr.Type {
//	return map[string]attr.Type{
//		"interface_map_id":  types.StringType,
//		"device_key":        types.StringType,
//		"device_profile_id": types.StringType,
//		"system_node_id":    types.StringType,
//	}
//}

//// listSwitches returns a []string enumerating switch roles (spine1,
//// leaf2_1, etc...) in the indicated Blueprint
//func listSwitches(ctx context.Context, client *goapstra.client, bpId goapstra.ObjectId, diags *diag.Diagnostics) []string {
//	var switchQr struct {
//		Items []struct {
//			System struct {
//				Label string `json:"label"`
//				Id    string `json:"id"`
//			} `json:"n_system"`
//		} `json:"items"`
//	}
//	err := client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("system")},
//			{"name", goapstra.QEStringVal("n_system")},
//			{"system_type", goapstra.QEStringVal("switch")},
//		}).
//		Do(&switchQr)
//	if err != nil {
//		diags.AddError("error querying graphdb for switch nodes", err.Error())
//		return nil
//	}
//
//	result := make([]string, len(switchQr.Items))
//	for i := range switchQr.Items {
//		result[i] = switchQr.Items[i].System.Label
//	}
//	return result
//}

func (o *resourceDatacenterBlueprint) apiVersion() (*version.Version, error) {
	if o.client == nil {
		return nil, nil
	}
	return version.NewVersion(o.client.ApiVersion())
}

func (o *resourceDatacenterBlueprint) cfgVersionMin() (*version.Version, error) {
	return o.minClientVersion, nil
}

func (o *resourceDatacenterBlueprint) cfgVersionMax() (*version.Version, error) {
	return o.maxClientVersion, nil
}

func (o *resourceDatacenterBlueprint) checkVersion(ctx context.Context, diags *diag.Diagnostics) {
	checkVersionCompatibility(ctx, o, diags)
}
