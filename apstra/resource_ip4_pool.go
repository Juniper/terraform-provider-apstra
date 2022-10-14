package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceAsnPool{}

type resourceIp4Pool struct {
	client *goapstra.Client
}

func (o *resourceIp4Pool) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "apstra_ip4_pool"
}

func (o *resourceIp4Pool) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceIp4Pool) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource creates an IPv4 resource pool",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Apstra ID number of the resource pool",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Pool name displayed in the Apstra web UI",
				Type:                types.StringType,
				Required:            true,
			},
		},
	}, nil
}

func (o *resourceIp4Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rIp4Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Ip4 Pool
	id, err := o.client.CreateIp4Pool(ctx, &goapstra.NewIpPoolRequest{
		DisplayName: plan.Name.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating new IPv4 Pool", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &rIp4Pool{
		Id:   types.String{Value: string(id)},
		Name: plan.Name,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceIp4Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rIp4Pool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Ip4 pool from API and then update what is in state from what the API returns
	pool, err := o.client.GetIp4Pool(ctx, goapstra.ObjectId(state.Id.Value))
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

	// Set state
	diags = resp.State.Set(ctx, &rIp4Pool{
		Id:   types.String{Value: string(pool.Id)},
		Name: types.String{Value: pool.DisplayName},
	})
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceIp4Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan rIp4Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state rIp4Pool
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentPool, err := o.client.GetIp4Pool(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning("API error",
				fmt.Sprintf("error fetching existing IPv4 pool - pool '%s' not found", state.Id.Value),
			)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("API error",
			fmt.Sprintf("error fetching IPv4 pool '%s' - %s", state.Id.Value, err.Error()),
		)
		return
	}

	// Generate API request body from plan
	send := &goapstra.NewIpPoolRequest{
		DisplayName: plan.Name.Value,
		Subnets:     make([]goapstra.NewIpSubnet, len(currentPool.Subnets)),
	}
	for i, s := range currentPool.Subnets {
		send.Subnets[i] = goapstra.NewIpSubnet{Network: s.Network.String()}
	}

	// Create/Update ASN pool
	err = o.client.UpdateIp4Pool(ctx, goapstra.ObjectId(state.Id.Value), send)
	if err != nil {
		resp.Diagnostics.AddError("error updating IPv4 pool", err.Error())
		return
	}

	// Set new state
	diags = resp.State.Set(ctx, &rIp4Pool{
		Id:   types.String{Value: state.Id.Value},
		Name: types.String{Value: plan.Name.Value},
	})
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (o *resourceIp4Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rIp4Pool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete IPv4 pool by calling API
	err := o.client.DeleteIp4Pool(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound {
			resp.Diagnostics.AddError(
				"error deleting IPv4 pool", err.Error())
		}
		return
	}
}

type rIp4Pool struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
