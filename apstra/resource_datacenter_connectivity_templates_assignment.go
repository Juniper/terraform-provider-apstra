package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterConnectivityTemplatesAssignment{}
var _ resourceWithSetDcBpClientFunc = &resourceDatacenterConnectivityTemplatesAssignment{}
var _ resourceWithSetBpLockFunc = &resourceDatacenterConnectivityTemplatesAssignment{}

type resourceDatacenterConnectivityTemplatesAssignment struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_connectivity_templates_assignment"
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource assigns one or more Connectivity Templates to an " +
			"Application Point. Application Points are graph nodes including interfaces at the " +
			"fabric edge, and switches within the fabric.",
		Attributes: blueprint.ConnectivityTemplatesAssignment{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.ConnectivityTemplatesAssignment
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
			fmt.Sprintf("failed locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	addIds, _ := plan.AddDelRequest(ctx, nil, &resp.Diagnostics)
	err = bp.SetApplicationPointConnectivityTemplates(ctx, apstra.ObjectId(plan.ApplicationPointId.ValueString()), addIds)
	if err != nil {
		resp.Diagnostics.AddError("failed applying Connectivity Template", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.ConnectivityTemplatesAssignment
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

	// currentCtIds come from the API, may include CTs unrelated to this resource
	currentCtIds, err := bp.GetInterfaceConnectivityTemplates(ctx, apstra.ObjectId(state.ApplicationPointId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed reading Connectivity Template assignments", err.Error())
		return
	}

	// stateCtIds come from the history of this resource. What CTs have been previously assigned?
	var stateCtIds []apstra.ObjectId
	resp.Diagnostics.Append(state.ConnectivityTemplateIds.ElementsAs(ctx, &stateCtIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// remainingCtIds are the previously assigned IDs (state) which are still assigned (current)
	remainingCtIds := utils.SliceIntersectionOfAB(currentCtIds, stateCtIds)
	state.ConnectivityTemplateIds = utils.SetValueOrNull(ctx, types.StringType, remainingCtIds, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.ConnectivityTemplatesAssignment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state blueprint.ConnectivityTemplatesAssignment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// calculate the add/del sets
	addIds, delIds := plan.AddDelRequest(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// add any required CTs
	err = bp.SetApplicationPointConnectivityTemplates(ctx, apstra.ObjectId(plan.ApplicationPointId.ValueString()), addIds)
	if err != nil {
		resp.Diagnostics.AddError("failed assigning connectivity templates", err.Error())
		return
	}

	// clear any undesired CTs
	err = bp.DelApplicationPointConnectivityTemplates(ctx, apstra.ObjectId(plan.ApplicationPointId.ValueString()), delIds)
	if err != nil {
		resp.Diagnostics.AddError("failed clearing connectivity template assignments", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.ConnectivityTemplatesAssignment
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
			fmt.Sprintf("failed locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	var delIds []apstra.ObjectId
	resp.Diagnostics.Append(state.ConnectivityTemplateIds.ElementsAs(ctx, &delIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.DelApplicationPointConnectivityTemplates(ctx, apstra.ObjectId(state.ApplicationPointId.ValueString()), delIds)
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed clearing connectivity template assignments", err.Error())
		return
	}
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterConnectivityTemplatesAssignment) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
