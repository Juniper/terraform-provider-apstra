package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

var _ resource.ResourceWithConfigure = &resourceAsnPool{}
var _ resource.ResourceWithValidateConfig = &resourceAsnPool{}

type resourceIp4Pool struct {
	client *goapstra.Client
}

func (o *resourceIp4Pool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_pool"
}

func (o *resourceIp4Pool) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = resourceGetClient(ctx, req, resp)
}

func (o *resourceIp4Pool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an IPv4 resource pool",
		Attributes:          ip4Pool{}.resourceAttributesWrite(),
	}
}

func (o *resourceIp4Pool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config ip4Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnets := make([]ip4PoolSubnet, len(config.Subnets.Elements()))
	d := config.Subnets.ElementsAs(ctx, &subnets, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jNets []*net.IPNet
	for i, subnet := range subnets {
		setVal, d := types.ObjectValueFrom(ctx, subnet.attrTypes(), &subnet)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, iNet, err := net.ParseCIDR(subnets[i].CIDR.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("subnets").AtSetValue(setVal),
				"failure parsing cidr notation", fmt.Sprintf("error parsing %q - %s", subnets[i], err.Error()))
			return
		}

		for j, jNet := range jNets {
			if iNet.Contains(jNet.IP) || jNet.Contains(iNet.IP) {
				resp.Diagnostics.AddAttributeError(
					path.Root("subnets"),
					"pool has overlapping subnets",
					fmt.Sprintf("subnets %q and %q overlap", subnets[i], subnets[j]))
				return
			}
		}

		// no overlap. append iNet to the jNets slice
		jNets = append(jNets, iNet)

		//for j := range subnets {
		//	if j == i {
		//		continue // don't compare a subnet to itself
		//	}
		//	_, jNet, err := net.ParseCIDR(subnets[j])
		//	if err != nil {
		//		resp.Diagnostics.AddAttributeError(
		//			path.Root("subnets").AtSetValue(types.StringValue(subnets[i])),
		//			"failure parsing cidr notation", fmt.Sprintf("error parsing '%s' - %s", subnets[i], err.Error()))
		//		return
		//	}
		//	if iNet.Contains(jNet.IP) || jNet.Contains(iNet.IP) {
		//		resp.Diagnostics.AddAttributeError(
		//			path.Root("subnets"),
		//			"pool has overlapping subnets",
		//			fmt.Sprintf("subnet '%s' and '%s' overlap", subnets[i], subnets[j]))
		//		return
		//	}
		//}
	}
}

func (o *resourceIp4Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan ip4Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new IPv4 Pool
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateIp4Pool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new IPv4 Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	var ace goapstra.ApstraClientErr
	p, err := o.client.GetIp4Pool(ctx, id)
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"IPv4 Pool not found",
				fmt.Sprintf("Just-created IPv4 Pool with ID %q not found", id))
			return
		}
		resp.Diagnostics.AddError("Error retrieving IPv4 Pool", err.Error())
		return
	}

	// create state object
	var state ip4Pool
	state.loadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceIp4Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state ip4Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Ip4 pool from API and then update what is in state from what the API returns
	p, err := o.client.GetIp4Pool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading IPv4 pool", err.Error())
			return
		}
	}

	// create new state object
	var newState ip4Pool
	newState.loadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceIp4Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan ip4Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update IPv4 Pool
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var ace goapstra.ApstraClientErr
	err := o.client.UpdateIp4Pool(ctx, goapstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.State.RemoveResource(ctx)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("error updating IPv4 Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	p, err := o.client.GetIp4Pool(ctx, goapstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"IPv4 Pool not found",
				fmt.Sprintf("Recently updated IPv4 Pool with ID %q not found", plan.Id.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Error retrieving IPv4 Pool", err.Error())
		return
	}

	// create new state object
	var state ip4Pool
	state.loadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceIp4Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state ip4Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete IPv4 pool by calling API
	err := o.client.DeleteIp4Pool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound {
			resp.Diagnostics.AddError(
				"error deleting IPv4 pool", err.Error())
		}
		return
	}
}
