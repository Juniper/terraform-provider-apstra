package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterPropertySet{}

type resourceDatacenterPropertySet struct {
	client *apstra.Client
}

func (o *resourceDatacenterPropertySet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_property_set"
}

func (o *resourceDatacenterPropertySet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceDatacenterPropertySet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource imports a property set into a Blueprint.",
		Attributes:          blueprint.DatacenterPropertySet{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterPropertySet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}
	// Retrieve values from plan
	var plan blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	// Convert the plan into an API Request
	var keys []string
	resp.Diagnostics.Append(plan.Keys.ElementsAs(ctx, &keys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ps_id, err := bpClient.ImportPropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()), keys...)
	if err != nil {
		resp.Diagnostics.AddError("Error importing DatacenterPropertySet", err.Error())
		return
	}

	// Read it back
	var api *apstra.TwoStageL3ClosPropertySet
	var ace apstra.ApstraClientErr

	api, err = bpClient.GetPropertySet(ctx, ps_id)
	if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"DatacenterPropertySet not found",
			fmt.Sprintf("DatacenterPropertySet with ID %q not found", plan.Id.ValueString()))
		return
	}
	// create new state object
	var state blueprint.DatacenterPropertySet
	state.Id = types.StringValue(string(api.Id))
	state.BlueprintId = plan.BlueprintId
	state.Keys = plan.Keys
	state.LoadApiData(ctx, api, false, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterPropertySet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}
	var plan blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))

	var api *apstra.TwoStageL3ClosPropertySet
	var ace apstra.ApstraClientErr

	switch {
	case !plan.Label.IsNull():
		api, err = bpClient.GetPropertySetByName(ctx, plan.Label.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"DatacenterPropertySet not found",
				fmt.Sprintf("DatacenterPropertySet with label %q not found", plan.Label.ValueString()))
			return
		}
	case !plan.Id.IsNull():
		api, err = bpClient.GetPropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"DatacenterPropertySet not found",
				fmt.Sprintf("DatacenterPropertySet with ID %q not found", plan.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving DatacenterPropertySet", err.Error())
		return
	}

	// create new state object
	var state blueprint.DatacenterPropertySet
	state.Id = types.StringValue(string(api.Id))
	state.BlueprintId = plan.BlueprintId
	//If the user uses a blank set of keys, we are importing everything, so, we do not want to update the list.
	state.Keys = plan.Keys
	state.LoadApiData(ctx, api, false, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceDatacenterPropertySet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))

	var api *apstra.TwoStageL3ClosPropertySet
	keys := make([]string, len(plan.Keys.Elements()))
	resp.Diagnostics.Append(plan.Keys.ElementsAs(ctx, &keys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Update Property Set
	err = bpClient.UpdatePropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()), keys...)
	if err != nil {
		resp.Diagnostics.AddError("error updating Property Set", err.Error())
		return
	}
	var ace apstra.ApstraClientErr
	// set state
	api, err = bpClient.GetPropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"DatacenterPropertySet not found",
			fmt.Sprintf("DatacenterPropertySet with ID %q not found. This should not happen", plan.Id.ValueString()))
		return
	}
	var state blueprint.DatacenterPropertySet
	state.Id = types.StringValue(string(api.Id))
	state.BlueprintId = plan.BlueprintId
	state.Keys = plan.Keys
	state.LoadApiData(ctx, api, false, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete resource
func (o *resourceDatacenterPropertySet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))

	// Delete Property Set by calling API
	err = bpClient.DeletePropertySet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != apstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Property Set", err.Error())
			return
		}
	}
}
