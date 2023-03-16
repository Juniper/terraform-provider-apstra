package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourceBlueprintDeploy{}

type resourceBlueprintDeploy struct {
	client          *goapstra.Client
	commentTemplate *blueprint.CommentTemplate
	lockFunc        func(context.Context, *goapstra.TwoStageL3ClosMutex) error
	unlockFunc      func(context.Context, goapstra.ObjectId) error
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

	bp, err := o.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error creating blueprint client", err.Error())
		return
	}

	err = o.lockFunc(ctx, bp.Mutex)
	if err != nil {
		resp.Diagnostics.AddError("error locking blueprint mutex", err.Error())
		return
	}

	plan.Deploy(ctx, o.commentTemplate, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = o.unlockFunc(ctx, goapstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error unlocking blueprint %q", plan.BlueprintId.ValueString()), err.Error())
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

	plan.Deploy(ctx, o.commentTemplate, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintDeploy) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// nothing to do
}
