package apstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type resourceAsnPoolRangeType struct{}

func (r resourceAsnPoolRangeType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"pool_id": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			// todo: validator
			"first": {
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			// todo: validator
			"last": {
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
		},
	}, nil
}

func (r resourceAsnPoolRangeType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceAsnPoolRange{
		p: *(p.(*provider)),
	}, nil
}

type resourceAsnPoolRange struct {
	p provider
}

func (r resourceAsnPoolRange) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceAsnPoolRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.First.Value > plan.Last.Value {
		resp.Diagnostics.AddError(
			"create asn pool range input error",
			fmt.Sprintf("first ASN cannot be larger than last ASN, but %d>%d ", plan.First.Value, plan.Last.Value),
		)
		return
	}

	// Create new ASN Pool
	err := r.p.client.CreateAsnPoolRange(ctx, goapstra.ObjectId(plan.PoolId.Value), &goapstra.AsnRange{
		First: uint32(plan.First.Value),
		Last:  uint32(plan.Last.Value),
	})
	if err != nil {
		// todo ensure that CreateAsnPoolRange returns ace.Notfound
		// todo check for ace.Notfound
		resp.Diagnostics.AddError(
			"error creating new asn pool",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	// Set State
	diags = resp.State.Set(ctx, ResourceAsnPoolRange{
		PoolId: types.String{Value: plan.PoolId.Value},
		First:  types.Int64{Value: plan.First.Value},
		Last:   types.Int64{Value: plan.Last.Value},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceAsnPoolRange) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ResourceAsnPoolRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool info from API and then update what is in state from what the API returns
	found, err := r.p.client.AsnPoolRangeExists(ctx, goapstra.ObjectId(state.PoolId.Value), &goapstra.AsnRange{
		First: uint32(state.First.Value),
		Last:  uint32(state.Last.Value),
	})
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// ASN pool deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading ASN pool",
				fmt.Sprintf("could not read ASN pool '%s' - %s", state.PoolId.Value, err),
			)
			return
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Reset state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceAsnPoolRange) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// No update method because Read() will never report a state change, only
	// resource existence (or not)
}

// Delete resource
func (r resourceAsnPoolRange) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceAsnPoolRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ASN pool range by calling API
	err := r.p.client.DeleteAsnPoolRange(ctx, goapstra.ObjectId(state.PoolId.Value), &goapstra.AsnRange{
		First: uint32(state.First.Value),
		Last:  uint32(state.First.Value),
	})
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// ASN pool deleted outside of terraform, so the range within the pool is irrelevant
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error removing ASN pool range",
				fmt.Sprintf("could not read ASN pool '%s' while deleting range %d-%d- %s",
					state.PoolId.Value, state.First.Value, state.Last.Value, err),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r resourceAsnPoolRange) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	//Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
