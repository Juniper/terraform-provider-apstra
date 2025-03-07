package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resource.ResourceWithValidateConfig = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resourceWithSetClient               = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resourceWithSetDcBpClientFunc       = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resourceWithSetBpLockFunc           = (*resourceDatacenterVirtualNetworkBindings)(nil)
)

type resourceDatacenterVirtualNetworkBindings struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterVirtualNetworkBindings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network_bindings"
}

func (o *resourceDatacenterVirtualNetworkBindings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterVirtualNetworkBindings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource assigns Leaf Switches to a Virtual Network and sets " +
			"their interface addresses. It is intended for use with Apstra 5.0 and later configurations where Virtual Networks " +
			"are not bound to switches as a part of VN creation.",
		Attributes: blueprint.VirtualNetworkBindings{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterVirtualNetworkBindings) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config.
	var config blueprint.VirtualNetworkBindings
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bindings []blueprint.VirtualNetworkBinding
	resp.Diagnostics.Append(config.Bindings.ElementsAs(ctx, &bindings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	leafIds := make(map[string]struct{})
	for i, binding := range bindings {
		if _, ok := leafIds[binding.LeafId.ValueString()]; ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("bindings").AtSetValue(config.Bindings.Elements()[i]),
				constants.ErrInvalidConfig,
				fmt.Sprintf("Leaf ID %s used more than once", binding.LeafId),
			)
		}
	}

	// check apstra version
	if o.client != nil && !compatibility.VnEmptyBindingsOk.Check(version.Must(version.NewVersion(o.client.ApiVersion()))) {
		resp.Diagnostics.AddError(
			"incompatible Apstra server version",
			fmt.Sprintf("this resource requires Apstra %s, have version %s", compatibility.VnEmptyBindingsOk, o.client.ApiVersion()),
		)
	}
}

func (o *resourceDatacenterVirtualNetworkBindings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.VirtualNetworkBindings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
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

	// Get redundancy group info for this blueprint
	rgiMap := getRgiMap(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a bindings request, using nil for prior state
	request := plan.Request(ctx, rgiMap, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the bindings using the SDK's Update method so that we do not wipe out any existing bindings
	err = bp.UpdateVirtualNetworkLeafBindings(ctx, *request)
	if err != nil {
		resp.Diagnostics.AddError("failed to set virtual network bindings", err.Error())
		return
	}

	plan.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)
	// do not return on error - we need to set the state below

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterVirtualNetworkBindings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.VirtualNetworkBindings
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	vn, err := bp.GetVirtualNetwork(ctx, apstra.ObjectId(state.VirtualNetworkId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed to read virtual network %q", state.VirtualNetworkId.ValueString()), err.Error())
		return
	}

	// Get redundancy group info for this blueprint
	rgiMap := getRgiMap(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.LoadApiData(ctx, vn.Data, rgiMap, resp.Private, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterVirtualNetworkBindings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.VirtualNetworkBindings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
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

	// Get redundancy group info for this blueprint
	rgiMap := getRgiMap(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a bindings request
	request := plan.Request(ctx, rgiMap, req.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the bindings using the SDK's Update method so that we do not wipe out any existing bindings
	err = bp.UpdateVirtualNetworkLeafBindings(ctx, *request)
	if err != nil {
		resp.Diagnostics.AddError("failed to set virtual network bindings", err.Error())
		return
	}

	plan.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)
	// do not return on error - we need to set the state below

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterVirtualNetworkBindings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.VirtualNetworkBindings
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Get redundancy group info for this blueprint
	rgiMap := getRgiMap(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create an empty plan using state values
	plan := blueprint.VirtualNetworkBindings{
		BlueprintId:      state.BlueprintId,
		VirtualNetworkId: state.VirtualNetworkId,
		Bindings:         types.SetNull(types.ObjectType{AttrTypes: blueprint.VirtualNetworkBinding{}.AttrTypes()}),
	}

	// create a bindings request (private enumerates work to be un-done)
	request := plan.Request(ctx, rgiMap, req.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.UpdateVirtualNetworkLeafBindings(ctx, *request)
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to set virtual network bindings", err.Error())
		return
	}
}

func (o *resourceDatacenterVirtualNetworkBindings) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterVirtualNetworkBindings) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceDatacenterVirtualNetworkBindings) setClient(client *apstra.Client) {
	o.client = client
}

func getRgiMap(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) map[string]*apstra.RedundancyGroupInfo {
	allRgiInfo, err := bp.GetAllRedundancyGroupInfo(ctx)
	if err != nil {
		diags.AddError("failed to read redundancy groups from blueprint", err.Error())
		return nil
	}

	result := make(map[string]*apstra.RedundancyGroupInfo, len(allRgiInfo)*3)
	for _, rgiInfo := range allRgiInfo {
		result[rgiInfo.Id.String()] = &rgiInfo
		result[rgiInfo.SystemIds[0].String()] = &rgiInfo
		result[rgiInfo.SystemIds[1].String()] = &rgiInfo
	}

	return result
}
