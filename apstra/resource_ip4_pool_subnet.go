package apstra

import (
	"context"
	"errors"
	"fmt"
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

type resourceIp4PoolSubnetType struct{}

func (r resourceIp4PoolSubnetType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"pool_id": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			"cidr": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
				//Validators:    []tfsdk.AttributeValidator{int64validator.Between(minAsn, maxAsn)}, //todo validate cidr notation
			},
		},
	}, nil
}

func (r resourceIp4PoolSubnetType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceIp4Subnet{
		p: *(p.(*provider)),
	}, nil
}

type resourceIp4Subnet struct {
	p provider
}

func (r resourceIp4Subnet) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceIp4Subnet
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the supplied cidr text
	_, parsed, err := net.ParseCIDR(plan.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"error parsing CIDR string",
			fmt.Sprintf("Could parse '%s' : ", err.Error()),
		)
		return
	}

	// Add the new subnet to the pool
	err = r.p.client.AddSubnetToIp4Pool(ctx, goapstra.ObjectId(plan.PoolId.Value), parsed)
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
	diags = resp.State.Set(ctx, ResourceIp4Subnet{
		PoolId: types.String{Value: plan.PoolId.Value},
		Cidr:   types.String{Value: plan.Cidr.Value},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceIp4Subnet) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ResourceIp4Subnet
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, stateSubnet, err := net.ParseCIDR(state.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"error reading parsing IPv4 CIDR from state",
			fmt.Sprintf("could not parse '%s' - %s", state.Cidr.Value, err),
		)
		return
	}

	// Get IP pool info from API and then update what is in state from what the API returns
	pool, err := r.p.client.GetIp4Pool(ctx, goapstra.ObjectId(state.PoolId.Value))
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

	var foundSomething bool
	for _, foundSubnet := range pool.Subnets {
		if foundSubnet.Network.String() == state.Cidr.String() {
			// the result we want
			foundSomething = true
			break
		}
		if foundSubnet.Network.Contains(stateSubnet.IP) || stateSubnet.Contains(foundSubnet.Network.IP) {
			// overlap - this is not our subnet, but it's in the way
			state.Cidr = types.String{Value: foundSubnet.Network.String()}
			foundSomething = true
			break
		}
	}

	if foundSomething == false {
		resp.State.RemoveResource(ctx)
		return
	}

	// Reset state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	return
}

func (r resourceIp4Subnet) Update(_ context.Context, _ tfsdk.UpdateResourceRequest, _ *tfsdk.UpdateResourceResponse) {
	// No update method because Read() will never report a state change, only
	// resource existence (or not)
}

func (r resourceIp4Subnet) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceIp4Subnet
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, stateSubnet, err := net.ParseCIDR(state.Cidr.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"error reading parsing IPv4 CIDR from state",
			fmt.Sprintf("could not parse '%s' - %s", state.Cidr.Value, err),
		)
		return
	}

	// Delete IPv4 pool subnet by calling API
	err = r.p.client.DeleteSubnetFromIp4Pool(ctx, goapstra.ObjectId(state.PoolId.Value), stateSubnet)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// IPv4 pool subnet deleted outside terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error removing IPv4 pool subnet",
				fmt.Sprintf("could not read IPv4 pool '%s' while deleting subnet '%s' - %s",
					state.PoolId.Value, state.Cidr.Value, err),
			)
			return
		}
	}
}
