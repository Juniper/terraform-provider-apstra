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
)

var _ resource.ResourceWithConfigure = &resourceDatacenterGenericSystem{}

type resourceDatacenterGenericSystem struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceDatacenterGenericSystem) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_generic_system"
}

func (o *resourceDatacenterGenericSystem) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDatacenterGenericSystem) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Generic System within a Datacenter Blueprint.",
		Attributes:          blueprint.DatacenterGenericSystem{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterGenericSystem) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterGenericSystem
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
			fmt.Sprintf("failed to lock blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// prep a generic system creation request
	request := plan.CreateRequest(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// unfortunately we only learn the link IDs, not the generic system ID
	linkIds, err := bp.CreateLinksWithNewServer(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("failed to create generic system", err.Error())
		return
	}

	// use link IDs to learn the generic system ID
	genericSystemId, err := bp.SystemNodeFromLinkIds(ctx, linkIds, apstra.SystemNodeRoleGeneric)
	if err != nil {
		resp.Diagnostics.AddError("failed to determine new generic system ID from links", err.Error())
	}
	plan.Id = types.StringValue(genericSystemId.String())

	// set generic system properties sending <nil> for prior state
	plan.SetProperties(ctx, bp, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// pull Apstra-generated strings as needed
	err = plan.ReadSystemProperties(ctx, bp, false)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to retrieve properties from new generic system %s", plan.Id), err.Error())
		// don't return here - still want to set the state
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a blueprint client
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
		return
	}

	// Read various fields using the web UI's system API endpoint. This has a
	// side effect of discovering whether the generic system has been deleted.
	err = state.ReadSystemProperties(ctx, bp, true)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to retrieve generic system %s", state.Id), err.Error())
		return
	}

	// read tags and link info using other combinations of API calls
	state.ReadTags(ctx, bp, &resp.Diagnostics)
	state.ReadLinks(ctx, bp, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterGenericSystem) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a blueprint client
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, plan.BlueprintId), err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to lock blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	plan.UpdateHostnameAndName(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set generic system properties using prior state to skip unnecessary API calls
	plan.SetProperties(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.UpdateTags(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.UpdateLinkSet(ctx, &state, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
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
			fmt.Sprintf("failed to lock blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Delete generic system
	err = bp.DeleteGenericSystem(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to delete generic system", err.Error())
	}
}
