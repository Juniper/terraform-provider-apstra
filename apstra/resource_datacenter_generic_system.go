package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterGenericSystem{}
var _ resourceWithSetBpClientFunc = &resourceDatacenterGenericSystem{}
var _ resourceWithSetBpLockFunc = &resourceDatacenterGenericSystem{}

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
