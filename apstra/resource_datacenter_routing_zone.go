package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceDatacenterRoutingZone{}
	_ resource.ResourceWithModifyPlan     = &resourceDatacenterRoutingZone{}
	_ resource.ResourceWithValidateConfig = &resourceDatacenterRoutingZone{}
	_ resourceWithSetDcBpClientFunc       = &resourceDatacenterRoutingZone{}
	_ resourceWithSetBpLockFunc           = &resourceDatacenterRoutingZone{}
	_ resourceWithSetClient               = &resourceDatacenterRoutingZone{}
)

type resourceDatacenterRoutingZone struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterRoutingZone) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zone"
}

func (o *resourceDatacenterRoutingZone) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterRoutingZone) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a Routing Zone within a Datacenter Blueprint.",
		Attributes:          blueprint.DatacenterRoutingZone{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterRoutingZone) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// config-only validation begins here

	// config + api version validation begins here

	// cannot proceed to config + api version validation if the provider has not been configured
	if o.client == nil {
		return
	}

	apiVersion, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse API version %q", o.client.ApiVersion()), err.Error())
		return
	}

	// validate the configuration
	constraints := config.VersionConstraints(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		compatibility.ValidateConfigConstraints(
			ctx,
			compatibility.ValidateConfigConstraintsRequest{
				Version:     apiVersion,
				Constraints: constraints,
			},
		)...,
	)
}

func (o *resourceDatacenterRoutingZone) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
	// Whenever these "prior" attributes are found `true` and the corresponding
	// config element is `null`, we can conclude that the attribute has just
	// been removed from the configuration and set the attribute to `unknown` to
	// achieve modification and record a new choice made by the API.

	// No state means there couldn't have been a previous config.
	// No plan means we're doing Delete().
	// Both are un-interesting to this plan modifier.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Retrieve values from config
	var config blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan
	var plan blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// null config with prior configured value means vlan id was removed
	if config.VlanId.IsNull() && state.HadPriorVlanIdConfig.ValueBool() {
		plan.VlanId = types.Int64Unknown()
		plan.HadPriorVlanIdConfig = types.BoolValue(false)
	}

	// null config with prior configured value means vni was removed
	if config.Vni.IsNull() && state.HadPriorVniConfig.ValueBool() {
		plan.Vni = types.Int64Unknown()
		plan.HadPriorVniConfig = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingZone) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterRoutingZone
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

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// make a security zone request
	plan.VrfName = plan.Name // copy whatever the user set as name in to VrfName
	request := plan.Request(ctx, bp.Client(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the security zone
	id, err := bp.CreateSecurityZone(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating security zone", err.Error())
		return
	}

	// record the new security zone ID
	plan.Id = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if !plan.DhcpServers.IsNull() {
		dhcpServerRequest := plan.DhcpServerRequest(ctx, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		err = bp.SetSecurityZoneDhcpServers(ctx, id, dhcpServerRequest)
		if err != nil {
			resp.Diagnostics.AddError("error setting security zone dhcp servers", err.Error())
			return
		}
	}

	// Set prior config trackers according to whether a value was planned. Must be done before plan.Read()
	plan.HadPriorVlanIdConfig = types.BoolValue(!plan.VlanId.IsUnknown())
	plan.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())

	// read any apstra-assigned values associated with the new routing zone
	err = plan.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("failed while fetching detail of just-created Routing Zone", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingZone) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	// Create "newState" with zero values (null) to ensure that the Read() method doesn't short circuit.
	var newState blueprint.DatacenterRoutingZone
	newState.Id = state.Id
	newState.HadPriorVlanIdConfig = state.HadPriorVlanIdConfig
	newState.HadPriorVniConfig = state.HadPriorVniConfig

	// read the current status from the API
	err = newState.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("failed while reading blueprint %s routing zone %s details", bp.Id(), state.Id),
			err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceDatacenterRoutingZone) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state.
	var state blueprint.DatacenterRoutingZone
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

	// create a request we'll use when invoking UpdateSecurityZone
	request := plan.Request(ctx, bp.Client(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set new "prior" markers
	plan.HadPriorVlanIdConfig = types.BoolValue(utils.HasValue(plan.VlanId))
	plan.HadPriorVniConfig = types.BoolValue(utils.HasValue(plan.Vni))

	// send the update
	err = bp.UpdateSecurityZone(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error updating security zone", err.Error())
		return
	}

	// update DHCP server list if necessary
	if !plan.DhcpServers.Equal(state.DhcpServers) {
		dhcpRequest := plan.DhcpServerRequest(ctx, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		err = bp.SetSecurityZoneDhcpServers(ctx, plan.Id.ValueString(), dhcpRequest)
		if err != nil {
			resp.Diagnostics.AddError("error updating security zone dhcp servers", err.Error())
			return
		}
	}

	// collect any values calculated by apstra
	err = plan.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("failed while updating routing zone", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingZone) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterRoutingZone
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

	// Delete the routing zone
	err = bp.DeleteSecurityZone(ctx, state.Id.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting routing zone", err.Error())
	}
}

func (o *resourceDatacenterRoutingZone) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterRoutingZone) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceDatacenterRoutingZone) setClient(client *apstra.Client) {
	o.client = client
}
