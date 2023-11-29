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

var _ resource.ResourceWithConfigure = &resourceDatacenterRack{}

type resourceDatacenterRack struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceDatacenterRack) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_rack"
}

func (o *resourceDatacenterRack) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDatacenterRack) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a new Rack in a Datacenter Blueprint.",
		Attributes:          blueprint.Rack{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterRack) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan blueprint.Rack
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a blueprint client
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found",
				plan.BlueprintId), err.Error())
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

	// create the rack and squirrel away the rack ID
	id, err := bpClient.CreateRack(ctx, plan.Request())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Datacenter Rack", err.Error())
		return
	}
	plan.Id = types.StringValue(id.String())

	// get or set the rack name
	if plan.Name.IsUnknown() {
		err = plan.GetName(ctx, o.client)
		if err != nil {
			resp.Diagnostics.AddError("failed to fetch rack name", err.Error())
			// do not return
		}
	} else {
		plan.SetName(ctx, o.client, &resp.Diagnostics)
		// do not check resp.Diagnostics.HasError()
	}

	// update the plan with the rack ID and set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRack) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.Rack
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// read the name (and confirm the rack still exists)
	err := state.GetName(ctx, o.client)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to fetch rack name", err.Error())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterRack) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.Rack
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state blueprint.Rack
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the name if necessary
	if !plan.Name.Equal(state.Name) {
		plan.SetName(ctx, o.client, &resp.Diagnostics)
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRack) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.Rack
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a blueprint client
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("unable to get blueprint client", err.Error())
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

	// Delete Rack by calling API
	err = bpClient.DeleteRack(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("unable to delete Rack %s from Blueprint %s", state.Id, state.BlueprintId),
			err.Error())
	}
}
