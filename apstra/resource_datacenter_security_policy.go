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

var _ resource.ResourceWithConfigure = &resourceDatacenterSecurityPolicy{}
var _ resource.ResourceWithImportState = &resourceDatacenterSecurityPolicy{}

type resourceDatacenterSecurityPolicy struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterSecurityPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_security_policy"
}

func (o *resourceDatacenterSecurityPolicy) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.getBpClientFunc = ResourceGetTwoStageL3ClosClientFunc(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDatacenterSecurityPolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a Security Policy within a Datacenter Blueprint.",
		Attributes:          blueprint.DatacenterSecurityPolicy{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterSecurityPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	state := blueprint.DatacenterSecurityPolicy{
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
				"External Gateway not found",
				fmt.Sprintf("Blueprint %q Security Policy with ID %s not found", bp.Id(), state.Id))
			return
		}
		resp.Diagnostics.AddError("Failed to read Security Policy", err.Error())
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterSecurityPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterSecurityPolicy
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

	// create a request object
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the security policy
	id, err := bp.CreatePolicy(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating security policy", err.Error())
		return
	}

	// set the ID into the plan object and then read it back from the API to get the rule IDs
	plan.Id = types.StringValue(id.String())
	err = plan.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("failed reading just-created Security Policy", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (o *resourceDatacenterSecurityPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterSecurityPolicy
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

	// read the state
	err = state.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed reading Security Policy", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterSecurityPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterSecurityPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// create a request object
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the security policy
	err = bp.UpdatePolicy(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error creating security policy", err.Error())
		return
	}

	// read the security policy back from the API to get the rule IDs
	err = plan.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("failed reading just-updated Security Policy", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (o *resourceDatacenterSecurityPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterSecurityPolicy
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

	// Delete the security policy
	err = bp.DeletePolicy(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting security policy", err.Error())
	}
}
