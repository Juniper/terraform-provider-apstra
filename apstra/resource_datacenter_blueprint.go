package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterBlueprint{}
var _ resource.ResourceWithValidateConfig = &resourceDatacenterBlueprint{}
var _ resourceWithSetClient = &resourceDatacenterBlueprint{}
var _ resourceWithSetBpClientFunc = &resourceDatacenterBlueprint{}
var _ resourceWithSetBpLockFunc = &resourceDatacenterBlueprint{}
var _ resourceWithSetBpUnlockFunc = &resourceDatacenterBlueprint{}

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
	var config blueprint.Blueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// config-only validation begins here (there is none)

	// cannot proceed to config + api version validation if the provider has not been configured
	if o.client == nil {
		return
	}

	// config + api version validation begins here

	// get the api version from the client
	apiVersion, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse API version %q", o.client.ApiVersion()), err.Error())
		return
	}

	// validate the configuration
	resp.Diagnostics.Append(
		apiversions.ValidateConstraints(
			ctx,
			apiversions.ValidateConstraintsRequest{
				Version:     apiVersion,
				Constraints: config.VersionConstraints(),
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

	// Make a blueprint creation request.
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the blueprint.
	id, err := o.client.CreateBlueprintFromTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("failed creating Rack Based Blueprint", err.Error())
		return
	}

	// commit the ID to the state
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, id.String())
	if err != nil {
		resp.Diagnostics.AddError("failed locking blueprint mutex", err.Error())
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, id.String())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// set the fabric addressing policy, passing no prior state
	plan.SetFabricAddressingPolicy(ctx, bp, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// todo do i need a fabricSettings plan.SetFabricSettings() here?
	// retrieve blueprint status
	apiData, err := o.client.GetBlueprintStatus(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint after creation", err.Error())
	}
	fapData, err := bp.GetFabricAddressingPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint Fabric Addressing Policy after creation", err.Error())
		return
	}
	fabSettings, err := bp.GetFabricSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint Fabric Settings after creation", err.Error())
		return
	}

	// load blueprint status
	plan.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.LoadFabricAddressingPolicy(ctx, fapData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.LoadFabricSettings(ctx, fabSettings, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.GetFabricLinkAddressing(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

	// Some interesting details are in BlueprintStatus.
	apiData, err := bp.Client().GetBlueprintStatus(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("fetching blueprint %q", state.Id.ValueString()),
			err.Error(),
		)
		return
	}

	fapData, err := bp.GetFabricAddressingPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to read fabric addressing policy", err.Error())
		return
	}

	fabricSettings, err := bp.GetFabricSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to read fabric settings", err.Error())
	}

	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	state.LoadFabricAddressingPolicy(ctx, fapData, &resp.Diagnostics)
	state.LoadFabricSettings(ctx, fabricSettings, &resp.Diagnostics)
	state.GetFabricLinkAddressing(ctx, bp, &resp.Diagnostics)

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

	// set the blueprint name
	plan.SetName(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set the fabric addressing policy
	plan.SetFabricAddressingPolicy(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// todo do i need a plan.SetFabricSettings() here?  review with ChrisM
	plan.SetFabricSettings(ctx, bp, &resp.Diagnostics) // todo no &state here like the SetFabricAddressingPolicy, is it needed?
	if resp.Diagnostics.HasError() {
		return
	}
	// fetch and load blueprint info
	apiData, err := bp.Client().GetBlueprintStatus(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("failed retrieving Datacenter Blueprint after update", err.Error())
	}
	plan.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	fapData, err := bp.GetFabricAddressingPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed retrieving Datacenter Blueprint Fabric AddressingPolicy after update", err.Error())
	}
	plan.LoadFabricAddressingPolicy(ctx, fapData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//todo review this with Chris
	fabSettings, err := bp.GetFabricSettings(ctx)
	plan.LoadFabricSettings(ctx, fabSettings, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("failed retrieving Datacenter Blueprint Fabric Settings after update", err.Error())
		return
	}

	plan.GetFabricLinkAddressing(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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
