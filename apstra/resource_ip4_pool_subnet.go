package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

var _ resource.ResourceWithConfigure = &resourceIp4PoolSubnet{}
var _ resource.ResourceWithValidateConfig = &resourceIp4PoolSubnet{}

type resourceIp4PoolSubnet struct {
	client *goapstra.Client
}

func (o *resourceIp4PoolSubnet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_pool_subnet"
}

func (o *resourceIp4PoolSubnet) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceIp4PoolSubnet) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource allocates a CIDR block (Apstra web UI uses the term 'subnet') to an IPv4 resource pool",
		Attributes: map[string]tfsdk.Attribute{
			"pool_id": {
				MarkdownDescription: "Apstra ID of IPv4 pool to which the addresses should be allocated",
				Type:                types.StringType,
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"cidr": {
				MarkdownDescription: "IPv4 allocation in CIDR notation",
				Type:                types.StringType,
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
		},
	}, nil
}

func (o *resourceIp4PoolSubnet) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config rIp4PoolSubnet
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, parsed, err := net.ParseCIDR(config.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("cidr"),
			fmt.Sprintf("error parsing CIDR block '%s'", config.Cidr.Value),
			err.Error())
	}

	if config.Cidr.Value != parsed.String() {
		resp.Diagnostics.AddAttributeError(
			path.Root("cidr"),
			"CIDR block formatting problem",
			fmt.Sprintf("you entered '%s', did you mean '%s'?",
				config.Cidr.Value, parsed.String()))
	}
}

func (o *resourceIp4PoolSubnet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rIp4PoolSubnet
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the supplied cidr text
	_, parsed, err := net.ParseCIDR(plan.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("cidr"),
			fmt.Sprintf("error parsing CIDR block '%s'",
				plan.Cidr.Value), err.Error())
		return
	}

	// Add the new subnet to the pool
	err = o.client.AddSubnetToIp4Pool(ctx, goapstra.ObjectId(plan.PoolId.Value), parsed)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrExists) { // these are okay
			resp.Diagnostics.AddError(
				"error creating new IPv4 Pool Subnet",
				"Could not create IPv4 Pool Subnet, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Set State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceIp4PoolSubnet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state rIp4PoolSubnet
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the state cidr text
	_, parsedFromState, err := net.ParseCIDR(state.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("cidr"),
			fmt.Sprintf("error parsing CIDR block '%s'",
				state.Cidr.Value), err.Error())
		return
	}

	// Get IP pool info from API and then update what is in state from what the API returns
	pool, err := o.client.GetIp4Pool(ctx, goapstra.ObjectId(state.PoolId.Value))
	ace := goapstra.ApstraClientErr{}
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading parent IPv4 pool",
				fmt.Sprintf("could not read IPv4 pool '%s' - %s", state.PoolId.Value, err),
			)
			return
		}
	}

	subnetIdx := -1
	for i, subnet := range pool.Subnets {
		if subnet.Network.String() == state.Cidr.Value {
			// the result we want
			subnetIdx = i
			break
		}
		if subnet.Network.Contains(parsedFromState.IP) || parsedFromState.Contains(subnet.Network.IP) {
			// overlap - this is not our subnet, but it's in the way
			state.Cidr = types.String{Value: subnet.Network.String()}
			subnetIdx = i
			break
		}
	}

	if subnetIdx < 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Cidr = types.String{Value: pool.Subnets[subnetIdx].Network.String()}

	// Reset state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	return
}

func (o *resourceIp4PoolSubnet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	var plan rIp4PoolSubnet
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state rIp4PoolSubnet
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, parsedFromPlan, err := net.ParseCIDR(plan.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddError("error parsing value from plan", err.Error())
	}

	// fetch parent pool from API
	pool, err := o.client.GetIp4Pool(ctx, goapstra.ObjectId(state.PoolId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// parent resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading IPv4 pool", err.Error())
			return
		}
	}

	var subnetsRequest []goapstra.NewIpSubnet
	// loop over subnets in pool response, populating subnetsRequest as we go...
	// DO NOT copy the planned subnet, nor anything that overlaps with it.
	for _, subnet := range pool.Subnets {
		if subnet.Network.String() == parsedFromPlan.String() {
			// our exact subnet - this shouldn't happen during Update() but whatever
			continue
		}
		if subnet.Network.Contains(parsedFromPlan.IP) || parsedFromPlan.Contains(subnet.Network.IP) {
			// overlap - this is not our subnet, but it's in the way
			continue
		}
		subnetsRequest = append(subnetsRequest, goapstra.NewIpSubnet{Network: subnet.Network.String()})
	}

	// finally, append our intended subnet to the subnetsRequest
	subnetsRequest = append(subnetsRequest, goapstra.NewIpSubnet{Network: parsedFromPlan.String()})

	err = o.client.UpdateIp4Pool(ctx, goapstra.ObjectId(state.PoolId.Value), &goapstra.NewIpPoolRequest{
		DisplayName: pool.DisplayName,
		Subnets:     subnetsRequest,
	})
}

func (o *resourceIp4PoolSubnet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rIp4PoolSubnet
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, stateSubnet, err := net.ParseCIDR(state.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddError("error reading parsing IPv4 CIDR from state", err.Error())
		return
	}

	// Delete IPv4 pool subnet by calling API
	err = o.client.DeleteSubnetFromIp4Pool(ctx, goapstra.ObjectId(state.PoolId.Value), stateSubnet)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// IPv4 pool subnet deleted outside terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error removing IPv4 pool subnet", err.Error())
			return
		}
	}
}

type rIp4PoolSubnet struct {
	PoolId types.String `tfsdk:"pool_id"`
	Cidr   types.String `tfsdk:"cidr"`
}
