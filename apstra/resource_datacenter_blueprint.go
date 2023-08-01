package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterBlueprint{}
var _ resource.ResourceWithValidateConfig = &resourceDatacenterBlueprint{}
var _ versionValidator = &resourceDatacenterBlueprint{}

type resourceDatacenterBlueprint struct {
	client           *apstra.Client
	minClientVersion *version.Version
	maxClientVersion *version.Version
	lockFunc         func(context.Context, string) error
	unlockFunc       func(context.Context, string) error
}

func (o *resourceDatacenterBlueprint) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *resourceDatacenterBlueprint) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
	o.unlockFunc = ResourceGetBlueprintUnlockFunc(ctx, req, resp)
}

func (o *resourceDatacenterBlueprint) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource instantiates a Datacenter Blueprint from a template.",
		Attributes:          blueprint.Blueprint{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterBlueprint) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Cannot proceed without a client
	if o.client == nil {
		return
	}

	var config blueprint.Blueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the min/max API versions required by the client. These elements set within 'o'
	// do not persist after ValidateConfig exits even though 'o' is a pointer receiver.
	o.minClientVersion, o.maxClientVersion = config.MinMaxApiVersions(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if o.client == nil {
		// Bail here because we can't validate config's API version needs if the client doesn't exist.
		// This method should be called again (after the provider's Configure() method) with a non-nil
		// client pointer.
		return
	}

	// validate version compatibility between the API server and the configuration's min/max needs.
	o.checkVersion(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
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

	// Create blueprint client
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("failed creating blueprint client", err.Error())
	}

	// set the fabric addressing policy
	plan.SetFabricAddressingPolicy(ctx, bpClient, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve blueprint status
	apiData, err := o.client.GetBlueprintStatus(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint after creation", err.Error())
	}

	fapData, err := bpClient.GetFabricAddressingPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint Fabric Addressing Policy after creation", err.Error())
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

	// Some interesting details are in BlueprintStatus.
	apiData, err := o.client.GetBlueprintStatus(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("fetching blueprint %q", state.Id.ValueString()),
			err.Error(),
		)
		return
	}

	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apiData.Id)
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	fapData, err := bpClient.GetFabricAddressingPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to read fabric addressing policy", err.Error())
		return
	}

	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	state.LoadFabricAddressingPolicy(ctx, fapData, &resp.Diagnostics)

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

	// create a blueprint client
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("failed creating blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed locking blueprint mutex", err.Error())
		return
	}

	// set the blueprint name
	plan.SetName(ctx, bpClient, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set the fabric addressing policy
	plan.SetFabricAddressingPolicy(ctx, bpClient, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch and load blueprint info
	apiData, err := o.client.GetBlueprintStatus(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("failed retrieving Datacenter Blueprint after update", err.Error())
	}
	plan.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	fapData, err := bpClient.GetFabricAddressingPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed retrieving Datacenter Blueprint Fabric AddressingPolicy after update", err.Error())
	}
	plan.LoadFabricAddressingPolicy(ctx, fapData, &resp.Diagnostics)
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

	bpID := apstra.ObjectId(state.Id.ValueString())

	err := o.client.DeleteBlueprint(ctx, bpID)
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Blueprint", err.Error())
	}

	// Unlock the blueprint mutex.
	err = o.unlockFunc(ctx, bpID.String())
	if err != nil {
		resp.Diagnostics.AddError("error unlocking blueprint mutex", err.Error())
	}
}

func (o *resourceDatacenterBlueprint) apiVersion() (*version.Version, error) {
	if o.client == nil {
		return nil, nil
	}
	return version.NewVersion(o.client.ApiVersion())
}

func (o *resourceDatacenterBlueprint) cfgVersionMin() (*version.Version, error) {
	return o.minClientVersion, nil
}

func (o *resourceDatacenterBlueprint) cfgVersionMax() (*version.Version, error) {
	return o.maxClientVersion, nil
}

func (o *resourceDatacenterBlueprint) checkVersion(ctx context.Context, diags *diag.Diagnostics) {
	checkVersionCompatibility(ctx, o, diags)
}
