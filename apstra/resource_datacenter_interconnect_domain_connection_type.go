package tfapstra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	ierrors "github.com/Juniper/terraform-provider-apstra/internal/errors"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &resourceDatacenterInterconnectDomainConnectionType{}
	_ resource.ResourceWithImportState = &resourceDatacenterInterconnectDomainConnectionType{}
	_ resourceWithSetDcBpClientFunc    = &resourceDatacenterInterconnectDomainConnectionType{}
	_ resourceWithSetBpLockFunc        = &resourceDatacenterInterconnectDomainConnectionType{}
)

type resourceDatacenterInterconnectDomainConnectionType struct {
	lockFunc        func(context.Context, string) error
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (r *resourceDatacenterInterconnectDomainConnectionType) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain_connection_type"
}

func (r *resourceDatacenterInterconnectDomainConnectionType) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, r, req, resp)
}

func (r *resourceDatacenterInterconnectDomainConnectionType) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource configures per-VN DCI details within a Blueprint.",
		Attributes:          blueprint.InterconnectDomainConnectionType{}.ResourceAttributes(),
	}
}

func (r *resourceDatacenterInterconnectDomainConnectionType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var importID struct {
		BlueprintID          string `json:"blueprint_id"`
		InterconnectDomainID string `json:"interconnect_domain_id"`
		VirtualNetworkID     string `json:"virtual_network_id"`
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

	if importID.VirtualNetworkID == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'virtual_network_id' element of import ID string cannot be empty")
		return
	}

	// create a state object preloaded with the critical details we need in advance
	state := blueprint.InterconnectDomainConnectionType{
		BlueprintID:          types.StringValue(importID.BlueprintID),
		InterconnectDomainID: types.StringValue(importID.InterconnectDomainID),
		VirtualNetworkID:     types.StringValue(importID.VirtualNetworkID),
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

	err = state.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(ierrors.ImportError(r), err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDatacenterInterconnectDomainConnectionType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.InterconnectDomainConnectionType
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
		resp.Diagnostics.AddError(fmt.Sprintf("error locking blueprint %s mutex", plan.BlueprintID), err.Error())
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.UpdateEVPNInterconnectGroup(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(ierrors.CreateError(r), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDatacenterInterconnectDomainConnectionType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.InterconnectDomainConnectionType
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

	err = state.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		if errors.As(err, new(ierrors.ResourceNotFound)) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(ierrors.ReadError(r), err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDatacenterInterconnectDomainConnectionType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.InterconnectDomainConnectionType
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
		resp.Diagnostics.AddError(fmt.Sprintf("error locking blueprint %s mutex", plan.BlueprintID), err.Error())
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = bp.UpdateEVPNInterconnectGroup(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(ierrors.UpdateError(r), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDatacenterInterconnectDomainConnectionType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.InterconnectDomainConnectionType
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
		resp.Diagnostics.AddError(fmt.Sprintf("error locking blueprint %s mutex", state.BlueprintID), err.Error())
		return
	}

	// Clear the DCI Connection Type settings by disabling L2 and L3.
	request := apstra.EVPNInterconnectGroup{
		InterconnectVirtualNetworks: map[string]apstra.InterconnectVirtualNetwork{
			state.VirtualNetworkID.ValueString(): {L2Enabled: false, L3Enabled: false},
		},
	}
	_ = request.SetID(state.InterconnectDomainID.ValueString())
	err = bp.UpdateEVPNInterconnectGroup(ctx, request)
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(ierrors.DeleteError(r), err.Error())
	}
}

func (r *resourceDatacenterInterconnectDomainConnectionType) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	r.getBpClientFunc = f
}

func (r *resourceDatacenterInterconnectDomainConnectionType) setBpLockFunc(f func(context.Context, string) error) {
	r.lockFunc = f
}
