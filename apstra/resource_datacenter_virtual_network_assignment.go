package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	ierrors "github.com/Juniper/terraform-provider-apstra/internal/errors"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
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
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network_assignment"
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
		MarkdownDescription: docCategoryDatacenter + "This resource assigns a Virtual Network to a Leaf Switch and Access Switches within a *Datacenter* Blueprint.",
		Attributes:          blueprint.VirtualNetworkAssignment{}.ResourceAttributes(),
	}
}

func (r *resourceDatacenterVirtualNetworkAssignment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var identity blueprint.VirtualNetworkAssignmentIdentity

	switch {
	case req.Identity != nil: // Unpack the provided identity into our identity struct.
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	case req.ID != "": // Our identity struct will parse the legacy import ID string into its fields.
		identity.ParseLegacyID(ctx, req.ID, &resp.Diagnostics)
	default: // Neither identity nor legacy import ID string provided.
		resp.Diagnostics.AddError(
			"Missing import input",
			"Provide either a legacy import ID string (JSON format) or an identity object.",
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
	// Retrieve values from plan, and set the identity.
	var plan blueprint.VirtualNetworkAssignment
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.Append(resp.Identity.Set(ctx, plan.Identity())...); resp.Diagnostics.HasError() {
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

	// Create a request object.
	request := plan.Request(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the VN bindings.
	err = bp.UpdateVirtualNetworkLeafBindings(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(ierrors.CreateError(r), err.Error())
		return
	}

	// Set private state.
	private.ResourceDatacenterVirtualNetworkAssignment{
		IPv4ModeNull:    plan.IPv4Mode.IsNull(),
		IPv4AddressNull: plan.IPv4Address.IsNull(),
		IPv6ModeNull:    plan.IPv6Mode.IsNull(),
		IPv6AddressNull: plan.IPv6Address.IsNull(),
	}.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)

	// Set the state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDatacenterVirtualNetworkAssignment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state and set the identity.
	var state blueprint.VirtualNetworkAssignment
	if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.Append(resp.Identity.Set(ctx, state.Identity())...); resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design.
	bp, err := r.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.Append(resp.Identity.Set(ctx, state.Identity())...)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintID), err.Error())
		return
	}

	ok := state.Read(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return // Order here matters: If we have an error, we don't want to remove the resource from state.
	}
	if !ok {
		resp.State.RemoveResource(ctx) // Remove the resource because we failed to find our assignment while reading the API *without error*.
		return
	}

	// Reset non-computed attributes for which the user has not expressed an opinion.
	privateState := new(private.ResourceDatacenterVirtualNetworkAssignment)
	privateState.LoadPrivateState(ctx, resp.Private, &resp.Diagnostics)
	if privateState.IPv4ModeNull {
		state.IPv4Mode = types.StringNull()
	}
	if privateState.IPv4AddressNull {
		state.IPv4Address = cidrtypes.NewIPv4PrefixNull()
	}
	if privateState.IPv6ModeNull {
		state.IPv6Mode = types.StringNull()
	}
	if privateState.IPv6AddressNull {
		state.IPv6Address = cidrtypes.NewIPv6PrefixNull()
	}

	// Set the state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDatacenterVirtualNetworkAssignment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan and set the identity.
	var plan blueprint.VirtualNetworkAssignment
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.Append(resp.Identity.Set(ctx, plan.Identity())...); resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design.
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

	// Create a request object.
	request := plan.Request(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the VN bindings.
	err = bp.UpdateVirtualNetworkLeafBindings(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(ierrors.UpdateError(r), err.Error())
		return
	}

	// Set private state.
	private.ResourceDatacenterVirtualNetworkAssignment{
		IPv4ModeNull:    plan.IPv4Mode.IsNull(),
		IPv4AddressNull: plan.IPv4Address.IsNull(),
		IPv6ModeNull:    plan.IPv6Mode.IsNull(),
		IPv6AddressNull: plan.IPv6Address.IsNull(),
	}.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)

	// Set the state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDatacenterVirtualNetworkAssignment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.VirtualNetworkAssignment
	if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}

	// Get a client for the datacenter reference design.
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

	// In order to signal removal of the binding, we need to know if the leaf switch is a standalone device or a member of a redundancy group.
	var vnBindings map[string]*datacenter.VNBinding
	groupID, peerID, err := bp.GetRedundancyGroupBySystem(ctx, state.LeafID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed while determining group membership", err.Error())
	}

	if peerID == nil {
		// The leaf ID represents a standalone switch. Signal removal to the API by sending a nil binding for the leaf ID.
		vnBindings = map[string]*datacenter.VNBinding{ // Leef ID represents a standalone switch.
			state.LeafID.ValueString(): nil,
		}
	} else {
		// The leaf ID represents a member of a redundancy group. Send a nil binding for the redundancy group ID.
		vnBindings = map[string]*datacenter.VNBinding{ // Leaf ID represents a switch that is part of a redundant pair.
			*groupID: nil,
		}
	}

	// Create the request to remove the binding.
	request := apstra.VirtualNetworkBindingsRequest{
		VirtualNetworkID: state.VNID.ValueString(),
		Bindings:         vnBindings,
	}

	// Update the VN bindings.
	err = bp.UpdateVirtualNetworkLeafBindings(ctx, request)
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(ierrors.DeleteError(r), err.Error())
	}
}

func (r *resourceDatacenterVirtualNetworkAssignment) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	r.getBpClientFunc = f
}

func (r *resourceDatacenterVirtualNetworkAssignment) setBpLockFunc(f func(context.Context, string) error) {
	r.lockFunc = f
}
