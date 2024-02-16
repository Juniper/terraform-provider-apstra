package tfapstra

import (
	"context"
	"errors"
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
var _ resource.ResourceWithValidateConfig = &resourceDeviceAllocation{}
var _ resourceWithSetBpClientFunc = &resourceDeviceAllocation{}
var _ resourceWithSetBpLockFunc = &resourceDeviceAllocation{}
var _ resourceWithSetExperimental = &resourceDeviceAllocation{}

type resourceDeviceAllocation struct {
	experimental    types.Bool
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDeviceAllocation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_device_allocation"
}

func (o *resourceDeviceAllocation) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDeviceAllocation) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config.
	var config blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ValidateConfig(ctx, o.experimental, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
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

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// if the user gave us system attributes, make sure that we're pointed at a switch
	if !plan.SystemAttributes.IsUnknown() {
		plan.EnsureSystemIsSwitchBeforeCreate(ctx, bp, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Ensure the following are populated:
	//   - SystemNodeId (from node_name)
	//   - InitialInterfaceMapId
	//   - DeviceProfileNodeId
	//   - InterfaceMapName
	plan.PopulateDataFromGraphDb(ctx, bp.Client(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.BlueprintId.IsNull() {
		resp.Diagnostics.AddError("blueprint does not exist", "blueprint vanished while we were working on it")
	}

	plan.SetInterfaceMap(ctx, bp, &resp.Diagnostics)
	if plan.BlueprintId.IsNull() {
		resp.Diagnostics.AddError("blueprint does not exist", "blueprint vanished while we were working on it")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DeviceKey.IsNull() {
		plan.SetNodeSystemId(ctx, bp.Client(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	switch {
	case plan.DeployMode.IsNull(): // not expected with Optional+Computed, nothing to do here
	case plan.DeployMode.IsUnknown(): // config is null, get the Computed value
		deployMode, err := utils.GetNodeDeployMode(ctx, bp, plan.NodeId.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to fetch node deploy mode", err.Error())
			return
		}
		plan.DeployMode = types.StringValue(deployMode)
	default: // value provided via config
		err = utils.SetNodeDeployMode(ctx, bp, plan.NodeId.ValueString(), plan.DeployMode.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed while setting node deploy mode", err.Error())
			return
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set system attributes
	plan.SetSystemAttributes(ctx, nil, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// read back any apstra-assigned attributes
	plan.GetSystemAttributes(ctx, bp, &resp.Diagnostics)
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

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	state.GetDeviceKey(ctx, bp.Client(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() { // not found?
		resp.State.RemoveResource(ctx)
		return
	}

	state.GetCurrentInterfaceMapId(ctx, bp.Client(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() { // not found?
		resp.State.RemoveResource(ctx)
		return
	}

	state.GetCurrentDeviceProfileId(ctx, bp.Client(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() { // not found?
		resp.State.RemoveResource(ctx)
		return
	}

	deployMode, err := utils.GetNodeDeployMode(ctx, bp, state.NodeId.ValueString())
	if err != nil {
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
	}
	state.DeployMode = types.StringValue(deployMode)

	state.GetInterfaceMapName(ctx, bp.Client(), &resp.Diagnostics)
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

	// read any apstra-assigned values
	state.GetSystemAttributes(ctx, bp, &resp.Diagnostics)

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

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	switch {
	case plan.DeployMode.IsNull(): // not expected with Optional+Computed, nothing to do here
	case plan.DeployMode.IsUnknown(): // config is null, get the Computed value
		deployMode, err := utils.GetNodeDeployMode(ctx, bp, plan.NodeId.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed reading node deploy mode", err.Error())
			return
		}

		plan.DeployMode = types.StringValue(deployMode)
	default: // value provided via config
		err := utils.SetNodeDeployMode(ctx, bp, plan.NodeId.ValueString(), plan.DeployMode.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed setting node deploy mode", err.Error())
			return
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	state.DeployMode = plan.DeployMode

	if !plan.DeviceKey.Equal(state.DeviceKey) {
		// device key has changed
		state.DeviceKey = plan.DeviceKey // copy user input directly from plan
		state.SetNodeSystemId(ctx, bp.Client(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// update the system attributes as necessary
	plan.SetSystemAttributes(ctx, &state, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy the planned system attributes into the state
	state.SystemAttributes = plan.SystemAttributes

	// read any apstra-assigned values
	state.GetSystemAttributes(ctx, bp, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDeviceAllocation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	state.InitialInterfaceMapId = types.StringNull()
	state.SetInterfaceMap(ctx, bp, &resp.Diagnostics)

	state.DeviceKey = types.StringNull() // 'null' triggers clearing the 'system_id' field.
	state.SetNodeSystemId(ctx, bp.Client(), &resp.Diagnostics)
}

func (o *resourceDeviceAllocation) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDeviceAllocation) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceDeviceAllocation) setExperimental(b bool) {
	o.experimental = types.BoolValue(b)
}
