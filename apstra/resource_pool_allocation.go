package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourcePoolAllocation{}
var _ resource.ResourceWithValidateConfig = &resourcePoolAllocation{}

type resourcePoolAllocation struct {
	client *goapstra.Client
}

func (o *resourcePoolAllocation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint_resource_pool_allocation"
}

func (o *resourcePoolAllocation) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourcePoolAllocation) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource allocates a resource pool to a role within a Blueprint.",
		Attributes:          blueprint.PoolAllocation{}.ResourceAttributes(),
	}
}

func (o *resourcePoolAllocation) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Cannot proceed without a client
	if o.client == nil {
		return
	}

	// Retrieve values from config
	var config blueprint.PoolAllocation
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Validate(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourcePoolAllocation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.PoolAllocation
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the plan (depends on state at Apstra)
	plan.Validate(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error creating client for Apstra Blueprint", err.Error())
	}

	// Create a resource allocation request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the new allocation
	err = client.SetResourceAllocation(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error setting resource allocation", err.Error())
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourcePoolAllocation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.PoolAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error creating client for Apstra Blueprint", err.Error())
		return
	}

	// Create an allocation request (because it's got the ResourceGroup object inside)
	allocationRequest := state.Request(ctx, &resp.Diagnostics)
	apiData, err := client.GetResourceAllocation(ctx, &allocationRequest.ResourceGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error getting %q resource allocation", allocationRequest.ResourceGroup.Name.String()),
			err.Error())
		return
	}

	// Load the API response into state
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourcePoolAllocation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.PoolAllocation
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the plan (depends on state at Apstra)
	plan.Validate(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error creating client for Apstra Blueprint", err.Error())
	}

	// Create a resource allocation request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the new allocation
	err = client.SetResourceAllocation(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error setting resource allocation", err.Error())
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourcePoolAllocation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.PoolAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Clear the poolIds so they'll get un-allocated
	state.PoolIds = types.SetNull(types.StringType)

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error creating client for Apstra Blueprint", err.Error())
	}

	// Create a resource allocation request
	request := state.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the empty allocation
	err = client.SetResourceAllocation(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error setting resource allocation", err.Error())
	}
}
