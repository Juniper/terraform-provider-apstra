package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterBlueprint{}
var _ resource.ResourceWithValidateConfig = &resourceDatacenterBlueprint{}
var _ versionValidator = &resourceDatacenterBlueprint{}

type resourceDatacenterBlueprint struct {
	client           *goapstra.Client
	minClientVersion *version.Version
	maxClientVersion *version.Version
}

func (o *resourceDatacenterBlueprint) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *resourceDatacenterBlueprint) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
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
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateBlueprintFromTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Rack Based Blueprint", err.Error())
	}

	apiData, err := o.client.GetBlueprintStatus(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint after creation", err.Error())
	}

	// Create new state object
	var state blueprint.Blueprint
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.FabricAddressing = plan.FabricAddressing // blindly copy because resource.RequiresReplace()
	state.TemplateId = plan.TemplateId             // blindly copy because resource.RequiresReplace()

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterBlueprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state blueprint.Blueprint
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// some interesting details are in blueprintStatus
	apiData, err := o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error fetching Blueprint",
			fmt.Sprintf("Could not read %q - %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	// create new state object with some obvious values
	var newState blueprint.Blueprint
	newState.LoadApiData(ctx, apiData, &resp.Diagnostics)
	newState.FabricAddressing = state.FabricAddressing // blindly copy because resource.RequiresReplace()
	newState.TemplateId = state.TemplateId             // blindly copy because resource.RequiresReplace()

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceDatacenterBlueprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve plan
	var plan blueprint.Blueprint
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// name change is the only possible update method (other attributes trigger replace)
	plan.SetName(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Datacenter Blueprint after creation", err.Error())
	}

	// Create new state object
	var state blueprint.Blueprint
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.FabricAddressing = plan.FabricAddressing // blindly copy because resource.RequiresReplace()
	state.TemplateId = plan.TemplateId             // blindly copy because resource.RequiresReplace()

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete resource
func (o *resourceDatacenterBlueprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state blueprint.Blueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteBlueprint(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Blueprint", err.Error())
		return
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
