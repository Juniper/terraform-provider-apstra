package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resourceWithSetBpClientFunc = &resourceDatacenterSubinterfaceAddressing{}
	_ resourceWithSetBpLockFunc   = &resourceDatacenterSubinterfaceAddressing{}
)

type resourceDatacenterSubinterfaceAddressing struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterSubinterfaceAddressing) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_subinterface_addressing"
}

func (o *resourceDatacenterSubinterfaceAddressing) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterSubinterfaceAddressing) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates IPv4 and IPv6 addressing on L3 " +
			"subinterfaces within a Datacenter Blueprint fabric. It is intended for use with subinterfaces created " +
			"as a side-effect of assigning Connectivity Templates containing IP Link primitives.",
		Attributes: blueprint.SubinterfaceAddressing{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterSubinterfaceAddressing) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blueprint.SubinterfaceAddressing
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
			fmt.Sprintf("failed to lock blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// create a subinterface addressing request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the subinterface
	err = bp.UpdateSubinterfaces(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to add subinterface addressing", err.Error())
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterSubinterfaceAddressing) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blueprint.SubinterfaceAddressing
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Blueprint %s not found", state.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("Failed to create Blueprint client", err.Error())
		return
	}

	// fetch the details from the API
	apiData, err := bp.GetSubinterface(ctx, apstra.ObjectId(state.SubinterfaceId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch subinterface node details", err.Error())
		return
	}

	// load the state
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterSubinterfaceAddressing) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan blueprint.SubinterfaceAddressing
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
			fmt.Sprintf("failed to lock blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// create a subinterface addressing request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the subinterface
	err = bp.UpdateSubinterfaces(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to add subinterface addressing", err.Error())
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterSubinterfaceAddressing) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blueprint.SubinterfaceAddressing
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
			fmt.Sprintf("failed to lock blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// create a subinterface addressing request which kills off IPv4 and IPv6 addressing
	request := map[apstra.ObjectId]apstra.TwoStageL3ClosSubinterface{
		apstra.ObjectId(state.SubinterfaceId.ValueString()): {
			Ipv4AddrType: apstra.InterfaceNumberingIpv4TypeNone,
			Ipv6AddrType: apstra.InterfaceNumberingIpv6TypeNone,
		},
	}

	// update the subinterface
	err = bp.UpdateSubinterfaces(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to clear subinterface addressing", err.Error())
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterSubinterfaceAddressing) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterSubinterfaceAddressing) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
