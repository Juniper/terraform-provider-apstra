package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	"terraform-provider-apstra/apstra/resources"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceAsnPool{}
var _ resource.ResourceWithValidateConfig = &resourceAsnPool{}

type resourceIpv4Pool struct {
	client *apstra.Client
}

func (o *resourceIpv4Pool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv4_pool"
}

func (o *resourceIpv4Pool) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceIpv4Pool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an IPv4 resource pool",
		Attributes:          resources.Ipv4Pool{}.ResourceAttributesWrite(),
	}
}

func (o *resourceIpv4Pool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config resources.Ipv4Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Subnets.IsUnknown() {
		return
	}

	subnets := make([]resources.Ipv4PoolSubnet, len(config.Subnets.Elements()))
	d := config.Subnets.ElementsAs(ctx, &subnets, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jNets []*net.IPNet // Each subnet will be checked for overlap with members of jNets, then appended to jNets
	for i := range subnets {
		// setVal is used to path AttributeErrors correctly
		setVal, d := types.ObjectValueFrom(ctx, resources.Ipv4PoolSubnet{}.AttrTypes(), &subnets[i])
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if subnets[i].Network.IsUnknown() {
			continue
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

func (o *resourceIpv4Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan resources.Ipv4Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new IPv4 Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateIp4Pool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new IPv4 Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	var pool *apstra.IpPool
	for { // loop until creation complete
		pool, err = o.client.GetIp4Pool(ctx, id)
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddError(
					"IPv4 Pool not found",
					fmt.Sprintf("Just-created IPv4 Pool with ID %q not found", id))
				return
			}
			resp.Diagnostics.AddError("Error retrieving IPv4 Pool", err.Error())
			return
		}
		if pool.Status != apstra.PoolStatusCreating {
			break
		}
	}

	// create state object
	var state resources.Ipv4Pool
	state.LoadApiData(ctx, pool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceIpv4Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state resources.Ipv4Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Ipv4 pool from API and then update what is in state from what the API returns
	p, err := o.client.GetIp4Pool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading IPv4 pool", err.Error())
		return
	}

	// create new state object
	var newState resources.Ipv4Pool
	newState.LoadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceIpv4Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan resources.Ipv4Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update IPv4 Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	err := o.client.UpdateIp4Pool(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating IPv4 Pool", err.Error())
		return
	}

	// read pool back from Apstra to get usage statistics
	p, err := o.client.GetIp4Pool(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
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
	var state resources.Ipv4Pool
	state.LoadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete resource
func (o *resourceIpv4Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resources.Ipv4Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete IPv4 pool by calling API
	err := o.client.DeleteIp4Pool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			"error deleting IPv4 pool", err.Error())
	}
}
