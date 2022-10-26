package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"math"
)

const (
	minAsn = 1              // rfc4893 says 0 is okay, but apstra says "Must be between 1 and 4294967295"
	maxAsn = math.MaxUint32 // 4294967295 rfc4893
)

var _ resource.ResourceWithConfigure = &resourceAsnPoolRange{}
var _ resource.ResourceWithValidateConfig = &resourceAsnPoolRange{}

type resourceAsnPoolRange struct {
	client *goapstra.Client
}

func (o *resourceAsnPoolRange) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pool_range"
}

func (o *resourceAsnPoolRange) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errResourceConfigureProviderDataDetail,
			fmt.Sprintf(errResourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *resourceAsnPoolRange) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"pool_id": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"first": {
				Type:     types.Int64Type,
				Required: true,
				//PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Validators: []tfsdk.AttributeValidator{int64validator.Between(minAsn-1, maxAsn+1)},
			},
			"last": {
				Type:     types.Int64Type,
				Required: true,
				//PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Validators: []tfsdk.AttributeValidator{int64validator.Between(minAsn-1, maxAsn+1)},
			},
		},
	}, nil
}

func (o *resourceAsnPoolRange) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config rAsnPoolRange
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.First.Value > config.Last.Value {
		resp.Diagnostics.AddError(
			"swap 'first' and 'last'",
			fmt.Sprintf("first (%d) cannot be greater than last (%d)", config.First.Value, config.Last.Value),
		)
	}
}

func (o *resourceAsnPoolRange) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rAsnPoolRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new ASN Pool Range
	err := o.client.CreateAsnPoolRange(ctx, goapstra.ObjectId(plan.PoolId.Value), &goapstra.IntRangeRequest{
		First: uint32(plan.First.Value),
		Last:  uint32(plan.Last.Value),
	})
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrExists) { // these are okay
			resp.Diagnostics.AddError(
				"error creating new ASN Pool Range", err.Error())
			return
		}
	}

	// Set State
	diags = resp.State.Set(ctx, rAsnPoolRange{
		PoolId: types.String{Value: plan.PoolId.Value},
		First:  types.Int64{Value: plan.First.Value},
		Last:   types.Int64{Value: plan.Last.Value},
	})
	resp.Diagnostics.Append(diags...)
}

func (o *resourceAsnPoolRange) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rAsnPoolRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool info from API and then update what is in state from what the API returns
	asnPool, err := o.client.GetAsnPool(ctx, goapstra.ObjectId(state.PoolId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// parent resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading ASN pool",
				fmt.Sprintf("error reading parent ASN pool ID '%s' - %s", state.PoolId.Value, err),
			)
			return
		}
	}

	indexOf := asnPool.Ranges.IndexOf(goapstra.IntRange{
		First: uint32(state.First.Value),
		Last:  uint32(state.Last.Value),
	})

	if indexOf < 0 { // no exact match range in pool - what happened to it?
		// we assume that any range overlapping our intended range is *this pool*, but edited.
		// really need range IDs here.
		for _, testRange := range asnPool.Ranges {
			if goapstra.IntOverlap(goapstra.IntRange{
				First: uint32(state.First.Value),
				Last:  uint32(state.Last.Value),
			}, testRange) {
				// overlapping pool found!  -- we'll choose to recognize it as the pool we're looking for, but edited.
				state.First = types.Int64{Value: int64(testRange.First)}
				state.Last = types.Int64{Value: int64(testRange.Last)}
				diags = resp.State.Set(ctx, &rAsnPoolRange{
					PoolId: types.String{Value: state.PoolId.Value},
					First:  types.Int64{Value: int64(testRange.First)},
					Last:   types.Int64{Value: int64(testRange.Last)},
				})
				resp.Diagnostics.Append(diags...)
				return
			}
		}
	}

	// Reset state using old data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	return
}

func (o *resourceAsnPoolRange) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	var plan rAsnPoolRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state rAsnPoolRange
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch parent pool from API
	asnPool, err := o.client.GetAsnPool(ctx, goapstra.ObjectId(state.PoolId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// parent resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading ASN pool",
				fmt.Sprintf("error reading parent ASN pool ID '%s' - %s", state.PoolId.Value, err),
			)
			return
		}
	}

	// we'll send 'ranges' to Apstra in our update
	var ranges []goapstra.IntfIntRange

	// copy current ranges which do not overlap our target
	for _, r := range asnPool.Ranges {
		if !goapstra.IntOverlap(r, goapstra.IntRange{
			First: uint32(plan.First.Value),
			Last:  uint32(plan.Last.Value),
		}) {
			ranges = append(ranges, r)
		}
	}

	// add our intended range to the slice
	ranges = append(ranges, goapstra.IntRange{
		First: uint32(plan.First.Value),
		Last:  uint32(plan.Last.Value),
	})

	data, err := json.Marshal(&goapstra.AsnPoolRequest{
		DisplayName: asnPool.DisplayName,
		Ranges:      ranges,
	})
	if err != nil {
		resp.Diagnostics.AddError("error in json marshal", err.Error())
	}
	tflog.Trace(ctx, string(data))

	err = o.client.UpdateAsnPool(ctx, goapstra.ObjectId(state.PoolId.Value), &goapstra.AsnPoolRequest{
		DisplayName: asnPool.DisplayName,
		Ranges:      ranges,
	})
	if err != nil {
		resp.Diagnostics.AddError("error updating ASN pool range", err.Error())
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceAsnPoolRange) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rAsnPoolRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ASN pool range by calling API
	err := o.client.DeleteAsnPoolRange(ctx, goapstra.ObjectId(state.PoolId.Value), &goapstra.IntRangeRequest{
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

type rAsnPoolRange struct {
	PoolId types.String `tfsdk:"pool_id"`
	First  types.Int64  `tfsdk:"first"`
	Last   types.Int64  `tfsdk:"last"`
}
