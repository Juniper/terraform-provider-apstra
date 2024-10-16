package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceTemplatePodBased{}
	_ resourceWithSetClient          = &resourceTemplatePodBased{}
)

type resourceTemplatePodBased struct {
	client *apstra.Client
}

func (o *resourceTemplatePodBased) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_pod_based"
}

func (o *resourceTemplatePodBased) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceTemplatePodBased) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a Pod Based Template for a 5-stage Clos design",
		Attributes:          design.TemplatePodBased{}.ResourceAttributes(),
	}
}

func (o *resourceTemplatePodBased) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan design.TemplatePodBased
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a CreatePodBasedTemplateRequest
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the PodBasedTemplate object (nested objects are referenced by ID)
	id, err := o.client.CreatePodBasedTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Pod-based template", err.Error())
		return
	}

	// retrieve the rack-based template object with fully-enumerated embedded objects
	api, err := o.client.GetPodBasedTemplate(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Pod-based template info after creation", err.Error())
		return
	}

	// parse the API response into a state object
	var state design.TemplatePodBased
	state.Id = types.StringValue(string(id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)

	// copy nested object IDs (those not available from the API) from the plan into the state
	state.CopyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceTemplatePodBased) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state design.TemplatePodBased
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Pod Based Template from API and then update what is in state from what the API returns
	api, err := o.client.GetPodBasedTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error reading Pod Based Template",
			fmt.Sprintf("Could not Read %q - %s", state.Id.ValueString(), err),
		)
		return
	}

	// Create new state object
	var newState design.TemplatePodBased
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
func (o *resourceTemplatePodBased) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan design.TemplatePodBased
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a CreatePodBasedTemplateRequest
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update
	err := o.client.UpdatePodBasedTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Pod Based Template",
			fmt.Sprintf("Could not update %q - %s", plan.Id.ValueString(), err),
		)
		return
	}

	api, err := o.client.GetPodBasedTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"error retrieving recently updated Pod Based Template",
			fmt.Sprintf("Could not fetch %q - %s", plan.Id.ValueString(), err),
		)
		return
	}

	// Create new state object
	var newState design.TemplatePodBased
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
func (o *resourceTemplatePodBased) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state design.TemplatePodBased
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Pod based Template by calling API
	err := o.client.DeleteTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			"error deleting Pod based Template",
			fmt.Sprintf("could not delete Pod based Template %q - %s", state.Id.ValueString(), err),
		)
		return
	}
}

func (o *resourceTemplatePodBased) setClient(client *apstra.Client) {
	o.client = client
}
