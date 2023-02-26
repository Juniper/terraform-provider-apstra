package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceAgentProfile{}

type resourceTemplateRackBased struct {
	client *goapstra.Client
}

func (o *resourceTemplateRackBased) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_rack_based"
}

func (o *resourceTemplateRackBased) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = resourceGetClient(ctx, req, resp)
}

func (o *resourceTemplateRackBased) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Rack Based Template for as a 3-stage Clos design, or for use as " +
			"pod in a 5-stage design.",
		Attributes: templateRackBased{}.resourceAttributes(),
	}
}

func (o *resourceTemplateRackBased) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan agentProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Agent Profile
	id, err := o.client.CreateAgentProfile(ctx, plan.request(ctx, &resp.Diagnostics))
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent Profile",
			"Could not create, unexpected error: "+err.Error(),
		)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// create state object
	state := agentProfile{
		Id:          types.StringValue(string(id)),
		Name:        plan.Name,
		Platform:    plan.Platform,
		HasUsername: types.BoolValue(false),
		HasPassword: types.BoolValue(false),
		Packages:    plan.Packages,
		OpenOptions: plan.OpenOptions,
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceTemplateRackBased) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state agentProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Agent Profile from API and then update what is in state from what the API returns
	ap, err := o.client.GetAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading Agent Profile",
				fmt.Sprintf("Could not Read '%s' - %s", state.Id.ValueString(), err),
			)
			return
		}
	}

	// Create new state object
	var newState agentProfile
	newState.loadApiResponse(ctx, ap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceTemplateRackBased) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get current state
	var state agentProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan agentProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update new Agent Profile
	err := o.client.UpdateAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()), plan.request(ctx, &resp.Diagnostics))
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Agent Profile",
			fmt.Sprintf("Could not Update '%s' - %s", state.Id.ValueString(), err),
		)
		return
	}

	ap, err := o.client.GetAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Agent Profile",
			fmt.Sprintf("Could not Update '%s' - %s", state.Id.ValueString(), err),
		)
		return
	}

	// Create new state object
	var newState agentProfile
	newState.loadApiResponse(ctx, ap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete resource
func (o *resourceTemplateRackBased) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state agentProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Agent Profile by calling API
	err := o.client.DeleteAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError(
				"error deleting Agent Profile",
				fmt.Sprintf("could not delete Agent Profile '%s' - %s", state.Id.ValueString(), err),
			)
			return
		}
	}
}
