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
	sysredundancyinfo "github.com/Juniper/terraform-provider-apstra/internal/system_redundancy_info"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &resourceDatacenterVirtualNetworkAssignment{}
	_ resource.ResourceWithImportState = &resourceDatacenterVirtualNetworkAssignment{}
	_ resource.ResourceWithIdentity    = &resourceDatacenterVirtualNetworkAssignment{}
	_ resourceWithSetDcBpClientFunc    = &resourceDatacenterVirtualNetworkAssignment{}
	_ resourceWithSetBpLockFunc        = &resourceDatacenterVirtualNetworkAssignment{}
)

type resourceDatacenterVirtualNetworkAssignment struct {
	lockFunc        func(context.Context, string) error
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (r *resourceDatacenterVirtualNetworkAssignment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network_assigment"
}

func (r *resourceDatacenterVirtualNetworkAssignment) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, r, req, resp)
}

func (r *resourceDatacenterVirtualNetworkAssignment) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: blueprint.VirtualNetworkAssignmentIdentity{}.Attributes(),
	}
}

func (r *resourceDatacenterVirtualNetworkAssignment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource assigns a Virtual Network to a Leaf Switch or a Leaf Switch Redundancy Group within a Blueprint.",
		Attributes:          blueprint.InterconnectDomainL3Policy{}.ResourceAttributes(),
	}
}

func (r *resourceDatacenterVirtualNetworkAssignment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	hasIdentity := req.Identity != nil
	hasID := req.ID != ""

	var identity blueprint.VirtualNetworkAssignmentIdentity

	switch {
	case hasID && hasIdentity: // Why do we have both?
		resp.Diagnostics.AddError(
			"Ambiguous import input",
			"Provide either legacy import ID or identity, not both.",
		)
	case hasIdentity: // Unpack the provided identity into our identity struct.
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	case hasID: // Our identity struct will parse the legacy import ID string into its fields.
		identity.ParseLegacyID(ctx, req.ID, &resp.Diagnostics)
	default: // Neither identity nor legacy import ID string provided.
		resp.Diagnostics.AddError(
			"Missing import input",
			"Provide either a legacy import ID string or an identity object.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design.
	bp, err := r.getBpClientFunc(ctx, identity.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, identity.BlueprintID), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, identity.BlueprintID), err.Error())
		return
	}

	// Fetch the state from the API.
	state := identity.GetState(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the identity and state.
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDatacenterVirtualNetworkAssignment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.VirtualNetworkAssignment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design.
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

	plan.LeafRGID = types.StringPointerValue(sysredundancyinfo.LookupRG(ctx, bp, plan.LeafID.ValueString(), &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.FetchLeafRedundancyGroupID(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.Request(ctx, nil, &resp.Diagnostics)
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
		resp.Diagnostics.AddError(fmt.Sprintf("error locking blueprint %s mutex", state.BlueprintID), err.Error())
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
		resp.Diagnostics.AddError(ierrors.DeleteError(r), err.Error())
	}
}

func (r *resourceDatacenterInterconnectDomainL3Policy) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	r.getBpClientFunc = f
}

func (r *resourceDatacenterInterconnectDomainL3Policy) setBpLockFunc(f func(context.Context, string) error) {
	r.lockFunc = f
}
