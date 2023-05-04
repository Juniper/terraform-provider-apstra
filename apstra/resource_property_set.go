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
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourcePropertySet{}

type resourcePropertySet struct {
	client *apstra.Client
}

func (o *resourcePropertySet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_property_set"
}

func (o *resourcePropertySet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourcePropertySet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Tag in the Apstra Design tab.",
		Attributes:          design.PropertySet{}.ResourceAttributes(),
	}
}

func (o *resourcePropertySet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}
	// Retrieve values from plan
	var plan design.PropertySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Convert the plan into an API Request
	psRequest := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	psid, err := o.client.CreatePropertySet(ctx, psRequest)
	if err != nil {
		resp.Diagnostics.AddError("error creating new PropertySet", err.Error())
		return
	}
	// Save the tag ID
	plan.Id = types.StringValue(psid.String())
	plan.Blueprints = types.SetNull(types.StringType)
	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourcePropertySet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config design.PropertySet
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.PropertySet
	var ace apstra.ApstraClientErr

	switch {
	case !config.Label.IsNull():
		api, err = o.client.GetPropertySetByLabel(ctx, config.Label.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"PropertySet not found",
				fmt.Sprintf("PropertySet with label %q not found", config.Label.ValueString()))
			return
		}
	case !config.Id.IsNull():
		api, err = o.client.GetPropertySet(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"PropertySet not found",
				fmt.Sprintf("PropertySet with ID %q not found", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving PropertySet", err.Error())
		return
	}

	// create new state object
	var state design.PropertySet
	state.Id = types.StringValue(api.Id.String())
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if utils.JSONEqual(state.Values, config.Values, &resp.Diagnostics) {

		state.Values = config.Values
	}
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	//
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourcePropertySet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan design.PropertySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	psReq := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Property Set
	err := o.client.UpdatePropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()), psReq)
	if err != nil {
		resp.Diagnostics.AddError("Request Made to ::", plan.Id.ValueString())
		resp.Diagnostics.AddError("error updating Property Set", err.Error())
		return
	}
	var ace apstra.ApstraClientErr
	// set state
	api, err := o.client.GetPropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"PropertySet not found",
			fmt.Sprintf("PropertySet with ID %q not found. This should not happen", plan.Id.ValueString()))
		return
	}
	var d design.PropertySet
	d.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if !utils.JSONEqual(d.Values, plan.Values, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error Update seems to have failed. Keys Not Updated.", plan.Values.ValueString())
		return
	}
	if d.Label != plan.Label {
		resp.Diagnostics.AddError("Error Update seems to have failed. Name Not Updated.", plan.Label.ValueString())
		return
	}
	plan.Blueprints = d.Blueprints
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourcePropertySet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state design.PropertySet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Property Set by calling API
	err := o.client.DeletePropertySet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != apstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Property Set", err.Error())
			return
		}
	}
}