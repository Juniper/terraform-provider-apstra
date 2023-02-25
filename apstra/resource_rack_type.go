package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	vlanMin = 1
	vlanMax = 4094

	poIdMin = 0
	poIdMax = 4096
)

var _ resource.ResourceWithConfigure = &resourceRackType{}

type resourceRackType struct {
	client *goapstra.Client
}

func (o *resourceRackType) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_type"
}

func (o *resourceRackType) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = resourceGetClient(ctx, req, resp)
}

func (o *resourceRackType) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Rack Type in the Apstra Design tab.",
		Attributes:          rackType{}.resourceAttributes(),
	}
}

func (o *resourceRackType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a RackTypeRequest
	rtRequest := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the RackType object (nested objects are referenced by ID)
	id, err := o.client.CreateRackType(ctx, rtRequest)
	if err != nil {
		resp.Diagnostics.AddError("error creating rack type", err.Error())
		return
	}

	// retrieve the RackType object with fully-enumerated embedded objects
	rt, err := o.client.GetRackType(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving rack type info after creation", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	validateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	state := rackType{}
	state.Id = types.StringValue(string(rt.Id))
	state.loadApiData(ctx, rt.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the plan into the state
	state.copyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceRackType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state
	var state rackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch the rack type detail from the API
	rt, err := o.client.GetRackType(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading rack type", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	validateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a new state object
	var newState rackType
	newState.Id = types.StringValue(string(rt.Id))
	newState.loadApiData(ctx, rt.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the previous state into the new state
	newState.copyWriteOnlyElements(ctx, &state, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve plan
	var plan rackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a RackTypeRequest
	rtRequest := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// send the request to Apstra
	err := o.client.UpdateRackType(ctx, goapstra.ObjectId(plan.Id.ValueString()), rtRequest)
	if err != nil {
		resp.Diagnostics.AddError("error while updating Rack Type", err.Error())
		return
	}

	// retrieve the RackType object with fully-enumerated embedded objects
	rt, err := o.client.GetRackType(ctx, goapstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error retrieving rack type info after creation", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	validateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	state := &rackType{}
	state.Id = types.StringValue(string(rt.Id))
	state.loadApiData(ctx, rt.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the (old) into state
	state.copyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceRackType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	// Retrieve values from state
	var state rackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteRackType(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			return // 404 is okay in Delete()
		}
		resp.Diagnostics.AddError("error deleting Rack Type", err.Error())
	}
}
