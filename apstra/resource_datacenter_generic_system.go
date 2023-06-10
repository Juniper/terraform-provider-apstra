package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
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
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan.
	var plan blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client for the datacenter reference design
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
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

	// prep a generic system creation request
	request := plan.CreateRequest(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// unfortunately we only learn the link IDs, not the generic system ID
	linkIds, err := bp.CreateLinksWithNewServer(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating generic system", err.Error())
		return
	}

	// use link IDs to learn the generic system ID
	genericSystemId, err := bp.SystemNodeFromLinkIds(ctx, linkIds, apstra.SystemNodeRoleGeneric)
	if err != nil {
		resp.Diagnostics.AddError("failed to determine new generic system ID from links", err.Error())
	}
	plan.Id = types.StringValue(genericSystemId.String())

	// pull Apstra-generated strings if not specified by the user
	if plan.Label.IsNull() || plan.Hostname.IsNull() {
		err = plan.GetLabelAndHostname(ctx, bp)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error parsing labels from new generic server %s", plan.Id), err.Error())
			// don't return here - still want to set the state
		}
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a blueprint client
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
		return
	}

	// read string fields directly from the graph node. This has the side
	// effect of discovering whether the generic system has been deleted.
	err = state.GetLabelAndHostname(ctx, bp)
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed reading labels from generic server %s", state.Id), err.Error())
		return
	}

	// read tags and link info using other combinations of API calls
	state.ReadTags(ctx, bp, &resp.Diagnostics)
	state.ReadLinks(ctx, bp, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterGenericSystem) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

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
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	plan.UpdateLabelAndHostname(ctx, &state, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.UpdateTags(ctx, &state, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.UpdateLinks(ctx, &state, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
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

	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
		return
	}

	err = bp.DeleteGenericSystem(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != apstra.ErrNotfound {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting generic system", err.Error())
	}
}
