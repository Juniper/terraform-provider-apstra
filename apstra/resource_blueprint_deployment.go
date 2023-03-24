package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceBlueprintDeploy{}

type resourceBlueprintDeploy struct {
	client          *goapstra.Client
	commentTemplate *blueprint.CommentTemplate
	lockFunc        func(context.Context, string) error
	unlockFunc      func(context.Context, string) error
}

func (o *resourceBlueprintDeploy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_deployment"
}

func (o *resourceBlueprintDeploy) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.commentTemplate = &blueprint.CommentTemplate{
		ProviderVersion:  ResourceGetProviderVersion(ctx, req, resp),
		TerraformVersion: ResourceGetTerraformVersion(ctx, req, resp),
	}
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
	o.unlockFunc = ResourceGetBlueprintUnlockFunc(ctx, req, resp)
}

func (o *resourceBlueprintDeploy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource commits a staging Blueprint after checking for build errors.",
		Attributes:          blueprint.Deploy{}.ResourceAttributes(),
	}
}

func (o *resourceBlueprintDeploy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.Deploy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	defer func() {
		err := o.unlockFunc(ctx, plan.BlueprintId.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error unlocking blueprint %q mutex", plan.BlueprintId.ValueString()),
				err.Error())
		}
	}()

	if !utils.BlueprintExists(ctx, o.client, goapstra.ObjectId(plan.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.Diagnostics.AddError("Blueprint not found", fmt.Sprintf("Blueprint %q not found", plan.BlueprintId.ValueString()))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Deploy(ctx, o.commentTemplate, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintDeploy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.Deploy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.BlueprintExists(ctx, o.client, goapstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Read(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.HasUncommittedChanges.ValueBool() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceBlueprintDeploy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.Deploy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.BlueprintExists(ctx, o.client, goapstra.ObjectId(plan.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	defer func() {
		err := o.unlockFunc(ctx, plan.BlueprintId.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error unlocking blueprint %q mutex", plan.BlueprintId.ValueString()),
				err.Error())
		}
	}()

	plan.Deploy(ctx, o.commentTemplate, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintDeploy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.Deploy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No need to proceed if the blueprint no longer exists
	if !utils.BlueprintExists(ctx, o.client, goapstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		return
	}

	// Unlock the blueprint mutex.
	err := o.unlockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error unlocking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
	}
}
