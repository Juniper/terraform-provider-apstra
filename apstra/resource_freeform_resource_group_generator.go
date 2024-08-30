package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/freeform"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceFreeformGroupGenerator{}
	_ resourceWithSetFfBpClientFunc  = &resourceFreeformGroupGenerator{}
	_ resourceWithSetBpLockFunc      = &resourceFreeformGroupGenerator{}
)

type resourceFreeformGroupGenerator struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceFreeformGroupGenerator) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_resource_group_generator"
}

func (o *resourceFreeformGroupGenerator) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceFreeformGroupGenerator) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This resource creates a Group Generator in a Freeform Blueprint.",
		Attributes:          freeform.GroupGenerator{}.ResourceAttributes(),
	}
}

func (o *resourceFreeformGroupGenerator) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan freeform.GroupGenerator
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
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

	// Convert the plan into an API Request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the resource
	id, err := bp.CreateGroupGenerator(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new Group Generator", err.Error())
		return
	}

	// set state
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformGroupGenerator) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state freeform.GroupGenerator
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	api, err := bp.GetGroupGenerator(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Freeform Group Generator", err.Error())
		return
	}

	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceFreeformGroupGenerator) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan freeform.GroupGenerator
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
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

	// Convert the plan into an API Request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the Group
	err = bp.UpdateGroupGenerator(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Freeform Group Generator", err.Error())
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformGroupGenerator) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state freeform.GroupGenerator
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
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

	// Delete Resource Generator by calling API
	err = bp.DeleteGroupGenerator(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Freeform Group", err.Error())
		return
	}
}

func (o *resourceFreeformGroupGenerator) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceFreeformGroupGenerator) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
