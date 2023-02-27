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

type resourceIpv6Pool struct {
	client *goapstra.Client
}

func (o *resourceIpv6Pool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_pool"
}

func (o *resourceIpv6Pool) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = resourceGetClient(ctx, req, resp)
}

func (o *resourceIpv6Pool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an IPv6 resource pool",
		Attributes:          ipv6Pool{}.resourceAttributesWrite(),
	}
}

func (o *resourceIpv6Pool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config ipv6Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnets := make([]ipv6PoolSubnet, len(config.Subnets.Elements()))
	d := config.Subnets.ElementsAs(ctx, &subnets, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jNets []*net.IPNet // Each subnet will be checked for overlap with members of jNets, then appended to jNets
	for i := range subnets {
		// setVal is used to path AttributeErrors correctly
		setVal, d := types.ObjectValueFrom(ctx, ipv6PoolSubnet{}.attrTypes(), &subnets[i])
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// parse the subnet string
		_, iNet, err := net.ParseCIDR(subnets[i].Network.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("subnets").AtSetValue(setVal),
				"failure parsing CIDR notation", fmt.Sprintf("error parsing %q - %s", subnets[i], err.Error()))
			return
		}

		// insist the user give us the all-zeros host address: 192.168.1.0/24 not 192.168.1.50/24
		if iNet.String() != subnets[i].Network.ValueString() {
			resp.Diagnostics.AddAttributeError(
				path.Root("subnets").AtSetValue(setVal),
				errInvalidConfig,
				fmt.Sprintf("%q doesn't specify a network base address. Did you mean %q?",
					subnets[i].Network.ValueString(), iNet.String()),
			)
		}

		// check for overlaps with previous subnets
		for j := range jNets {
			if iNet.Contains(jNets[j].IP) || jNets[j].Contains(iNet.IP) {
				resp.Diagnostics.AddAttributeError(
					path.Root("subnets"),
					"pool has overlapping subnets",
					fmt.Sprintf("subnets %q and %q overlap", iNet.String(), jNets[j].String()))
				return
			}
		}

		// no overlap. append iNet to the jNets slice
		jNets = append(jNets, iNet)
	}
}

func (o *resourceIpv6Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan ipv6Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new IPv6 Pool
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateIp6Pool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new IPv6 Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	var ace goapstra.ApstraClientErr
	p, err := o.client.GetIp6Pool(ctx, id)
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"IPv6 Pool not found",
				fmt.Sprintf("Just-created IPv6 Pool with ID %q not found", id))
			return
		}
		resp.Diagnostics.AddError("Error retrieving IPv6 Pool", err.Error())
		return
	}

	// create state object
	var state ipv6Pool
	state.loadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceIpv6Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state ipv6Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Ipv6 pool from API and then update what is in state from what the API returns
	p, err := o.client.GetIp6Pool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading IPv6 pool", err.Error())
			return
		}
	}

	// create new state object
	var newState ipv6Pool
	newState.loadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceIpv6Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan ipv6Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update IPv6 Pool
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var ace goapstra.ApstraClientErr
	err := o.client.UpdateIp6Pool(ctx, goapstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.State.RemoveResource(ctx)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("error updating IPv6 Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	p, err := o.client.GetIp6Pool(ctx, goapstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"IPv6 Pool not found",
				fmt.Sprintf("Recently updated IPv6 Pool with ID %q not found", plan.Id.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Error retrieving IPv6 Pool", err.Error())
		return
	}

	// create new state object
	var state ipv6Pool
	state.loadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete resource
func (o *resourceIpv6Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state ipv6Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete IPv6 pool by calling API
	err := o.client.DeleteIp6Pool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound {
			resp.Diagnostics.AddError(
				"error deleting IPv6 pool", err.Error())
		}
	}
}
