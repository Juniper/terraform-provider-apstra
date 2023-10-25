package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceRackType{}

type resourceRackType struct {
	client *apstra.Client
}

func (o *resourceRackType) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_type"
}

func (o *resourceRackType) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceRackType) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a Rack Type in the Apstra Design tab.",
		Attributes:          design.RackType{}.ResourceAttributes(),
	}
}

func (o *resourceRackType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan design.RackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a RackTypeRequest
	rtRequest := plan.Request(ctx, &resp.Diagnostics)
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
	design.ValidateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	state := design.RackType{}
	state.Id = types.StringValue(string(rt.Id))
	state.LoadApiData(ctx, rt.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the plan into the state
	state.CopyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceRackType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state design.RackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch the rack type detail from the API
	rt, err := o.client.GetRackType(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading rack type", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	design.ValidateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a new state object
	var newState design.RackType
	newState.LoadApiData(ctx, rt.Data, &resp.Diagnostics)
	newState.Id = types.StringValue(string(rt.Id))
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the previous state into the new state
	newState.CopyWriteOnlyElements(ctx, &state, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve plan
	var plan design.RackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a RackTypeRequest
	rtRequest := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// send the Request to Apstra
	err := o.client.UpdateRackType(ctx, apstra.ObjectId(plan.Id.ValueString()), rtRequest)
	if err != nil {
		resp.Diagnostics.AddError("error while updating Rack Type", err.Error())
		return
	}

	// retrieve the RackType object with fully-enumerated embedded objects
	rt, err := o.client.GetRackType(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error retrieving rack type info after creation", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	design.ValidateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	var state design.RackType
	state.LoadApiData(ctx, rt.Data, &resp.Diagnostics)
	state.Id = types.StringValue(string(rt.Id))
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the (old) into state
	state.CopyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceRackType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state design.RackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteRackType(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay in Delete()
		}
		resp.Diagnostics.AddError("error deleting Rack Type", err.Error())
		return
	}
}
