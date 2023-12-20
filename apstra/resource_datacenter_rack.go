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
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterRack) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_rack"
}

func (o *resourceDatacenterRack) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.getBpClientFunc = ResourceGetTwoStageL3ClosClientFunc(ctx, req, resp)
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

	// create the rack and squirrel away the rack ID
	id, err := bp.CreateRack(ctx, plan.Request())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Datacenter Rack", err.Error())
		return
	}
	plan.Id = types.StringValue(id.String())

	// fetch the rack name chosen by Apstra
	oldName, err := plan.GetName(ctx, bp.Client())
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch rack name", err.Error())
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Name is optional. Did the user supply one?
	if plan.Name.IsUnknown() {
		// no user supplied name.
		// save the apstra-generated name
		plan.Name = types.StringValue(oldName)
	} else {
		// user has provided a name.
		// set the rack name
		plan.SetName(ctx, bp.Client(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		// one-shot rename objects in the rack. "oldName" is used as the original
		// value in a substring replace operation.
		if plan.SystemNameOneShot.ValueBool() {
			plan.SetSystemNames(ctx, bp.Client(), oldName, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
				return
			}
		}
	}

	//set the state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRack) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.Rack
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

	// read the name (and confirm the rack still exists)
	name, err := state.GetName(ctx, bp.Client())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to fetch rack name", err.Error())
		return
	}

	// set state.
	state.Name = types.StringValue(name)
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

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// update the name if necessary
	if !plan.Name.Equal(state.Name) {
		plan.SetName(ctx, bp.Client(), &resp.Diagnostics)
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

	// Delete Rack by calling API
	err = bp.DeleteRack(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("unable to delete Rack %s from Blueprint %s", state.Id, state.BlueprintId),
			err.Error())
	}
}
