package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceConnectivityTemplateAssignment{}

type resourceConnectivityTemplateAssignment struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceConnectivityTemplateAssignment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_connectivity_template_assignment"
}

func (o *resourceConnectivityTemplateAssignment) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceConnectivityTemplateAssignment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource assigns a Connectivity Template to an Application Point (probably an " +
			"interface at the fabric edge, possibly a switch node in the fabric) within a Blueprint.",
		Attributes: blueprint.ConnectivityTemplateAssignment{}.ResourceAttributes(),
	}
}

func (o *resourceConnectivityTemplateAssignment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.ConnectivityTemplateAssignment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client for the datacenter reference design
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
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

	addIds, _ := plan.AddDelRequest(ctx, nil, &resp.Diagnostics)
	err = bpClient.SetInterfaceConnectivityTemplates(ctx, apstra.ObjectId(plan.ConnectivityTemplateId.ValueString()), addIds)
	if err != nil {
		resp.Diagnostics.AddError("failed applying Connectivity Template", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	return
}

func (o *resourceConnectivityTemplateAssignment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.ConnectivityTemplateAssignment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client for the datacenter reference design
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, state.BlueprintId), err.Error())
		return
	}

	// currentCtIds come from the API, may include CTs unrelated to this resource
	currentCtIds, err := bpClient.GetInterfaceConnectivityTemplates(ctx, apstra.ObjectId(state.ConnectivityTemplateId.String()))
	if err != nil {
		resp.Diagnostics.AddError("failed reading Connectivity Template assignments", err.Error())
		return
	}

	// stateCtIds come from the history of this resource. What CTs have been previously assigned?
	var stateCtIds []apstra.ObjectId
	resp.Diagnostics.Append(state.ApplicationPointIds.ElementsAs(ctx, &stateCtIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// remainingCtIds are the previously assigned IDs (state) which are still assigned (current)
	remainingCtIds := utils.SliceIntersectionOfAB(currentCtIds, stateCtIds)
	var d diag.Diagnostics
	state.ApplicationPointIds, d = types.SetValueFrom(ctx, types.StringType, remainingCtIds)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceConnectivityTemplateAssignment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.ConnectivityTemplateAssignment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state blueprint.ConnectivityTemplateAssignment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	state.DeployMode = plan.DeployMode
	state.SetNodeDeployMode(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DeviceKey.Equal(state.DeviceKey) {
		// device key has changed
		state.DeviceKey = plan.DeviceKey // copy user input directly from plan
		state.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceConnectivityTemplateAssignment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.ConnectivityTemplateAssignment
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

	state.InitialInterfaceMapId = types.StringNull()
	state.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)

	state.DeviceKey = types.StringNull() // 'null' triggers clearing the 'system_id' field.
	state.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
}
