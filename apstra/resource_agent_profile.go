package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceAgentProfile{}

type resourceAgentProfile struct {
	client *apstra.Client
}

func (o *resourceAgentProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_profile"
}

func (o *resourceAgentProfile) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceAgentProfile) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an Agent Profile. Note that credentials (username/password) " +
			"cannot be set using this resource because (a) Apstra doesn't allow them to be retrieved, so it's " +
			"impossible for terraform to detect drift and because (b) leaving credentials in the configuration/state " +
			"isn't a safe practice.",
		Attributes: agentProfile{}.resourceAttributes(),
	}
}

func (o *resourceAgentProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan agentProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateAgentProfile(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Agent Profile", err.Error())
		return
	}

	plan.Id = types.StringValue(string(id))
	plan.HasUsername = types.BoolValue(false) // safe to assume false at creation time
	plan.HasPassword = types.BoolValue(false) // safe to assume false at creation time

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceAgentProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state agentProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Agent Profile from API and then update what is in state from what the API returns
	ap, err := o.client.GetAgentProfile(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error reading Agent Profile",
			fmt.Sprintf("Could not Read %q - %s", state.Id.ValueString(), err),
		)
		return
	}

	// Create new state object
	var newState agentProfile
	newState.loadApiData(ctx, ap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceAgentProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan agentProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update new Agent Profile
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var ace apstra.ApstraClientErr
	err := o.client.UpdateAgentProfile(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound { // deleted manually since 'plan'?
			resp.State.RemoveResource(ctx)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("error updating Agent Profile", err.Error())
		return
	}

	ap, err := o.client.GetAgentProfile(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error updating Agent Profile", err.Error())
		return
	}

	// Create new state object
	var newState agentProfile
	newState.loadApiData(ctx, ap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceAgentProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state agentProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Agent Profile by calling API
	err := o.client.DeleteAgentProfile(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != apstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Agent Profile", err.Error())
			return
		}
	}
}
