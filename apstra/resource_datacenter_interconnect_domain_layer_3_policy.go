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
	_ resource.ResourceWithConfigure   = &resourceDatacenterInterconnectDomainL3Policy{}
	_ resource.ResourceWithImportState = &resourceDatacenterInterconnectDomainL3Policy{}
	_ resourceWithSetDcBpClientFunc    = &resourceDatacenterInterconnectDomainL3Policy{}
	_ resourceWithSetBpLockFunc        = &resourceDatacenterInterconnectDomainL3Policy{}
)

type resourceDatacenterInterconnectDomainL3Policy struct {
	lockFunc        func(context.Context, string) error
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (r *resourceDatacenterInterconnectDomainL3Policy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain_layer_3_policy"
}

func (r *resourceDatacenterInterconnectDomainL3Policy) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, r, req, resp)
}

func (r *resourceDatacenterInterconnectDomainL3Policy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource configures an Interconnect Domain Layer 3 Policy within a Blueprint.",
		Attributes:          blueprint.InterconnectDomainL3Policy{}.ResourceAttributes(),
	}
}

func (r *resourceDatacenterInterconnectDomainL3Policy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var importID struct {
		BlueprintID          string `json:"blueprint_id"`
		InterconnectDomainID string `json:"interconnect_domain_id"`
		RoutingZoneID        string `json:"routing_zone_id"`
	}

	// parse the user-supplied import ID string JSON
	err := json.Unmarshal([]byte(req.ID), &importID)
	if err != nil {
		resp.Diagnostics.AddError("failed parsing import ID JSON string", err.Error())
		return
	}

	if importID.BlueprintID == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'blueprint_id' element of import ID string cannot be empty")
		return
	}

	if importID.InterconnectDomainID == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'interconnect_domain_id' element of import ID string cannot be empty")
		return
	}

	if importID.RoutingZoneID == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'routing_zone_id' element of import ID string cannot be empty")
		return
	}

	// create a state object preloaded with the critical details we need in advance
	state := blueprint.InterconnectDomainL3Policy{
		BlueprintID:          types.StringValue(importID.BlueprintID),
		InterconnectDomainID: types.StringValue(importID.InterconnectDomainID),
		RoutingZoneID:        types.StringValue(importID.RoutingZoneID),
	}

	// get a client for the datacenter reference design
	bp, err := r.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, state.BlueprintID), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintID), err.Error())
		return
	}

	state.Read(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDatacenterInterconnectDomainL3Policy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.InterconnectDomainL3Policy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := r.getBpClientFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, plan.BlueprintID), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, plan.BlueprintID), err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = r.lockFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %s mutex", plan.BlueprintID),
			err.Error())
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.UpdateEVPNInterconnectGroup(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error setting Interconnect Domain Layer 3 Policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDatacenterInterconnectDomainL3Policy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.InterconnectDomainL3Policy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := r.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintID), err.Error())
		return
	}

	state.Read(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDatacenterInterconnectDomainL3Policy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.InterconnectDomainL3Policy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := r.getBpClientFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, plan.BlueprintID), err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = r.lockFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %s mutex", plan.BlueprintID),
			err.Error())
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.UpdateEVPNInterconnectGroup(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Interconnect Domain Layer 3 Policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDatacenterInterconnectDomainL3Policy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.InterconnectDomainL3Policy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := r.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintID), err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = r.lockFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %s mutex", state.BlueprintID),
			err.Error())
		return
	}

	// Clear the DCI L3 Policy
	request := apstra.EVPNInterconnectGroup{
		InterconnectSecurityZones: map[string]apstra.InterconnectSecurityZone{
			state.RoutingZoneID.ValueString(): {L3Enabled: false},
		},
	}
	_ = request.SetID(state.InterconnectDomainID.ValueString())
	err = bp.UpdateEVPNInterconnectGroup(ctx, request)
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Interconnect Domain Layer 3 Policy", err.Error())
	}
}

func (r *resourceDatacenterInterconnectDomainL3Policy) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	r.getBpClientFunc = f
}

func (r *resourceDatacenterInterconnectDomainL3Policy) setBpLockFunc(f func(context.Context, string) error) {
	r.lockFunc = f
}
