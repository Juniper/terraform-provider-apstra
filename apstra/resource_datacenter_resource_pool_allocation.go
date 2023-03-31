package tfapstra

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourcePoolAllocation{}

type resourcePoolAllocation struct {
	client     *apstra.Client
	lockFunc   func(context.Context, string) error
	unlockFunc func(context.Context, string) error
}

func (o *resourcePoolAllocation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_resource_pool_allocation"
}

func (o *resourcePoolAllocation) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
	o.unlockFunc = ResourceGetBlueprintUnlockFunc(ctx, req, resp)
}

func (o *resourcePoolAllocation) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource allocates a resource pool to a role within a Blueprint.",
		Attributes:          blueprint.PoolAllocation{}.ResourceAttributes(),
	}
}

// example API transaction
// {
//  "id": "rag_ip_sz:mWSoTAQpY5DSifaDZ50,leaf_loopback_ips",
//  "type": "ip",
//  "name": "sz:mWSoTAQpY5DSifaDZ50,leaf_loopback_ips",
//  "pool_ids": [ "66e3fd04-cbb1-4262-8556-01335dd9d040" ]
//}

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

	// Ensure the blueprint exists.
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(plan.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.Diagnostics.AddError("blueprint not found", fmt.Sprintf("blueprint %q not found", plan.BlueprintId.ValueString()))
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error creating client for Apstra Blueprint", err.Error())
	}

	// Create a resource allocation request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the new allocation
	switch {
	case !plan.RoutingZoneId.IsNull():
		err = client.SetResourceAllocation(ctx, request)
	default:
		err = client.SetResourceAllocation(ctx, request)
	}
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

	// Ensure the blueprint still exists.
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
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

	// Ensure the blueprint still exists.
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(plan.BlueprintId.ValueString()), &resp.Diagnostics) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
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

	// No need to proceed if the blueprint no longer exists
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Clear the poolIds so they'll get un-allocated
	state.PoolIds = types.SetNull(types.StringType)

	// Create a blueprint client
	client, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
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
