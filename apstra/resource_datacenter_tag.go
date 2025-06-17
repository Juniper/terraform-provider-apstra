package tfapstra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.ResourceWithConfigure   = &resourceDatacenterTag{}
	_ resource.ResourceWithImportState = &resourceDatacenterTag{}
	_ resourceWithSetDcBpClientFunc    = &resourceDatacenterTag{}
	_ resourceWithSetBpLockFunc        = &resourceDatacenterTag{}
)

type resourceDatacenterTag struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterTag) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_tag"
}

func (o *resourceDatacenterTag) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterTag) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a Tag within a Datacenter Blueprint.",
		Attributes:          blueprint.Tag{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterTag) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var importId struct {
		BlueprintId string `json:"blueprint_id"`
		Id          string `json:"id"`
		Name        string `json:"name"`
	}

	// parse the user-supplied import ID string JSON
	err := json.Unmarshal([]byte(req.ID), &importId)
	if err != nil {
		resp.Diagnostics.AddError("failed parsing import id JSON string", err.Error())
		return
	}

	if importId.BlueprintId == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'blueprint_id' element of import ID string is required")
	}
	if importId.Id == "" && importId.Name == "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "One of 'id' and 'name' element of import ID string is required")
	}
	if importId.Id != "" && importId.Name != "" {
		resp.Diagnostics.AddError(errImportJsonMissingRequiredField, "'id' and 'name' elements of import ID string cannot both be set")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// create a state object preloaded with details we have now
	state := blueprint.Tag{
		BlueprintId: types.StringValue(importId.BlueprintId),
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

	// read the tag from the API
	var tag apstra.TwoStageL3ClosTag
	switch {
	case importId.Name != "":
		tag, err = bp.GetTagByLabel(ctx, importId.Name)
	case importId.Id != "":
		tag, err = bp.GetTag(ctx, apstra.ObjectId(importId.Id))
	}
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch tag", err.Error())
		return
	}

	state.LoadApiData(ctx, tag.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterTag) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.Tag
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

	// prepare a request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the tag
	id, err := bp.CreateTag(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create tag", err.Error())
		return
	}

	// save the ID (not exposed to the user) in private state
	p := private.ResourceDatacenterTag{Id: id}
	p.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterTag) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state blueprint.Tag
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

	tags, err := bp.GetAllTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to get blueprint tags", err.Error())
		return
	}

	var tag *apstra.TwoStageL3ClosTag
	for _, v := range tags {
		if v.Data.Label == state.Name.ValueString() {
			tag = &v
			break
		}
	}
	if tag == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.LoadApiData(ctx, tag.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// save the ID (not exposed to the user) in private state
	p := private.ResourceDatacenterTag{Id: tag.Id}
	p.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterTag) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.Tag
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from private state
	var p private.ResourceDatacenterTag
	p.LoadPrivateState(ctx, req.Private, &resp.Diagnostics)
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

	// update the tag
	err = bp.UpdateTag(ctx, p.Id, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update tag", err.Error())
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterTag) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state blueprint.Tag
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from private state
	var p private.ResourceDatacenterTag
	p.LoadPrivateState(ctx, req.Private, &resp.Diagnostics)
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

	// delete the tag
	err = bp.DeleteTag(ctx, p.Id)
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to delete tag", err.Error())
		return
	}
}

func (o *resourceDatacenterTag) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterTag) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
