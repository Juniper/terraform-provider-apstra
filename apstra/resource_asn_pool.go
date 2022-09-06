package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type resourceAsnPoolType struct{}

func (r resourceAsnPoolType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
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

func (r resourceAsnPoolType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceAsnPool{
		p: *(p.(*provider)),
	}, nil
}

type resourceAsnPool struct {
	p provider
}

func (r resourceAsnPool) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceAsnPool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// prep tags for new ASN pool
	tags := asnPoolTagsFromPlan(plan.Tags)

	// Create new ASN Pool
	id, err := r.p.client.CreateAsnPool(ctx, &goapstra.AsnPoolRequest{
		DisplayName: plan.Name.Value,
		Tags:        tags,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new ASN Pool",
			"Could not create ASN Pool, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "created ASN pool", map[string]interface{}{"id": string(id)})

	// Generate resource state struct
	var result = ResourceAsnPool{
		Id:   types.String{Value: string(id)},
		Name: plan.Name,
		Tags: plan.Tags,
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceAsnPool) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ResourceAsnPool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool from API and then update what is in state from what the API returns
	asnPool, err := r.p.client.GetAsnPool(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading ASN pool",
				fmt.Sprintf("could not read ASN pool '%s' - %s", state.Id.Value, err),
			)
			return
		}
	}

	// Map response body to resource schema attribute
	// todo: error check state.Id vs. asnPool.Id
	state.Id = types.String{Value: string(asnPool.Id)}
	state.Name = types.String{Value: asnPool.DisplayName}
	state.Tags = asnPoolTagsFromApi(asnPool.Tags)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceAsnPool) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan ResourceAsnPool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state ResourceAsnPool
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch existing []goapstra.AsnRange
	//goland:noinspection GoPreferNilSlice
	poolFromApi, err := r.p.client.GetAsnPool(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.Diagnostics.AddError(
				fmt.Sprintf("cannot update %s", resourceAsnPoolName),
				fmt.Sprintf("error fetching existing ASN ranges - ASN pool '%s' not found", state.Id.Value),
			)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError(
			fmt.Sprintf("cannot update %s", resourceAsnPoolName),
			fmt.Sprintf("error fetching existing ASN ranges for ASN pool '%s' - %s", state.Id.Value, err.Error()),
		)
		return
	}

	// Generate API request body from plan
	send := &goapstra.AsnPoolRequest{
		DisplayName: plan.Name.Value,
		Tags:        asnPoolTagsFromPlan(plan.Tags),
	}
	send.Ranges = make([]goapstra.IntfAsnRange, len(poolFromApi.Ranges))
	for i, r := range poolFromApi.Ranges {
		send.Ranges[i] = r
	}

	// Create/Update ASN pool
	err = r.p.client.UpdateAsnPool(ctx, goapstra.ObjectId(state.Id.Value), send)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("cannot update %s", resourceAsnPoolName),
			fmt.Sprintf("cannot update %s '%s' - %s", resourceAsnPoolName, plan.Id.Value, err.Error()),
		)
		return
	}
	state.Name = plan.Name
	state.Tags = plan.Tags

	// Set new state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceAsnPool) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceAsnPool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool ID from state
	id := state.Id.Value

	// Delete ASN pool by calling API
	err := r.p.client.DeleteAsnPool(ctx, goapstra.ObjectId(id))
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

func asnPoolTagsFromPlan(in []types.String) []string {
	//goland:noinspection GoPreferNilSlice
	out := []string{}
	for _, t := range in {
		out = append(out, t.Value)
	}
	return out
}

func asnPoolTagsFromApi(in []string) []types.String {
	var out []types.String
	for _, t := range in {
		out = append(out, types.String{Value: t})
	}
	return out
}
