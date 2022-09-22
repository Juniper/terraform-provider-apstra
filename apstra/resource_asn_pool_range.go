package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math"
)

const (
	minAsn = 1              // rfc4893 says 0 is okay, but apstra says "Must be between 1 and 4294967295"
	maxAsn = math.MaxUint32 // 4294967295 rfc4893
)

type resourceAsnPoolRangeType struct{}

func (r resourceAsnPoolRange) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "apstra_asn_pool_range"
}

func (r resourceAsnPoolRange) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"pool_id": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"first": {
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Validators:    []tfsdk.AttributeValidator{int64validator.Between(minAsn, maxAsn)},
			},
			"last": {
				Type:          types.Int64Type,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Validators:    []tfsdk.AttributeValidator{int64validator.Between(minAsn, maxAsn)},
			},
		},
	}, nil
}

func (r resourceAsnPoolRangeType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceAsnPoolRange{
		p: *(p.(*Provider)),
	}, nil
}

func (r resourceAsnPoolRange) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config ResourceAsnPoolRange
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.First.Unknown || config.First.Null {
		return
	}

	if config.Last.Unknown || config.Last.Null {
		return
	}

	if config.First.Value > config.Last.Value {
		resp.Diagnostics.AddError(
			"swap 'first' and 'last'",
			fmt.Sprintf("first (%d) cannot be greater than last (%d)", config.First.Value, config.Last.Value),
		)
	}

}

type resourceAsnPoolRange struct {
	p Provider
}

func (r resourceAsnPoolRange) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	// todo: make this a validator
	if plan.First.Value > plan.Last.Value {
		resp.Diagnostics.AddError(
			"create asn pool range input error",
			fmt.Sprintf("first ASN cannot be larger than last ASN, but %d>%d ", plan.First.Value, plan.Last.Value),
		)
		return
	}

	// Create new ASN Pool Range
	err := r.p.client.CreateAsnPoolRange(ctx, goapstra.ObjectId(plan.PoolId.Value), &goapstra.AsnRangeRequest{
		First: uint32(plan.First.Value),
		Last:  uint32(plan.Last.Value),
	})
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrExists) { // these are okay
			resp.Diagnostics.AddError(
				"error creating new ASN Pool Range",
				"Could not create ASN Pool Range, unexpected error: "+err.Error(),
			)
			return
		}
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

func (r resourceAsnPoolRange) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ResourceAsnPoolRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool info from API and then update what is in state from what the API returns
	asnPool, err := r.p.client.GetAsnPool(ctx, goapstra.ObjectId(state.PoolId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading ASN pool",
				fmt.Sprintf("could not read ASN pool to check pool range range '%d-%d' - %s", state.First.Value, state.Last.Value, err),
			)
			return
		}
	}

	indexOf := asnPool.Ranges.IndexOf(goapstra.AsnRange{
		First: uint32(state.First.Value),
		Last:  uint32(state.Last.Value),
	})

	if indexOf >= 0 { // exact match range in pool - this is what we expect
		// Reset state
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	} else { // exact match range not found - maybe it's been edited?
		// we assume that any range overlapping our intended range is *this pool*, but edited.
		// really need range IDs here.
		for _, testRange := range asnPool.Ranges {
			if goapstra.AsnOverlap(goapstra.AsnRange{
				First: uint32(state.First.Value),
				Last:  uint32(state.Last.Value),
			}, testRange) {
				// overlapping pool found!  -- we'll choose to recognize it as the pool we're looking for, but edited.
				state.First = types.Int64{Value: int64(testRange.First)}
				state.Last = types.Int64{Value: int64(testRange.Last)}
				diags = resp.State.Set(ctx, &state)
				resp.Diagnostics.Append(diags...)
				return
			}
		}
	}
}

func (r resourceAsnPoolRange) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// No update method because Read() will never report a state change, only
	// resource existence (or not)
}

func (r resourceAsnPoolRange) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceAsnPoolRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ASN pool range by calling API
	err := r.p.client.DeleteAsnPoolRange(ctx, goapstra.ObjectId(state.PoolId.Value), &goapstra.AsnRangeRequest{
		First: uint32(state.First.Value),
		Last:  uint32(state.Last.Value),
	})
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// ASN pool deleted outside terraform, so the range within the pool is irrelevant
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
}
