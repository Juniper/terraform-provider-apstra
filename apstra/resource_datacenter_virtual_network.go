package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterVirtualNetwork{}
var _ resource.ResourceWithModifyPlan = &resourceDatacenterVirtualNetwork{}
var _ resource.ResourceWithValidateConfig = &resourceDatacenterVirtualNetwork{}

type resourceDatacenterVirtualNetwork struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceDatacenterVirtualNetwork) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network"
}

func (o *resourceDatacenterVirtualNetwork) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDatacenterVirtualNetwork) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Virtual Network within a Blueprint.",
		Attributes:          blueprint.DatacenterVirtualNetwork{}.ResourceAttributes(),
	}
}

// ValidateConfig ensures that when reserve_vlan is true, all vlan bindings are
// set and match each other.
func (o *resourceDatacenterVirtualNetwork) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validation only possible when reserve_vlan is set "true"
	if !config.ReserveVlan.ValueBool() {
		return // skip 'false', 'unknown', 'null' values
	}

	// validation not possible when bindings are unknown
	if config.Bindings.IsUnknown() {
		return
	}

	// validation not possible when any individual binding is unknown
	for _, v := range config.Bindings.Elements() {
		if v.IsUnknown() {
			return
		}
	}

	// extract bindings as a map
	var bindings map[string]blueprint.VnBinding
	resp.Diagnostics.Append(config.Bindings.ElementsAs(ctx, &bindings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validate each binding
	invalidConfigDueToNullVlan := false
	reservedVlanIds := make(map[int64]struct{})
	for _, binding := range bindings {
		if binding.VlanId.IsUnknown() {
			continue // skip further validation of unknown vlans
		}
		if binding.VlanId.IsNull() {
			invalidConfigDueToNullVlan = true
			continue
		}
		reservedVlanIds[binding.VlanId.ValueInt64()] = struct{}{}
	}

	// null vlan is invalid
	if invalidConfigDueToNullVlan {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			path.Root("bindings"),
			"'vlan_id' must be specified in each binding when 'reserve_vlan' is true"))
	}

	// we should have at most one VLAN ID across all bindings (zero when they're unknown)
	if len(reservedVlanIds) > 1 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			path.Root("bindings"),
			"'vlan_id' attributes must match when 'reserve_vlan' is true"))
		return
	}
}

func (o *resourceDatacenterVirtualNetwork) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// This plan modifier solves the same problem for two different
	// `Optional` + `Computed` attributes:
	//   - VlanId
	//   - Vni
	//
	// The problem is terraform's ordinary handling of `Optional` + `Computed`
	// attributes:
	//
	// https://discuss.hashicorp.com/t/schema-for-optional-computed-to-support-correct-removal-plan-in-framework/49055/5?u=hqnvylrx
	//
	//   The subject of what goes on behind the scenes of Terraform plan with
	//   regards to providers is pretty nuanced. Without going too much into the
	//   weeds, the behavior for Terraform for Optional + Computed attributes is
	//   to copy the prior state if there is no configuration for it.
	//
	// This means that a manually-configured VLAN ID or VNI won't get backed-out
	// via the API to allow Apstra to choose a new value from its pool.
	//
	// We work around that behavior by using trigger/tracker `Computed` boolean
	// attributes for each `Computed` + `Optional` resource:
	//   - HadPriorVlanIdConfig
	//   - HadPriorVniConfig
	//
	// Whenever these attributes are found `true`, but the corresponding config
	// element is `null`, we conclude that the attribute been removed from the
	// configuration and set the attribute to `unknown` to achieve modification
	// and record a new choice made by the API.

	// No state means there couldn't have been a previous config.
	// No plan means we're doing Delete().
	// Both are un-interesting to this plan modifier.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Retrieve values from config
	var config blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Retrieve values from plan
	var plan blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// null config with prior configured value means vni was removed
	if config.Vni.IsNull() && state.HadPriorVniConfig.ValueBool() {
		plan.Vni = types.Int64Unknown()
		plan.HadPriorVniConfig = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (o *resourceDatacenterVirtualNetwork) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client for the datacenter reference design
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, plan.BlueprintId), err.Error())
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

	// create a request object
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the virtual network
	id, err := bp.CreateVirtualNetwork(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating virtual network", err.Error())
		return
	}

	// update the plan with the received ObjectId and set the partial state in
	// case we have to bail due to error soon.
	plan.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	// fetch the virtual network to learn apstra-assigned VLAN assignments
	api, err := bp.GetVirtualNetwork(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error fetching just-created virtual network %q", id),
			err.Error())
		return
	}

	// Create a new state object and load the current state from the API. We're
	// instantiating a new object here because #170 (a creation race condition
	// in the API) means we can't completely rely on the API response.
	var state blueprint.DatacenterVirtualNetwork
	state.Id = types.StringValue(id.String())
	state.BlueprintId = plan.BlueprintId
	state.HadPriorVniConfig = plan.HadPriorVniConfig
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)

	// Don't rely on the API response for these values (#170). If the config
	// supplied a value, use it when setting state.
	if !plan.IPv4Subnet.IsUnknown() {
		state.IPv4Subnet = plan.IPv4Subnet
	}
	if !plan.IPv6Subnet.IsUnknown() {
		state.IPv6Subnet = plan.IPv6Subnet
	}
	if !plan.IPv4Gateway.IsUnknown() {
		state.IPv4Gateway = plan.IPv4Gateway
	}
	if !plan.IPv6Gateway.IsUnknown() {
		state.IPv6Gateway = plan.IPv6Gateway
	}
	if !plan.Vni.IsUnknown() {
		state.Vni = plan.Vni
	}
	if !plan.ReserveVlan.IsUnknown() {
		state.ReserveVlan = plan.ReserveVlan
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (o *resourceDatacenterVirtualNetwork) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client for the datacenter reference design
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
		return
	}

	// retrieve the virtual network
	vn, err := bp.GetVirtualNetwork(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
	}

	// load the API response and set the state
	state.LoadApiData(ctx, vn.Data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (o *resourceDatacenterVirtualNetwork) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client for the datacenter reference design
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, plan.BlueprintId), err.Error())
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

	// create a request object
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the virtual network according to the plan
	err = bp.UpdateVirtualNetwork(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating virtual network", err.Error())
	}

	// fetch the virtual network to learn apstra-assigned VLAN assignments
	api, err := bp.GetVirtualNetwork(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error fetching just-updated virtual network %q", plan.Id.ValueString()),
			err.Error())
		return
	}

	// Create a new state object and load the current state from the API. We're
	// instantiating a new object here because #170 (a creation race condition
	// in the API) means we can't completely rely on the API response.
	var state blueprint.DatacenterVirtualNetwork
	state.Id = plan.Id
	state.BlueprintId = plan.BlueprintId
	state.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)

	// Don't rely on the API response for these values (#170). If the config
	// supplied a value, use that when setting state.
	if !plan.IPv4Subnet.IsUnknown() {
		state.IPv4Subnet = plan.IPv4Subnet
	}
	if !plan.IPv6Subnet.IsUnknown() {
		state.IPv6Subnet = plan.IPv6Subnet
	}
	if !plan.IPv4Gateway.IsUnknown() {
		state.IPv4Gateway = plan.IPv4Gateway
	}
	if !plan.IPv6Gateway.IsUnknown() {
		state.IPv6Gateway = plan.IPv6Gateway
	}
	if !plan.ReserveVlan.IsUnknown() {
		state.ReserveVlan = plan.ReserveVlan
	}

	// if the plan modifier didn't take action...
	if plan.HadPriorVniConfig.IsUnknown() {
		// ...then the trigger value is set according to whether a VNI value is known.
		state.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())
	} else {
		state.HadPriorVniConfig = plan.HadPriorVniConfig
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (o *resourceDatacenterVirtualNetwork) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterVirtualNetwork
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a client for the datacenter reference design
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
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

	// Delete the virtual network
	err = bp.DeleteVirtualNetwork(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting virtual network", err.Error())
	}
}
