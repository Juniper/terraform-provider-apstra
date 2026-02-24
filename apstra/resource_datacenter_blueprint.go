package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceDatacenterBlueprint{}
	_ resource.ResourceWithValidateConfig = &resourceDatacenterBlueprint{}
	_ resourceWithSetClient               = &resourceDatacenterBlueprint{}
	_ resourceWithSetDcBpClientFunc       = &resourceDatacenterBlueprint{}
	_ resourceWithSetBpLockFunc           = &resourceDatacenterBlueprint{}
	_ resourceWithSetBpUnlockFunc         = &resourceDatacenterBlueprint{}
)

type resourceDatacenterBlueprint struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
	unlockFunc      func(context.Context, string) error
}

func (o *resourceDatacenterBlueprint) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *resourceDatacenterBlueprint) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterBlueprint) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource instantiates a Datacenter Blueprint from a template.",
		Attributes:          blueprint.Blueprint{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterBlueprint) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config.
	var config blueprint.Blueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// config-only validation begins here

	// config + api version validation begins here

	// cannot proceed to config + api version validation if the provider has not been configured
	if o.client == nil {
		return
	}

	apiVersion, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse API version %q", o.client.ApiVersion()), err.Error())
		return
	}

	// validate the configuration
	constraints := config.VersionConstraints(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		compatibility.ValidateConfigConstraints(
			ctx,
			compatibility.ValidateConfigConstraintsRequest{
				Version:     apiVersion,
				Constraints: constraints,
			},
		)...,
	)
}

func (o *resourceDatacenterBlueprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// make a blueprint creation request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the blueprint.
	id, err := o.client.CreateBlueprintFromTemplate(ctx, &request)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed creating Blueprint from Template %s", request.TemplateId), err.Error())
		return
	}

	// Commit the ID to the state in case we're not able to run to completion
	plan.Id = types.StringValue(id.String())
	// plan.AntiAffinityPolicy = types.ObjectNull(blueprint.AntiAffinityPolicy{}.AttrTypes())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, id.String())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// get the api version from the client
	apiVersion, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse API version %q", o.client.ApiVersion()), err.Error())
		return
	}

	// Apstra 4.2.1 allows us to set *some* fabric settings as part of blueprint creation.
	// Depending on the version and what's in the plan, we might not need to invoke SetFabricSettings().
	if !compatibility.FabricSettingsSetInCreate.Check(apiVersion) || plan.Ipv6Applications.ValueBool() {
		// Set the fabric settings
		plan.SetFabricSettings(ctx, bp, nil, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Retrieve blueprint status
	apiData, err := o.client.GetBlueprintStatus(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint after creation", err.Error())
		return
	}

	// Load blueprint status
	plan.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve and load the fabric settings
	plan.GetFabricSettings(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve any un-set RZ parameters
	if plan.UnderlayAddressing.IsUnknown() || plan.VTEPAddressing.IsUnknown() || plan.DisableIPv4.IsUnknown() {
		plan.GetDefaultRZParams(ctx, bp, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterBlueprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state blueprint.Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
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

	// Load the blueprint status
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve and load the fabric settings
	state.GetFabricSettings(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.GetDefaultRZParams(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceDatacenterBlueprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve plan.
	var plan blueprint.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve state.
	var state blueprint.Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed locking blueprint mutex", err.Error())
		return
	}

	// Update the blueprint name if necessary
	plan.SetName(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the blueprint settings if necessary
	plan.SetFabricSettings(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve blueprint status
	apiData, err := bp.Client().GetBlueprintStatus(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("failed retrieving Datacenter Blueprint after update", err.Error())
		return
	}

	// Load the blueprint status
	plan.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve and load the fabric settings
	plan.GetFabricSettings(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update default RZ parameters if necessary
	plan.SetDefaultRZParams(ctx, state, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve any un-set RZ parameters
	if plan.UnderlayAddressing.IsUnknown() || plan.VTEPAddressing.IsUnknown() || plan.DisableIPv4.IsUnknown() {
		plan.GetDefaultRZParams(ctx, bp, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceDatacenterBlueprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blueprint.Blueprint
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

func (o *resourceDatacenterBlueprint) setClient(client *apstra.Client) {
	o.client = client
}

func (o *resourceDatacenterBlueprint) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterBlueprint) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceDatacenterBlueprint) setBpUnlockFunc(f func(context.Context, string) error) {
	o.unlockFunc = f
}
