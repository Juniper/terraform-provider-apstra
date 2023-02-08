package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

var _ resource.ResourceWithConfigure = &resourceAsnPool{}
var _ resource.ResourceWithValidateConfig = &resourceAsnPool{}

type resourceAsnPool struct {
	client *goapstra.Client
}

func (o *resourceAsnPool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pool"
}

func (o *resourceAsnPool) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceAsnPool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an ASN resource pool",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Apstra ID number of the resource pool",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Pool name displayed in the Apstra web UI",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				Required:            true,
			},
			"ranges": schema.SetNestedAttribute{
				MarkdownDescription: "Ranges mark the begin/end AS numbers available from the pool",
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: rAsnPoolRange{}.attributes(),
				},
			},
		},
	}
}

func (o *resourceAsnPool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config rAsnPool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolRanges := make([]rAsnPoolRange, len(config.Ranges.Elements()))
	d := config.Ranges.ElementsAs(ctx, &poolRanges, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var okayRanges goapstra.IntRanges
	for _, poolRange := range poolRanges {
		setVal, d := types.ObjectValueFrom(ctx, poolRange.attrTypes(), &poolRange)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// validate this range configuration
		poolRange.validateConfig(ctx, path.Root("ranges").AtSetValue(setVal), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		first := uint32(poolRange.First.ValueInt64())
		last := uint32(poolRange.Last.ValueInt64())

		// check whether this range overlaps previous ranges
		if okayRanges.Overlaps(goapstra.IntRangeRequest{
			First: first,
			Last:  last,
		}) {
			resp.Diagnostics.AddAttributeError(
				path.Root("ranges").AtSetValue(setVal),
				"ASN range collision",
				fmt.Sprintf("ASN range %d - %d overlaps with a another range in this pool", first, last),
			)
			return
		}

		// no overlap, append this range to the list for future overlap checks
		okayRanges = append(okayRanges, goapstra.IntRange{
			First: first,
			Last:  last,
		})
	}

}

func (o *resourceAsnPool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rAsnPool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new ASN Pool
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateAsnPool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new ASN Pool", err.Error())
		return
	}

	plan.Id = types.StringValue(string(id))

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceAsnPool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rAsnPool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool from API and then update what is in state from what the API returns
	pool, err := o.client.GetAsnPool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading ASN pool", err.Error())
			return
		}
	}

	// create state object
	var newState rAsnPool
	newState.loadApiResponse(ctx, pool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceAsnPool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan rAsnPool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update new ASN Pool
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	err := o.client.UpdateAsnPool(ctx, goapstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new ASN Pool", err.Error())
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceAsnPool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rAsnPool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ASN pool by calling API
	err := o.client.DeleteAsnPool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound {
			resp.Diagnostics.AddError(
				"error deleting ASN pool", err.Error())
		}
		return
	}
}

type rAsnPool struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Ranges types.Set    `tfsdk:"ranges"`
}

func (o *rAsnPool) loadApiResponse(ctx context.Context, in *goapstra.AsnPool, diags *diag.Diagnostics) {
	ranges := make([]rAsnPoolRange, len(in.Ranges))
	for i, poolRange := range in.Ranges {
		ranges[i].loadApiResponse(ctx, &poolRange, diags)
	}
	if diags.HasError() {
		return
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.DisplayName)
	o.Ranges = setValueOrNull(ctx, rAsnPoolRange{}.attrType(), ranges, diags)
}

func (o *rAsnPool) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.AsnPoolRequest {
	response := goapstra.AsnPoolRequest{
		DisplayName: o.Name.ValueString(),
		Ranges:      make([]goapstra.IntfIntRange, len(o.Ranges.Elements())),
	}

	poolRanges := make([]rAsnPoolRange, len(o.Ranges.Elements()))
	d := o.Ranges.ElementsAs(ctx, &poolRanges, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	for i, poolRange := range poolRanges {
		response.Ranges[i] = poolRange.request(ctx, diags)
	}

	return &response
}

type rAsnPoolRange struct {
	First types.Int64 `tfsdk:"first"`
	Last  types.Int64 `tfsdk:"last"`
}

func (o rAsnPoolRange) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"first": types.Int64Type,
		"last":  types.Int64Type,
	}
}

func (o rAsnPoolRange) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o rAsnPoolRange) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"first": schema.Int64Attribute{
			Required:   true,
			Validators: []validator.Int64{int64validator.Between(minAsn-1, maxAsn+1)},
		},
		"last": schema.Int64Attribute{
			Required:   true,
			Validators: []validator.Int64{int64validator.Between(minAsn-1, maxAsn+1)},
		},
	}
}

func (o *rAsnPoolRange) validateConfig(_ context.Context, path path.Path, diags *diag.Diagnostics) {
	if o.First.ValueInt64() > o.Last.ValueInt64() {
		diags.AddAttributeError(
			path,
			"swap 'first' and 'last'",
			fmt.Sprintf("first (%d) cannot be greater than last (%d)", o.First.ValueInt64(), o.Last.ValueInt64()),
		)
	}
}

func (o *rAsnPoolRange) loadApiResponse(_ context.Context, in *goapstra.IntRange, _ *diag.Diagnostics) {
	o.First = types.Int64Value(int64(in.First))
	o.Last = types.Int64Value(int64(in.Last))
}

func (o *rAsnPoolRange) request(_ context.Context, _ *diag.Diagnostics) goapstra.IntfIntRange {
	return &goapstra.IntRangeRequest{
		First: uint32(o.First.ValueInt64()),
		Last:  uint32(o.Last.ValueInt64()),
	}
}
