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

var _ resource.ResourceWithConfigure = &resourceDatacenterRoutingPolicy{}

type resourceDatacenterRoutingPolicy struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceDatacenterRoutingPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_policy"
}

func (o *resourceDatacenterRoutingPolicy) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDatacenterRoutingPolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Routing Policy within a Blueprint.",
		Attributes:          blueprint.DatacenterRoutingPolicy{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterRoutingPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterRoutingPolicy
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

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := bp.CreateRoutingPolicy(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating routing policy", err.Error())
		return
	}

	plan.Id = types.StringValue(id.String())
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterRoutingPolicy
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

	rp, err := bp.GetRoutingPolicy(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error retrieving routing policy", err.Error())
		return
	}

	state.LoadApiData(ctx, rp.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterRoutingPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterRoutingPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.UpdateRoutingPolicy(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating routing policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterRoutingPolicy
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

	err = bp.DeleteRoutingPolicy(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting routing policy", err.Error())
	}
}
