package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.ResourceWithConfigure = &resourceBlueprintDeploy{}
var _ resourceWithSetClient = &resourceBlueprintDeploy{}
var _ resourceWithSetBpLockFunc = &resourceBlueprintDeploy{}
var _ resourceWithSetBpUnlockFunc = &resourceBlueprintDeploy{}
var _ resourceWithSetProviderVersion = &resourceBlueprintDeploy{}
var _ resourceWithSetTerraformVersion = &resourceBlueprintDeploy{}

type resourceBlueprintDeploy struct {
	client           *apstra.Client
	lockFunc         func(context.Context, string) error
	unlockFunc       func(context.Context, string) error
	providerVersion  string
	terraformVersion string
}

func (o *resourceBlueprintDeploy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_deployment"
}

func (o *resourceBlueprintDeploy) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceBlueprintDeploy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This resource commits a staging Blueprint after checking for build errors.",
		Attributes:          blueprint.Deploy{}.ResourceAttributes(),
	}
}

func (o *resourceBlueprintDeploy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(plan.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.Diagnostics.AddError("Blueprint not found", fmt.Sprintf("Blueprint %q not found", plan.BlueprintId.ValueString()))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	template := blueprint.CommentTemplate{
		ProviderVersion:  o.providerVersion,
		TerraformVersion: o.terraformVersion,
	}

	plan.Deploy(ctx, &template, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintDeploy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.Deploy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new state object so we don't overwrite the comment template in Read().
	newState := blueprint.Deploy{BlueprintId: state.BlueprintId}
	newState.Read(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if newState.HasUncommittedChanges.ValueBool() {
		resp.State.RemoveResource(ctx)
		return
	}

	// Re-use the old comment template, rather than the rendered template we got during Read().
	newState.Comment = state.Comment

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceBlueprintDeploy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.Deploy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(plan.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.State.RemoveResource(ctx)
		return
	}
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

	template := blueprint.CommentTemplate{
		ProviderVersion:  o.providerVersion,
		TerraformVersion: o.terraformVersion,
	}

	plan.Deploy(ctx, &template, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintDeploy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.Deploy
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

	// Unlock the blueprint mutex.
	err := o.unlockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error unlocking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
	}
}

func (o *resourceBlueprintDeploy) setClient(client *apstra.Client) {
	o.client = client
}

func (o *resourceBlueprintDeploy) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceBlueprintDeploy) setBpUnlockFunc(f func(context.Context, string) error) {
	o.unlockFunc = f
}

func (o *resourceBlueprintDeploy) setProviderVersion(v string) {
	o.providerVersion = v
}
func (o *resourceBlueprintDeploy) setTerraformVersion(v string) {
	o.terraformVersion = v
}
