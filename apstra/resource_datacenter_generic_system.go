package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure  = &resourceDatacenterGenericSystem{}
	_ resource.ResourceWithModifyPlan = &resourceDatacenterGenericSystem{}
	_ resourceWithSetDcBpClientFunc   = &resourceDatacenterGenericSystem{}
	_ resourceWithSetBpLockFunc       = &resourceDatacenterGenericSystem{}
)

type resourceDatacenterGenericSystem struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterGenericSystem) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_generic_system"
}

func (o *resourceDatacenterGenericSystem) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterGenericSystem) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a Generic System within a Datacenter Blueprint.",
		Attributes:          blueprint.DatacenterGenericSystem{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterGenericSystem) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return // we must be about to call Delete()
	}

	if req.State.Raw.IsNull() {
		return // we must be about to call Create()
	}

	// extract plan and state
	var plan, state blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// extract links from plan and state
	planLinks := plan.GetLinks(ctx, &resp.Diagnostics)
	stateLinks := state.GetLinks(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// digests uniquely identify an endpoint. Make a map for quick lookup by digest
	planLinksByDigest := make(map[string]blueprint.DatacenterGenericSystemLink, len(planLinks))
	for _, link := range planLinks {
		planLinksByDigest[link.Digest()] = link
	}

	// determine whether link changes force system replacement
	linksForceReplace := true // assume the worst
	for _, stateLink := range stateLinks {
		planLink, ok := planLinksByDigest[stateLink.Digest()]
		if !ok {
			continue // the link is new - continue to assume the worst
		}

		//	the link is not new, but other details may have changed...
		if stateLink.TargetSwitchIfTransformId.ValueInt64() == planLink.TargetSwitchIfTransformId.ValueInt64() {
			// plan and state link details (switch, port, transform) match for at least one link. The server survives.
			linksForceReplace = false
			break
		}

		// the target switch and port are the same, but the transform id has changed
		if !planLink.LagMode.Equal(stateLink.LagMode) {
			// the lag mode AND transform have changed. Server must be replaced. because we cannot update lag transform link-by-link
			break
		}

		// the target switch and port are the same, the transform id has changed and we are not part of a LAG
		if len(stateLinks) > 1 {
			// because we have other links, transform/speed changes like this will be handled by replacing
			// links one-at-a-time (the other links will prevent the server from being orphaned)
			linksForceReplace = false
			break
		}
	}

	if linksForceReplace {
		resp.RequiresReplace.Append(path.Root("links"))
	}
}

func (o *resourceDatacenterGenericSystem) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterGenericSystem
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

	// prep a generic system creation request
	request := plan.CreateRequest(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the new generic system. unfortunately we only learn the link IDs, not the generic system ID
	linkIds, err := bp.CreateLinksWithNewSystem(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("failed to create generic system", err.Error())
		return
	}

	// use link IDs to learn the generic system ID
	genericSystemId, err := bp.SystemNodeFromLinkIds(ctx, linkIds, apstra.SystemNodeRoleGeneric)
	if err != nil {
		sb := new(strings.Builder)
		for i, linkId := range linkIds {
			if i == 0 {
				sb.WriteString(`"` + string(linkId) + `"`)
			} else {
				sb.WriteString(`, "` + string(linkId) + `"`)
			}
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to determine new generic system ID from returned link IDs: [%s]", sb.String()),
			err.Error())
	}
	plan.Id = types.StringValue(genericSystemId.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // provisional state in case of error below

	// set generic system properties sending <nil> for prior state
	plan.SetProperties(ctx, bp, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// pull Apstra-generated strings as needed
	err = plan.ReadSystemProperties(ctx, bp, false)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to retrieve properties from new generic system %s", plan.Id), err.Error())
		// don't return here - still want to set the state
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
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

	// Read various fields using the web UI's system API endpoint. This has a
	// side effect of discovering whether the generic system has been deleted.
	err = state.ReadSystemProperties(ctx, bp, true)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to retrieve generic system %s", state.Id), err.Error())
		return
	}

	// read tags
	state.ReadTags(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// read link info
	state.ReadLinks(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterGenericSystem) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterGenericSystem
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
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
			fmt.Sprintf("failed to lock blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	plan.UpdateHostnameAndName(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set generic system properties using prior state to skip unnecessary API calls
	plan.SetProperties(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.UpdateTags(ctx, bp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.UpdateLinkSet(ctx, &state, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterGenericSystem) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.DatacenterGenericSystem
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

	// Delete generic system
	err = bp.DeleteGenericSystem(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}

		var pendingDiags diag.Diagnostics
		pendingDiags.AddError("failed to delete generic system", err.Error())

		var ace apstra.ClientErr
		if !(errors.As(err, &ace) && ace.Type() == apstra.ErrCtAssignedToLink && ace.Detail() != nil && state.ClearCtsOnDestroy.ValueBool()) {
			resp.Diagnostics.Append(pendingDiags...) // cannot handle error
			return
		}

		// attempt to handle error by clearing CTs
		state.ClearConnectivityTemplatesFromLinks(ctx, ace.Detail().(apstra.ErrCtAssignedToLinkDetail).LinkIds, bp, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(pendingDiags...)
			return
		}

		// try again
		err = bp.DeleteGenericSystem(ctx, apstra.ObjectId(state.Id.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("failed to delete generic system after clearing CTs from interfaces", err.Error())
			resp.Diagnostics.Append(pendingDiags...)
		}
	}
}

func (o *resourceDatacenterGenericSystem) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterGenericSystem) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
