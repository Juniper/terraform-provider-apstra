package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/resources"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceIntegerPool{}
var _ resource.ResourceWithValidateConfig = &resourceIntegerPool{}

type resourceIntegerPool struct {
	client *apstra.Client
}

func (o *resourceIntegerPool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integer_pool"
}

func (o *resourceIntegerPool) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceIntegerPool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an Integer resource pool",
		Attributes:          resources.IntegerPool{}.ResourceAttributes(),
	}
}

func (o *resourceIntegerPool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config resources.IntegerPool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validation not possible when ranges is unknown
	if config.Ranges.IsUnknown() {
		return
	}

	// validation not possible when any individual range is unknown
	for _, v := range config.Ranges.Elements() {
		if v.IsUnknown() {
			return
		}
	}

	// extract ranges
	var poolRanges []resources.IntegerPoolRange
	resp.Diagnostics.Append(config.Ranges.ElementsAs(ctx, &poolRanges, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var okayRanges apstra.IntRanges
	for _, poolRange := range poolRanges {
		// validation not possible without first and last values
		if poolRange.First.IsUnknown() || poolRange.Last.IsUnknown() {
			return
		}

		// extract set value for use in error pathing.
		// Note this doesn't currently work. https://github.com/hashicorp/terraform/issues/33491
		setVal, d := types.ObjectValueFrom(ctx, poolRange.AttrTypes(), &poolRange)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// grab the values from the range
		first := uint32(poolRange.First.ValueInt64())
		last := uint32(poolRange.Last.ValueInt64())

		// check whether this range overlaps previous ranges
		if okayRanges.Overlaps(apstra.IntRangeRequest{First: first, Last: last}) {
			resp.Diagnostics.AddAttributeError(
				path.Root("ranges").AtSetValue(setVal),
				"Integer range collision",
				fmt.Sprintf("Integer range %d-%d overlaps with another range in this pool", first, last),
			)
			return
		}

		// no overlap! append this range to the list for future overlap checks
		okayRanges = append(okayRanges, apstra.IntRange{First: first, Last: last})
	}
}

func (o *resourceIntegerPool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan resources.IntegerPool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Integer Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateIntegerPool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new Integer Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	var pool *apstra.IntPool
	for {
		pool, err = o.client.GetIntegerPool(ctx, id)
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"Integer Pool not found",
					fmt.Sprintf("Just-created Integer Pool with ID %q not found", id))
				return
			}
			resp.Diagnostics.AddError("Error retrieving Integer Pool", err.Error())
			return
		}
		if pool.Status != apstra.PoolStatusCreating {
			break
		}
	}

	// create state object
	var state resources.IntegerPool
	state.LoadApiData(ctx, pool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceIntegerPool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state resources.IntegerPool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Integer pool from API and then update what is in state from what the API returns
	p, err := o.client.GetIntegerPool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading Integer pool", err.Error())
		return
	}

	// create state object
	var newState resources.IntegerPool
	newState.LoadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceIntegerPool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan resources.IntegerPool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Integer Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	err := o.client.UpdateIntegerPool(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Integer Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	p, err := o.client.GetIntegerPool(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Integer Pool not found",
				fmt.Sprintf("Recently updated Integer Pool with ID %q not found", plan.Id.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Error retrieving Integer Pool", err.Error())
		return
	}

	// create new state object
	var state resources.IntegerPool
	state.LoadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceIntegerPool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resources.IntegerPool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Integer pool by calling API
	err := o.client.DeleteIntegerPool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Integer pool", err.Error())
		return
	}
}
