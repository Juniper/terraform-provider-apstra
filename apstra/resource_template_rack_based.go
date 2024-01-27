package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceTemplateRackBased{}
var _ resource.ResourceWithValidateConfig = &resourceTemplateRackBased{}
var _ versionValidator = &resourceTemplateRackBased{}

type resourceTemplateRackBased struct {
	client           *apstra.Client
	minClientVersion *version.Version
	maxClientVersion *version.Version
}

func (o *resourceTemplateRackBased) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_rack_based"
}

func (o *resourceTemplateRackBased) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceTemplateRackBased) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a Rack Based Template for as a 3-stage Clos design, or for use as " +
			"pod in a 5-stage design.",
		Attributes: design.TemplateRackBased{}.ResourceAttributes(),
	}
}

func (o *resourceTemplateRackBased) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config design.TemplateRackBased
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validate the configuration
	config.Validate(ctx, &resp.Diagnostics)
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

func (o *resourceTemplateRackBased) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan design.TemplateRackBased
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check compatibility of config against API version
	plan.CheckCompatibility(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a CreateRackBasedTemplateRequest
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the RackBasedTemplate object (nested objects are referenced by ID)
	id, err := o.client.CreateRackBasedTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating rack-based template", err.Error())
		return
	}

	// retrieve the rack-based template object with fully-enumerated embedded objects
	api, err := o.client.GetRackBasedTemplate(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving rack-based template info after creation", err.Error())
		return
	}

	// parse the API response into a state object
	state := design.TemplateRackBased{}
	state.Id = types.StringValue(string(id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)

	// copy nested object IDs (those not available from the API) from the plan into the state
	state.CopyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceTemplateRackBased) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state design.TemplateRackBased
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Rack Based Template from API and then update what is in state from what the API returns
	api, err := o.client.GetRackBasedTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error reading Rack Based Template",
			fmt.Sprintf("Could not Read %q - %s", state.Id.ValueString(), err),
		)
		return
	}

	// Create new state object
	var newState design.TemplateRackBased
	newState.Id = types.StringValue(string(api.Id))
	newState.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the state into the newState
	newState.CopyWriteOnlyElements(ctx, &state, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceTemplateRackBased) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan design.TemplateRackBased
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check compatibility of config against API version
	plan.CheckCompatibility(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a CreateRackBasedTemplateRequest
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update
	err := o.client.UpdateRackBasedTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Rack Based Template",
			fmt.Sprintf("Could not update %q - %s", plan.Id.ValueString(), err),
		)
		return
	}

	api, err := o.client.GetRackBasedTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"error retrieving recently updated Rack Based Template",
			fmt.Sprintf("Could not fetch %q - %s", plan.Id.ValueString(), err),
		)
		return
	}

	// Create new state object
	var newState design.TemplateRackBased
	newState.Id = types.StringValue(string(api.Id))
	newState.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the plan into the newState
	newState.CopyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete resource
func (o *resourceTemplateRackBased) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state design.TemplateRackBased
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Agent Profile by calling API
	err := o.client.DeleteTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			"error deleting Agent Profile",
			fmt.Sprintf("could not delete Agent Profile %q - %s", state.Id.ValueString(), err),
		)
		return
	}
}

func (o *resourceTemplateRackBased) apiVersion() (*version.Version, error) {
	if o.client == nil {
		return nil, nil
	}
	return version.NewVersion(o.client.ApiVersion())
}

func (o *resourceTemplateRackBased) cfgVersionMin() (*version.Version, error) {
	return o.minClientVersion, nil
}

func (o *resourceTemplateRackBased) cfgVersionMax() (*version.Version, error) {
	return o.maxClientVersion, nil
}

func (o *resourceTemplateRackBased) checkVersion(ctx context.Context, diags *diag.Diagnostics) {
	checkVersionCompatibility(ctx, o, diags)
}
