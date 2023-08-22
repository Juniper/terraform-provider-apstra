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

var _ resource.ResourceWithConfigure = &resourceVniPool{}
var _ resource.ResourceWithValidateConfig = &resourceVniPool{}

type resourceVniPool struct {
	client *apstra.Client
}

func (o *resourceVniPool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vni_pool"
}

func (o *resourceVniPool) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceVniPool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an VNI resource pool",
		Attributes:          resources.VniPool{}.ResourceAttributes(),
	}
}

func (o *resourceVniPool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config resources.VniPool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Ranges.IsUnknown() {
		return
	}

	poolRanges := make([]resources.VniPoolRange, len(config.Ranges.Elements()))
	d := config.Ranges.ElementsAs(ctx, &poolRanges, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var okayRanges apstra.IntRanges
	for _, poolRange := range poolRanges {
		setVal, d := types.ObjectValueFrom(ctx, poolRange.AttrTypes(), &poolRange)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		first := uint32(poolRange.First.ValueInt64())
		last := uint32(poolRange.Last.ValueInt64())

		// check whether this range overlaps previous ranges
		if okayRanges.Overlaps(apstra.IntRangeRequest{
			First: first,
			Last:  last,
		}) {
			resp.Diagnostics.AddAttributeError(
				path.Root("ranges").AtSetValue(setVal),
				"VNI range collision",
				fmt.Sprintf("VNI range %d - %d overlaps with another range in this pool", first, last),
			)
			return
		}

		// no overlap, append this range to the list for future overlap checks
		okayRanges = append(okayRanges, apstra.IntRange{
			First: first,
			Last:  last,
		})
	}
}

func (o *resourceVniPool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan resources.VniPool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new VNI Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateVniPool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new VNI Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	var pool *apstra.VniPool
	for {
		pool, err = o.client.GetVniPool(ctx, id)
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"VNI Pool not found",
					fmt.Sprintf("Just-created VNI Pool with ID %q not found", id))
				return
			}
			resp.Diagnostics.AddError("Error retrieving VNI Pool", err.Error())
			return
		}
		if pool.Status != apstra.PoolStatusCreating {
			break
		}
	}

	// create state object
	var state resources.VniPool
	state.LoadApiData(ctx, pool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceVniPool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state resources.VniPool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get VNI pool from API and then update what is in state from what the API returns
	p, err := o.client.GetVniPool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading VNI pool", err.Error())
		return
	}

	// create state object
	var newState resources.VniPool
	newState.LoadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceVniPool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan resources.VniPool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update VNI Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	err := o.client.UpdateVniPool(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating VNI Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	p, err := o.client.GetVniPool(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"VNI Pool not found",
				fmt.Sprintf("Recently updated VNI Pool with ID %q not found", plan.Id.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Error retrieving VNI Pool", err.Error())
		return
	}

	// create new state object
	var state resources.VniPool
	state.LoadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceVniPool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resources.VniPool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete VNI pool by calling API
	err := o.client.DeleteVniPool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting VNI pool", err.Error())
		return
	}
}
