// Copyright (c) Juniper Networks, Inc., 2022. All rights reserved.

package tfapstra

import (
	"context"
	"fmt"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Updated resource implementation with SviIps support

// DatacenterVirtualNetworkResourceModel extends the basic model with SviIps
type DatacenterVirtualNetworkResourceModel struct {
	blueprint.DatacenterVirtualNetwork
	SviIps types.Set `tfsdk:"svi_ips"`
}

// Update resource schema to include SviIps
func (o *resourceDatacenterVirtualNetwork) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Use the extended schema with SviIps
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a Virtual Network within a Blueprint.",
		Attributes:          blueprint.DatacenterVirtualNetwork{}.ResourceAttributesWithSviIps(),
	}
}

// Update Create method to handle SviIps
func (o *resourceDatacenterVirtualNetwork) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan DatacenterVirtualNetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Create a request object with SviIps
	request := plan.DatacenterVirtualNetwork.RequestWithSviIps(ctx, plan.SviIps, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the virtual network
	id, err := bp.CreateVirtualNetwork(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating virtual network", err.Error())
		return
	}

	// Update the plan with the received ObjectId and set the partial state
	plan.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	// Fetch the virtual network to learn apstra-assigned VLAN assignments
	var api *apstra.VirtualNetwork
	retryMax := 25
	for {
		if retryMax == 0 {
			break
		}
		retryMax--
		api, err = bp.GetVirtualNetwork(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error fetching just-created virtual network %q", id),
				err.Error())
			return
		}

		if plan.IPv4ConnectivityEnabled.ValueBool() && api.Data.Ipv4Subnet == nil {
			time.Sleep(200 * time.Millisecond)
			continue // try again
		}
		if plan.IPv6ConnectivityEnabled.ValueBool() && api.Data.Ipv6Subnet == nil {
			time.Sleep(200 * time.Millisecond)
			continue // try again
		}

		break
	}

	// Create a new state object and load the current state from the API
	var state DatacenterVirtualNetworkResourceModel
	state.Id = types.StringValue(id.String())
	state.BlueprintId = plan.BlueprintId
	state.HadPriorVniConfig = plan.HadPriorVniConfig
	state.DatacenterVirtualNetwork.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	
	// Load SviIps from API response
	state.SviIps = blueprint.LoadApiSviIps(ctx, api.Data.SviIps, &resp.Diagnostics)

	// Don't rely on the API response for these values (#170). If the config
	// supplied a value, use it when setting state.
	if !plan.IPv4Subnet.IsUnknown() {
		state.IPv4Subnet = plan.IPv4Subnet
	}
	if !plan.IPv6Subnet.IsUnknown() {
		state.IPv6Subnet = plan.IPv6Subnet
	}
	if !plan.IPv4Gateway.IsUnknown() {
		state.IPv4Gateway = plan.IPv4Gateway
	}
	if !plan.IPv6Gateway.IsUnknown() {
		state.IPv6Gateway = plan.IPv6Gateway
	}
	if !plan.Vni.IsUnknown() {
		state.Vni = plan.Vni
	}
	if !plan.ReserveVlan.IsUnknown() {
		state.ReserveVlan = plan.ReserveVlan
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update Read method to handle SviIps
func (o *resourceDatacenterVirtualNetwork) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state DatacenterVirtualNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Retrieve the virtual network
	vn, err := bp.GetVirtualNetwork(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
	}

	// Record whether we're expecting to find VN bindings
	bindingsShouldBeNull := state.Bindings.IsNull()

	// Load the API response
	state.DatacenterVirtualNetwork.LoadApiData(ctx, vn.Data, &resp.Diagnostics)
	
	// Load SviIps from API response
	state.SviIps = blueprint.LoadApiSviIps(ctx, vn.Data.SviIps, &resp.Diagnostics)

	// Wipe out the bindings if none were recorded in the state (bindings created some other way)
	if bindingsShouldBeNull && !state.Bindings.IsNull() {
		state.Bindings = types.MapNull(types.ObjectType{AttrTypes: blueprint.VnBinding{}.AttrTypes()})
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update Update method to handle SviIps
func (o *resourceDatacenterVirtualNetwork) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan DatacenterVirtualNetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Create a request object with SviIps
	request := plan.DatacenterVirtualNetwork.RequestWithSviIps(ctx, plan.SviIps, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the virtual network according to the plan
	err = bp.UpdateVirtualNetwork(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating virtual network", err.Error())
	}

	// Fetch the virtual network to learn apstra-assigned VLAN assignments
	api, err := bp.GetVirtualNetwork(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error fetching just-updated virtual network %q", plan.Id.ValueString()),
			err.Error())
		return
	}

	// Create a new state object and load the current state from the API
	var state DatacenterVirtualNetworkResourceModel
	state.Id = plan.Id
	state.BlueprintId = plan.BlueprintId
	state.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())
	state.DatacenterVirtualNetwork.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	
	// Load SviIps from API response
	state.SviIps = blueprint.LoadApiSviIps(ctx, api.Data.SviIps, &resp.Diagnostics)

	// Don't rely on the API response for these values (#170). If the config
	// supplied a value, use that when setting state.
	if !plan.IPv4Subnet.IsUnknown() {
		state.IPv4Subnet = plan.IPv4Subnet
	}
	if !plan.IPv6Subnet.IsUnknown() {
		state.IPv6Subnet = plan.IPv6Subnet
	}
	if !plan.IPv4Gateway.IsUnknown() {
		state.IPv4Gateway = plan.IPv4Gateway
	}
	if !plan.IPv6Gateway.IsUnknown() {
		state.IPv6Gateway = plan.IPv6Gateway
	}
	if !plan.ReserveVlan.IsUnknown() {
		state.ReserveVlan = plan.ReserveVlan
	}

	// If the plan modifier didn't take action...
	if plan.HadPriorVniConfig.IsUnknown() {
		// ...then the trigger value is set according to whether a VNI value is known.
		state.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())
	} else {
		state.HadPriorVniConfig = plan.HadPriorVniConfig
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}