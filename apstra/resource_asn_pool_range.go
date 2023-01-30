package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (o *resourceAsnPoolRange) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pool_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"first": schema.Int64Attribute{
				Required:   true,
				Validators: []validator.Int64{int64validator.Between(minAsn-1, maxAsn+1)},
			},
			"last": schema.Int64Attribute{
				Required:   true,
				Validators: []validator.Int64{int64validator.Between(minAsn-1, maxAsn+1)},
			},
		},
	}
}

func (o *resourceAsnPoolRange) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config rAsnPoolRange
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.First.ValueInt64() > config.Last.ValueInt64() {
		resp.Diagnostics.AddError(
			"swap 'first' and 'last'",
			fmt.Sprintf("first (%d) cannot be greater than last (%d)", config.First.ValueInt64(), config.Last.ValueInt64()),
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new ASN Pool Range
	err := o.client.CreateAsnPoolRange(ctx, goapstra.ObjectId(plan.PoolId.ValueString()), &goapstra.IntRangeRequest{
		First: uint32(plan.First.ValueInt64()),
		Last:  uint32(plan.Last.ValueInt64()),
	})
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrExists) { // these are okay
			resp.Diagnostics.AddError(
				"error creating new ASN Pool Range", err.Error())
			return
		}
	}

	// create state object
	state := rAsnPoolRange{
		PoolId: types.StringValue(plan.PoolId.ValueString()),
		First:  types.Int64Value(plan.First.ValueInt64()),
		Last:   types.Int64Value(plan.Last.ValueInt64()),
	}

	// set State
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceAsnPoolRange) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rAsnPoolRange
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool info from API and then update what is in state from what the API returns
	asnPool, err := o.client.GetAsnPool(ctx, goapstra.ObjectId(state.PoolId.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// parent resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading ASN pool",
				fmt.Sprintf("error reading parent ASN pool ID '%s' - %s", state.PoolId.ValueString(), err),
			)
			return
		}
	}

	indexOf := asnPool.Ranges.IndexOf(goapstra.IntRange{
		First: uint32(state.First.ValueInt64()),
		Last:  uint32(state.Last.ValueInt64()),
	})

	if indexOf < 0 { // no exact match range in pool - what happened to it?
		// we assume that any range overlapping our intended range is *this pool*, but edited.
		// really need range IDs here.
		for _, testRange := range asnPool.Ranges {
			if goapstra.IntRangeOverlap(goapstra.IntRange{
				First: uint32(state.First.ValueInt64()),
				Last:  uint32(state.Last.ValueInt64()),
			}, testRange) {
				// overlapping pool found!  -- we'll choose to recognize it as the pool we're looking for, but edited.
				state.First = types.Int64Value(int64(testRange.First))
				state.Last = types.Int64Value(int64(testRange.Last))
				resp.Diagnostics.Append(resp.State.Set(ctx, &rAsnPoolRange{
					PoolId: types.StringValue(state.PoolId.ValueString()),
					First:  types.Int64Value(int64(testRange.First)),
					Last:   types.Int64Value(int64(testRange.Last)),
				})...)
				return
			}
		}
	}

	// Reset state using old data
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	return
}

func (o *resourceAsnPoolRange) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	var plan rAsnPoolRange
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state rAsnPoolRange
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch parent pool from API
	asnPool, err := o.client.GetAsnPool(ctx, goapstra.ObjectId(state.PoolId.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// parent resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading ASN pool",
				fmt.Sprintf("error reading parent ASN pool ID '%s' - %s", state.PoolId.ValueString(), err),
			)
			return
		}
	}

	plannedRange := goapstra.IntRange{
		First: uint32(plan.First.ValueInt64()),
		Last:  uint32(plan.Last.ValueInt64()),
	}

	stateRange := goapstra.IntRange{
		First: uint32(state.First.ValueInt64()),
		Last:  uint32(state.Last.ValueInt64()),
	}

	// we'll send 'ranges' to Apstra in our update
	var ranges []goapstra.IntfIntRange

	// copy current ranges which do not overlap our target
	for _, r := range asnPool.Ranges {
		if goapstra.IntRangeOverlap(r, plannedRange) {
			continue // assume overlapping ranges are config drift
		}
		if goapstra.IntRangeEqual(r, stateRange) {
			continue // same as state? we're making a change here. omit.
		}
		ranges = append(ranges, r)
	}

	// add our intended range to the slice
	ranges = append(ranges, plannedRange)

	err = o.client.UpdateAsnPool(ctx, goapstra.ObjectId(state.PoolId.ValueString()), &goapstra.AsnPoolRequest{
		DisplayName: asnPool.DisplayName,
		Ranges:      ranges,
	})
	if err != nil {
		resp.Diagnostics.AddError("error updating ASN pool range", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceAsnPoolRange) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rAsnPoolRange
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ASN pool range by calling API
	err := o.client.DeleteAsnPoolRange(ctx, goapstra.ObjectId(state.PoolId.ValueString()), &goapstra.IntRangeRequest{
		First: uint32(state.First.ValueInt64()),
		Last:  uint32(state.Last.ValueInt64()),
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
					state.PoolId.ValueString(), state.First.ValueInt64(), state.Last.ValueInt64(), err),
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
