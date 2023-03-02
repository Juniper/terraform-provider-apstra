package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	var plan blueprint.PoolAllocation
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Validate(ctx, o.client, &resp.Diagnostics)
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

	// Parse 'role' into a ResourceGroupName
	var rgName goapstra.ResourceGroupName
	err = rgName.FromString(plan.Role.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error parsing role %q", plan.Role.ValueString()), err.Error())
		return
	}

	// Collect the current allocations for the given resource name+type
	rga, err := client.GetResourceAllocation(ctx, &goapstra.ResourceGroup{
		Type: rgName.Type(),
		Name: rgName,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error retrieving resource allocation for %q in blueprint %q", rgName.String(), client.Id()),
			err.Error())
		return
	}

	// If the desired pool is on the list, then we're done.
	for _, poolId := range rga.PoolIds {
		if poolId.String() == plan.PoolId.ValueString() {
			// nothing to do because the pool is already allocated!
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
	}

	// Add our pool to the list and send it back to Apstra
	rga.PoolIds = append(rga.PoolIds, goapstra.ObjectId(plan.PoolId.ValueString()))
	err = client.SetResourceAllocation(ctx, rga)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error allocating pool %q to role %q in blueprint %q",
				plan.PoolId.ValueString(), plan.Role.ValueString(), plan.BlueprintId.ValueString()),
			err.Error(),
		)
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourcePoolAllocation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (o *resourcePoolAllocation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (o *resourcePoolAllocation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
