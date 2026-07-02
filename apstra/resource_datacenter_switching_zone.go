package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceDatacenterSwitchingZone{}
	//_ resource.ResourceWithImportState    = &resourceDatacenterSwitchingZone{}
	_ resource.ResourceWithModifyPlan     = &resourceDatacenterSwitchingZone{}
	_ resource.ResourceWithValidateConfig = &resourceDatacenterSwitchingZone{}
	_ resourceWithSetDcBpClientFunc       = &resourceDatacenterSwitchingZone{}
	_ resourceWithSetBpLockFunc           = &resourceDatacenterSwitchingZone{}
	_ resourceWithSetClient               = &resourceDatacenterSwitchingZone{}
)

type resourceDatacenterSwitchingZone struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterSwitchingZone) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_switching_zone"
}

func (o *resourceDatacenterSwitchingZone) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterSwitchingZone) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a Switching Zone within a Datacenter Blueprint. " +
			"Requires Apstra " + compatibility.SwitchingZoneOK.String() + ".",
		Attributes: blueprint.DatacenterSwitchingZone{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterSwitchingZone) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterSwitchingZone
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

	if !compatibility.SwitchingZoneOK.Check(apiVersion) {
		resp.Diagnostics.AddError(
			"Resource requires Apstra "+compatibility.SwitchingZoneOK.String(),
			"Resource requires Apstra "+compatibility.SwitchingZoneOK.String(),
		)
		return
	}
}

func (o *resourceDatacenterSwitchingZone) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// This plan modifier updates the route_target attribute when the user nullifies (comments out)
	// a value they'd previously set.
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
	// This means that a manually-configured route_target won't get backed-out
	// via the API to allow Apstra to choose a new value.
	//
	// The desired behavior is for Apstra to select a new value when the user
	// abandons a value they're previously configured.
	//
	// The Create() and Update() methods record whether or not a user-assigned
	// value was present in private state. We check that value. If true, but
	// the current config is `null`, we set the planned value to Unknown.

	// No state means there couldn't have been a previous config.
	// No plan means we're doing Delete().
	// Both are un-interesting to this plan modifier.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Retrieve values from config
	var config blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan
	var plan blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ps private.ResourceDatacenterSwitchingZone
	ps.LoadPrivateState(ctx, req.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// null config with prior configured value means route_target was removed.
	if config.RouteTarget.IsNull() && ps.HasRouteTarget {
		plan.RouteTarget = customtypes.NewRouteTargetUnknown() // update the plan
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...) // set the plan
}

func (o *resourceDatacenterSwitchingZone) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan.
	var plan blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintID), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintID.ValueString()),
			err.Error())
		return
	}

	// make a switching zone request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the switching zone
	id, err := bp.CreateSwitchingZone(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating switching zone", err.Error())
		return
	}

	// record the new switching zone ID
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	// record private state
	private.ResourceDatacenterSwitchingZone{HasRouteTarget: utils.HasValue(config.RouteTarget)}.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)

	// read any apstra-assigned values associated with the new switching zone
	plan.ReadComputedValues(ctx, bp, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterSwitchingZone) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// retrieve the switching zone
	sz, err := bp.GetSwitchingZone(ctx, state.ID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to retrieve Switching Zone %s", state.ID), err.Error())
		return
	}

	// load the result and set the state
	state.LoadApiData(ctx, sz, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterSwitchingZone) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan.
	var plan blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintID.ValueString()),
			err.Error())
		return
	}

	// create a request we'll use when invoking UpdateSwitchingZone
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//// set new "prior" markers
	//plan.HadPriorVlanIdConfig = types.BoolValue(utils.HasValue(plan.VlanId))
	//plan.HadPriorVniConfig = types.BoolValue(utils.HasValue(plan.Vni))

	// send the update
	err = bp.UpdateSwitchingZone(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error updating switching zone", err.Error())
		return
	}

	// record private state
	private.ResourceDatacenterSwitchingZone{HasRouteTarget: utils.HasValue(config.RouteTarget)}.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)

	// collect any values calculated by apstra
	plan.ReadComputedValues(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterSwitchingZone) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintID.ValueString()),
			err.Error())
		return
	}

	// Delete the switching zone
	err = bp.DeleteSwitchingZone(ctx, state.ID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting switching zone", err.Error())
	}
}

func (o *resourceDatacenterSwitchingZone) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterSwitchingZone) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceDatacenterSwitchingZone) setClient(client *apstra.Client) {
	o.client = client
}
