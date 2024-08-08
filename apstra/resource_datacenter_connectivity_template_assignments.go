package tfapstra

import (
	"context"
	"errors"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceDatacenterConnectivityTemplateAssignments{}
	_ resourceWithSetDcBpClientFunc  = &resourceDatacenterConnectivityTemplateAssignments{}
	_ resourceWithSetBpLockFunc      = &resourceDatacenterConnectivityTemplateAssignments{}
)

type resourceDatacenterConnectivityTemplateAssignments struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterConnectivityTemplateAssignments) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_connectivity_template_assignments"
}

func (o *resourceDatacenterConnectivityTemplateAssignments) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterConnectivityTemplateAssignments) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource assigns a Connectivity Template to one or more " +
			"Application Points. Application Points are graph nodes including interfaces at the " +
			"fabric edge, and switches within the fabric.",
		Attributes: blueprint.ConnectivityTemplateAssignments{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterConnectivityTemplateAssignments) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.ConnectivityTemplateAssignments
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

	request := plan.Request(ctx, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.SetApplicationPointsConnectivityTemplates(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed while assigning Connectivity Template %s to Application Points", plan.ConnectivityTemplateId),
			err.Error())
	}

	// Fetch IP link IDs
	plan.GetIpLinkIds(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConnectivityTemplateAssignments) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.ConnectivityTemplateAssignments
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

	apToCtMap, err := bp.GetApplicationPointsConnectivityTemplatesByCt(ctx, apstra.ObjectId(state.ConnectivityTemplateId.ValueString()))
	if err != nil {
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed while reading Application Point assignments for Connectivity Template %s", state.ConnectivityTemplateId),
			err.Error())
	}

	var apIds []attr.Value
	for apId, ctInfo := range apToCtMap {
		if ctInfo[apstra.ObjectId(state.ConnectivityTemplateId.ValueString())] {
			apIds = append(apIds, types.StringValue(apId.String()))
		}
	}

	// Load application point IDs
	state.ApplicationPointIds = types.SetValueMust(types.StringType, apIds)

	// Fetch IP link IDs
	state.GetIpLinkIds(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterConnectivityTemplateAssignments) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.ConnectivityTemplateAssignments
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state blueprint.ConnectivityTemplateAssignments
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

	request := plan.Request(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.SetApplicationPointsConnectivityTemplates(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed while assigning Connectivity Template %s to Application Points", plan.ConnectivityTemplateId),
			err.Error())
	}

	// Fetch IP link IDs
	plan.GetIpLinkIds(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConnectivityTemplateAssignments) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.ConnectivityTemplateAssignments
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

	// extract the application point IDs from state - we use these to calculate the deletion request
	var applicationPointIds []apstra.ObjectId
	resp.Diagnostics.Append(state.ApplicationPointIds.ElementsAs(ctx, &applicationPointIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use the application point IDs to generate a deletion request
	request := make(map[apstra.ObjectId]map[apstra.ObjectId]bool, len(applicationPointIds))
	for _, applicationPointId := range applicationPointIds {
		request[applicationPointId] = map[apstra.ObjectId]bool{apstra.ObjectId(state.ConnectivityTemplateId.ValueString()): false}
	}

	// send the request
	err = bp.SetApplicationPointsConnectivityTemplates(ctx, request)
	if err != nil {
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed clearing connectivity template from application points", err.Error())
		return
	}
}

func (o *resourceDatacenterConnectivityTemplateAssignments) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterConnectivityTemplateAssignments) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
