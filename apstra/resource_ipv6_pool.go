package tfapstra

import (
	"context"
	"fmt"
	"net"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/resources"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceIpv6Pool{}
	_ resource.ResourceWithValidateConfig = &resourceIpv6Pool{}
	_ resourceWithSetClient               = &resourceIpv6Pool{}
)

type resourceIpv6Pool struct {
	client *apstra.Client
}

func (o *resourceIpv6Pool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_pool"
}

func (o *resourceIpv6Pool) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceIpv6Pool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryResources + "This resource creates an IPv6 resource pool",
		Attributes:          resources.Ipv6Pool{}.ResourceAttributes(),
	}
}

func (o *resourceIpv6Pool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config resources.Ipv6Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validation not possible when subnets is unknown
	if config.Subnets.IsUnknown() {
		return
	}

	// validation not possible when any individual range is unknown
	for _, v := range config.Subnets.Elements() {
		if v.IsUnknown() {
			return
		}
	}

	// extract Subnets
	var subnets []resources.Ipv6PoolSubnet
	resp.Diagnostics.Append(config.Subnets.ElementsAs(ctx, &subnets, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jNets []*net.IPNet
	// Each subnet will be checked for overlap with members of jNets
	// (j is inner loop iterator variable), then appended to jNets
	for i := range subnets {
		// skip unknown values
		if subnets[i].Network.IsUnknown() {
			continue
		}

		// parse the subnet string; error ignored because it's already been validated
		_, iNet, _ := net.ParseCIDR(subnets[i].Network.ValueString())

		// check for overlaps with previous subnets
		for _, jNet := range jNets {
			if iNet.Contains(jNet.IP) || jNet.Contains(iNet.IP) {
				// no return so we catch all overlap errors
				resp.Diagnostics.AddAttributeError(
					path.Root("subnets"),
					"pool has overlapping subnets",
					fmt.Sprintf("subnets %q and %q overlap", iNet.String(), jNet.String()))
			}
		}

		// no overlap. append iNet to the jNets slice
		jNets = append(jNets, iNet)
	}
}

func (o *resourceIpv6Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan resources.Ipv6Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new IPv6 Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateIp6Pool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new IPv6 Pool", err.Error())
		return
	}

	plan.Id = types.StringValue(id.String())
	plan.SetMutablesToNull(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceIpv6Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state resources.Ipv6Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Ipv6 pool from API and then update what is in state from what the API returns
	p, err := o.client.GetIp6Pool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading IPv6 pool", err.Error())
		return
	}

	// create new state object
	var newState resources.Ipv6Pool
	newState.LoadApiData(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newState.SetMutablesToNull(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceIpv6Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan resources.Ipv6Pool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update IPv6 Pool
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	err := o.client.UpdateIp6Pool(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating IPv6 Pool", err.Error())
		return
	}

	plan.SetMutablesToNull(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceIpv6Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resources.Ipv6Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete IPv6 pool by calling API
	err := o.client.DeleteIp6Pool(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			"error deleting IPv6 pool", err.Error())
	}
}

func (o *resourceIpv6Pool) setClient(client *apstra.Client) {
	o.client = client
}
