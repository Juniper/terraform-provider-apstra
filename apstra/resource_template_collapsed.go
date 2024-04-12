package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceTemplateCollapsed{}
var _ resourceWithSetClient = &resourceTemplateCollapsed{}

type resourceTemplateCollapsed struct {
	client *apstra.Client
}

func (o *resourceTemplateCollapsed) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_collapsed"
}

func (o *resourceTemplateCollapsed) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceTemplateCollapsed) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a Template for a spine-less (collapsed) Blueprint",
		Attributes:          design.TemplateCollapsed{}.ResourceAttributes(),
	}
}

func (o *resourceTemplateCollapsed) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// retrieve values from plan
	var plan design.TemplateCollapsed
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a CreatePodBasedTemplateRequest
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the API version
	apiVer, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed parsing API Version %q", o.client.ApiVersion()), err.Error())
		return
	}

	// Apstra <= 4.2.0 requires an anti-affinity policy in the request
	if version.MustConstraints(version.NewConstraint(apiversions.Le420)).Check(apiVer) {
		request.AntiAffinityPolicy = &apstra.AntiAffinityPolicy{
			Algorithm: apstra.AlgorithmHeuristic,
			Mode:      apstra.AntiAffinityModeDisabled,
		}
	}

	// create the CollapsedTemplate object (nested objects are referenced by ID)
	id, err := o.client.CreateL3CollapsedTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Collapsed template", err.Error())
		return
	}

	// save the ID to the state in case we run into a problem later
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	// retrieve the Collapsed template object with fully-enumerated embedded objects
	api, err := o.client.GetL3CollapsedTemplate(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Collapsed template info after creation", err.Error())
		return
	}

	// load API response and set state
	plan.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceTemplateCollapsed) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state design.TemplateCollapsed
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Pod Based Template from API and then update what is in state from what the API returns
	api, err := o.client.GetL3CollapsedTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Could not Read %q", state.Id.ValueString()),
			err.Error(),
		)
		return
	}

	// load API response and set state
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceTemplateCollapsed) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// retrieve values from plan
	var plan design.TemplateCollapsed
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a CreateCollapsedTemplateRequest
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the API version
	apiVer, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed parsing API Version %q", o.client.ApiVersion()), err.Error())
		return
	}

	// Apstra <= 4.2.0 requires an anti-affinity policy in the request
	if version.MustConstraints(version.NewConstraint(apiversions.Le420)).Check(apiVer) {
		request.AntiAffinityPolicy = &apstra.AntiAffinityPolicy{
			Algorithm: apstra.AlgorithmHeuristic,
			Mode:      apstra.AntiAffinityModeDisabled,
		}
	}

	// update
	err = o.client.UpdateL3CollapsedTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Collapsed Template",
			fmt.Sprintf("Could not update %q - %s", plan.Id.ValueString(), err),
		)
		return
	}

	api, err := o.client.GetL3CollapsedTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"error retrieving recently updated Collapsed Template",
			fmt.Sprintf("Could not fetch %q - %s", plan.Id.ValueString(), err),
		)
		return
	}

	// load API response and set state
	plan.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceTemplateCollapsed) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state design.TemplateCollapsed
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Collapsed Template by calling API
	err := o.client.DeleteTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			"error deleting Collapsed Template",
			fmt.Sprintf("could not delete Collapsed Template %q - %s", state.Id.ValueString(), err),
		)
		return
	}
}

func (o *resourceTemplateCollapsed) setClient(client *apstra.Client) {
	o.client = client
}
