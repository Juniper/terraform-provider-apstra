package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/apstra/freeform"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceFreeformBlueprint{}
	_ resourceWithSetClient          = &resourceFreeformBlueprint{}
	_ resourceWithSetFfBpClientFunc  = &resourceFreeformBlueprint{}
	_ resourceWithSetBpLockFunc      = &resourceFreeformBlueprint{}
	_ resourceWithSetBpUnlockFunc    = &resourceFreeformBlueprint{}
)

type resourceFreeformBlueprint struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
	lockFunc        func(context.Context, string) error
	unlockFunc      func(context.Context, string) error
}

func (o *resourceFreeformBlueprint) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_blueprint"
}

func (o *resourceFreeformBlueprint) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceFreeformBlueprint) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This resource instantiates a Freeform Blueprint.",
		Attributes:          freeform.Blueprint{}.ResourceAttributes(),
	}
}

func (o *resourceFreeformBlueprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan freeform.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the blueprint.
	id, err := o.client.CreateFreeformBlueprint(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed creating Blueprint", err.Error())
		return
	}

	plan.Id = types.StringValue(id.String())

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformBlueprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state freeform.Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the freeform reference design
	bp, err := o.getBpClientFunc(ctx, state.Id.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Retrieve the blueprint status
	apiData, err := bp.Client().GetBlueprintStatus(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		// no 404 check or RemoveResource() here because Apstra's /api/blueprints
		// endpoint may return bogus 404s due to race condition (?)
		resp.Diagnostics.AddError(fmt.Sprintf("failed fetching blueprint %s status", state.Id), err.Error())
		return
	}

	state.Name = types.StringValue(apiData.Label)

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceFreeformBlueprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve plan.
	var plan freeform.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve state.
	var state freeform.Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the freeform reference design
	bp, err := o.getBpClientFunc(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Update the blueprint name if necessary
	plan.SetName(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceFreeformBlueprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state freeform.Blueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the blueprint
	err := o.client.DeleteBlueprint(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if !utils.IsApstra404(err) { // 404 is okay, but we do not return because we must unlock
			resp.Diagnostics.AddError("error deleting Blueprint", err.Error())
		}
	}

	// Unlock the blueprint mutex.
	err = o.unlockFunc(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("error unlocking blueprint mutex", err.Error())
	}
}

func (o *resourceFreeformBlueprint) setClient(client *apstra.Client) {
	o.client = client
}

func (o *resourceFreeformBlueprint) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceFreeformBlueprint) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceFreeformBlueprint) setBpUnlockFunc(f func(context.Context, string) error) {
	o.unlockFunc = f
}
