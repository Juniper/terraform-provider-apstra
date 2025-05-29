package tfapstra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &resourceDatacenterInterconnectDomainGateway{}
	_ resource.ResourceWithImportState = &resourceDatacenterInterconnectDomainGateway{}
	_ resourceWithSetDcBpClientFunc    = &resourceDatacenterInterconnectDomainGateway{}
	_ resourceWithSetBpLockFunc        = &resourceDatacenterInterconnectDomainGateway{}
)

type resourceDatacenterInterconnectDomainGateway struct {
	lockFunc        func(context.Context, string) error
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *resourceDatacenterInterconnectDomainGateway) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain_gateway"
}

func (o *resourceDatacenterInterconnectDomainGateway) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterInterconnectDomainGateway) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates an Interconnect Domain Gateway within a Blueprint.",
		Attributes:          blueprint.InterconnectDomainGateway{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterInterconnectDomainGateway) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var importId struct {
		BlueprintId string `json:"blueprint_id"`
		Id          string `json:"id"`
	}

	// parse the user-supplied import ID string JSON
	err := json.Unmarshal([]byte(req.ID), &importId)
	if err != nil {
		resp.Diagnostics.AddError("failed parsing import id JSON string", err.Error())
		return
	}

	if importId.BlueprintId == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'blueprint_id' element of import ID string cannot be empty")
		return
	}

	if importId.Id == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'id' element of import ID string cannot be empty")
		return
	}

	// create a state object preloaded with the critical details we need in advance
	state := blueprint.InterconnectDomainGateway{
		BlueprintId: types.StringValue(importId.BlueprintId),
		Id:          types.StringValue(importId.Id),
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, state.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintId), err.Error())
		return
	}

	err = state.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(
				"Interconnect Domain Gateway not found",
				fmt.Sprintf("Blueprint %q Interconnect Domain Gateway with ID %s not found", bp.Id(), state.Id))
			return
		}
		resp.Diagnostics.AddError("Failed to fetch Interconnect Domain Gateway", err.Error())
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterInterconnectDomainGateway) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.InterconnectDomainGateway
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, plan.BlueprintId), err.Error())
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

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := bp.CreateRemoteGateway(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Interconnect Domain gateway", err.Error())
		return
	}

	plan.Id = types.StringValue(id.String())
	err = plan.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch just created Interconnect Domain Gateway", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterInterconnectDomainGateway) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.InterconnectDomainGateway
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
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintId), err.Error())
		return
	}

	err = state.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to fetch Interconnect Domain Gateway", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterInterconnectDomainGateway) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.InterconnectDomainGateway
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, plan.BlueprintId), err.Error())
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

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.UpdateRemoteGateway(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating remote gateway", err.Error())
		return
	}

	err = plan.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch just updated Interconnect Domain Gateway", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterInterconnectDomainGateway) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.InterconnectDomainGateway
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
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintId), err.Error())
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

	// Delete the remote gateway
	err = bp.DeleteRemoteGateway(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting remote gateway", err.Error())
	}
}

func (o *resourceDatacenterInterconnectDomainGateway) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterInterconnectDomainGateway) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
