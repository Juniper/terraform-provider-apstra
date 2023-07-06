package tfapstra

import (
	"context"
	"errors"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/design"
)

var _ resource.ResourceWithConfigure = &resourceConfiglet{}

type resourceConfiglet struct {
	client *apstra.Client
}

func (o *resourceConfiglet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlet"
}

func (o *resourceConfiglet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceConfiglet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a specific Configlet.\n\n",
		Attributes:          design.Configlet{}.ResourceAttributes(),
	}
}

func (o *resourceConfiglet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan design.Configlet
	req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateConfiglet(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Configlet", err.Error())
		return
	}

	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceConfiglet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	var state design.Configlet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := o.client.GetConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read Configlet", err.Error())
		return
	}

	state.Id = types.StringValue(string(api.Id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceConfiglet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan design.Configlet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create request
	c := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Configlet
	err := o.client.UpdateConfiglet(ctx, apstra.ObjectId(plan.Id.ValueString()), c)
	if err != nil {
		resp.Diagnostics.AddError("error updating Configlet", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceConfiglet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state design.Configlet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Configlet by calling API
	err := o.client.DeleteConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != apstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Configlet", err.Error())
			return
		}
	}
}
