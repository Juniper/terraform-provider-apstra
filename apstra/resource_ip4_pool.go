package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceIp4PoolType struct{}

func (r resourceIp4Pool) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "apstra_ip4_pool"
}

func (r resourceIp4Pool) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"tags": {
				Optional: true,
				Type:     types.SetType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (r resourceIp4PoolType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceIp4Pool{
		p: *(p.(*Provider)),
	}, nil
}

type resourceIp4Pool struct {
	p Provider
}

func (r resourceIp4Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceIp4Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// prep tags for new pool
	tags := sliceTfStringToSliceString(plan.Tags)

	// Create new Pool
	id, err := r.p.client.CreateIp4Pool(ctx, &goapstra.NewIp4PoolRequest{
		DisplayName: plan.Name.Value,
		Tags:        tags,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new IPv4 Pool",
			"Could not create IPv4 Pool, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Id = types.String{Value: string(id)}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceIp4Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ResourceIp4Pool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Ip4 pool from API and then update what is in state from what the API returns
	pool, err := r.p.client.GetIp4Pool(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading IPv4 pool",
				fmt.Sprintf("could not read IPv4 pool '%s' - %s", state.Id.Value, err),
			)
			return
		}
	}

	// Map response body to resource schema attribute
	// todo: error check state.Id vs. asnPool.Id
	state.Id = types.String{Value: string(pool.Id)}
	state.Name = types.String{Value: pool.DisplayName}
	state.Tags = asnPoolTagsFromApi(pool.Tags)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceIp4Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan ResourceIp4Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state ResourceIp4Pool
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch existing []goapstra.AsnRange
	//goland:noinspection GoPreferNilSlice
	poolFromApi, err := r.p.client.GetIp4Pool(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.Diagnostics.AddError("API error",
				fmt.Sprintf("error fetching existing ASN ranges - ASN pool '%s' not found", state.Id.Value),
			)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("API error",
			fmt.Sprintf("error fetching existing ASN ranges for ASN pool '%s' - %s", state.Id.Value, err.Error()),
		)
		return
	}

	// Generate API request body from plan
	newReq := &goapstra.NewIp4PoolRequest{
		DisplayName: plan.Name.Value,
		Tags:        sliceTfStringToSliceString(plan.Tags),
	}

	for _, s := range poolFromApi.Subnets {
		newReq.Subnets = append(newReq.Subnets, goapstra.NewIp4Subnet{Network: s.Network.String()})
	}

	// Create/Update ASN pool
	err = r.p.client.UpdateIp4Pool(ctx, goapstra.ObjectId(state.Id.Value), newReq)
	if err != nil {
		resp.Diagnostics.AddError("API error",
			fmt.Sprintf("cannot update '%s' - %s", plan.Id.Value, err.Error()),
		)
		return
	}

	// Set new state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceIp4Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceIp4Pool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool ID from state
	id := state.Id.Value

	// Delete ASN pool by calling API
	err := r.p.client.DeleteIp4Pool(ctx, goapstra.ObjectId(id))
	if err != nil {
		resp.Diagnostics.AddError(
			"error deleting ASN pool",
			fmt.Sprintf("could not delete ASN pool '%s' - %s", id, err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}
