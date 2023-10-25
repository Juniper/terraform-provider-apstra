package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
)

var _ resource.ResourceWithConfigure = &resourceDeviceAllocation{}

type resourceDeviceAllocation struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceDeviceAllocation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_device_allocation"
}

func (o *resourceDeviceAllocation) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDeviceAllocation) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource allocates a Managed Device (probably a switch) to a node role" +
			" (spine1, etc...) within a Blueprint.",
		Attributes: blueprint.DeviceAllocation{}.ResourceAttributes(),
	}
}

func (o *resourceDeviceAllocation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure the blueprint exists.
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(plan.BlueprintId.ValueString()), &resp.Diagnostics) {
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.AddError("no such blueprint", fmt.Sprintf("blueprint %s not found", plan.BlueprintId))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Ensure the following are populated:
	//   - SystemNodeId (from node_name)
	//   - SystemNodeId
	//   - InitialInterfaceMapId
	//   - DeviceProfileNodeId
	plan.PopulateDataFromGraphDb(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.BlueprintId.IsNull() {
		resp.Diagnostics.AddError("blueprint does not exist", "blueprint vanished while we were working on it")
	}

	plan.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)
	if plan.BlueprintId.IsNull() {
		resp.Diagnostics.AddError("blueprint does not exist", "blueprint vanished while we were working on it")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DeviceKey.IsNull() {
		plan.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	plan.SetNodeDeployMode(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDeviceAllocation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy details from the state so we can look for changes due to FFE
	previousInterfaceMapCatalogId := state.InitialInterfaceMapId
	previousInterfaceMapName := state.InterfaceMapName

	// Ensure the blueprint still exists.
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		if resp.Diagnostics.HasError() {
			return
		}
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.GetDeviceKey(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() { // not found?
		resp.State.RemoveResource(ctx)
		return
	}

	state.GetCurrentInterfaceMapId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() { // not found?
		resp.State.RemoveResource(ctx)
		return
	}

	state.GetCurrentDeviceProfileId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() { // not found?
		resp.State.RemoveResource(ctx)
		return
	}

	state.GetNodeDeployMode(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.GetInterfaceMapName(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// special handling for FFE gyrations. The interface map ID might change,
	// but we shouldn't surface that difference in Read() if the interface map
	// map label (web UI "name") suggests the ID change is due to FFE.
	if !state.InitialInterfaceMapId.Equal(previousInterfaceMapCatalogId) {
		// Interface map ID in blueprint doesn't match the one used to create it.
		// Is it a manual change or the result of an FFE event?
		// Based on `aos/reference_design/fabric_expansion_util.py`:
		//     regex = '^(.+?)(_v([0-9]+))?$'
		// Note that the total name length is limited to 64 characters. Long
		// names are trimmed down to 64. The trimming happens in the chunk
		// preceding "_v[0-9]".
		nameRE := regexp.MustCompile(fmt.Sprintf("^%s_v[0-9]+$", previousInterfaceMapName.ValueString()))
		if nameRE.MatchString(state.InterfaceMapName.ValueString()) {
			// The change of InitialInterfaceMapId seems to be due to FFE
			state.InitialInterfaceMapId = types.StringValue(previousInterfaceMapCatalogId.ValueString())
		}
	}

	// InterfaceMapName must be immutable in order to be useful in detecting FFE modifications
	state.InterfaceMapName = types.StringValue(previousInterfaceMapName.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDeviceAllocation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	state.DeployMode = plan.DeployMode
	state.SetNodeDeployMode(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DeviceKey.Equal(state.DeviceKey) {
		// device key has changed
		state.DeviceKey = plan.DeviceKey // copy user input directly from plan
		state.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDeviceAllocation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No need to proceed if the blueprint no longer exists
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	state.InitialInterfaceMapId = types.StringNull()
	state.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)

	state.DeviceKey = types.StringNull() // 'null' triggers clearing the 'system_id' field.
	state.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
}
