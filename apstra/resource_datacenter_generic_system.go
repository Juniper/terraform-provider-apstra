package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterGenericSystem{}

type resourceDatacenterGenericSystem struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceDatacenterGenericSystem) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_generic_system"
}

func (o *resourceDatacenterGenericSystem) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDatacenterGenericSystem) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Generic System within a Datacenter Blueprint.",
		Attributes:          blueprint.DatacenterGenericSystem{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterGenericSystem) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan.
	var plan blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client for the datacenter reference design
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, plan.BlueprintId), err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// prep a generic system creation request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// unfortunately we only learn the link IDs, not the generic system ID
	linkIds, err := bp.CreateLinksWithNewServer(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating generic system", err.Error())
		return
	}

	// use link IDs to learn the generic system ID
	genericSystemId, err := bp.SystemNodeFromLinkIds(ctx, linkIds, apstra.SystemNodeRoleGeneric)
	if err != nil {
		resp.Diagnostics.AddError("failed to determine new generic system ID from links", err.Error())
	}
	plan.Id = types.StringValue(genericSystemId.String())

	// populate apstra-assigned link info into the correct terraform structure
	plan.PopulateLinkInfo(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//plan.PopulateGenericSystemInfo(ctx, bp, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a blueprint client
	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
		return
	}

	var node struct {
		Label    string `json:"label"`
		Hostname string `json:"hostname"`
	}
	err = bp.Client().GetNode(ctx, bp.Id(), apstra.ObjectId(state.Id.ValueString()), &node)
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
	}
	state.Label = types.StringValue(node.Label)
	state.Hostname = types.StringValue(node.Hostname)
	state.ReadLogicalDevice(ctx, bp, &resp.Diagnostics)
	state.ReadTags(ctx, bp, &resp.Diagnostics)
	//state.ReadLinks(ctx, bp, &resp.Diagnostics)

	//tags := blueprint.NodeTags(ctx, state.Id.ValueString(), bp, &resp.Diagnostics)

	//state.Tags = utils.SetValueOrNull(ctx, types.StringType, blueprint.NodeTags(ctx, state.Id.ValueString(), bp, &resp.Diagnostics))

	//sz, err := bp.GetSecurityZone(ctx, apstra.ObjectId(state.Id.ValueString()))
	//if err != nil {
	//	var ace apstra.ApstraClientErr
	//	if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
	//		resp.State.RemoveResource(ctx)
	//		return
	//	}
	//	resp.Diagnostics.AddError("error retrieving security zone", err.Error())
	//	return
	//}
	//
	//state.LoadApiData(ctx, sz.Data, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//dhcpServers, err := bp.GetSecurityZoneDhcpServers(ctx, sz.Id)
	//if err != nil {
	//	var ace apstra.ApstraClientErr
	//	if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
	//		resp.State.RemoveResource(ctx)
	//		return
	//	}
	//	resp.Diagnostics.AddError("error retrieving security zone", err.Error())
	//	return
	//}
	//
	//state.LoadApiDhcpServers(ctx, dhcpServers, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// set state
	//resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterGenericSystem) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//if o.client == nil {
	//	resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
	//	return
	//}
	//
	//// Retrieve values from plan.
	//var plan blueprint.DatacenterGenericSystem
	//resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// Create a blueprint client
	//bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	//if err != nil {
	//	resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, plan.BlueprintId), err.Error())
	//	return
	//}
	//
	//// Lock the blueprint mutex.
	//err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
	//		err.Error())
	//	return
	//}
	//
	//request := plan.Request(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//err = bp.UpdateSecurityZone(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	//if err != nil {
	//	resp.Diagnostics.AddError("error updating security zone", err.Error())
	//	return
	//}
	//
	//dhcpRequest := plan.DhcpServerRequest(ctx, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//err = bp.SetSecurityZoneDhcpServers(ctx, apstra.ObjectId(plan.Id.ValueString()), dhcpRequest)
	//if err != nil {
	//	resp.Diagnostics.AddError("error updating security zone dhcp servers", err.Error())
	//	return
	//}
	//
	//sz, err := bp.GetSecurityZone(ctx, apstra.ObjectId(plan.Id.ValueString()))
	//if err != nil {
	//	resp.Diagnostics.AddError("error retrieving just-created security zone", err.Error())
	//}
	//
	//// if the plan modifier didn't take action...
	//if plan.HadPriorVlanIdConfig.IsUnknown() {
	//	// ...then the trigger value is set according to whether a VLAN ID value is known.
	//	plan.HadPriorVlanIdConfig = types.BoolValue(!plan.VlanId.IsUnknown())
	//}
	//
	//// if the plan modifier didn't take action...
	//if plan.HadPriorVniConfig.IsUnknown() {
	//	// ...then the trigger value is set according to whether a VNI value is known.
	//	plan.HadPriorVniConfig = types.BoolValue(!plan.Vni.IsUnknown())
	//}
	//
	//plan.LoadApiData(ctx, sz.Data, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No need to proceed if the blueprint no longer exists
	if !utils.BlueprintExists(ctx, o.client, apstra.ObjectId(state.BlueprintId.ValueString()), &resp.Diagnostics) {
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
		return
	}

	err = bp.DeleteGenericSystem(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != apstra.ErrNotfound {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting generic system", err.Error())
	}
}
