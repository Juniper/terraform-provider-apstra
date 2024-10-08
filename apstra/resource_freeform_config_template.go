package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/freeform"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceFreeformConfigTemplate{}
	_ resourceWithSetFfBpClientFunc  = &resourceFreeformConfigTemplate{}
	_ resourceWithSetBpLockFunc      = &resourceFreeformConfigTemplate{}
)

type resourceFreeformConfigTemplate struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceFreeformConfigTemplate) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_config_template"
}

func (o *resourceFreeformConfigTemplate) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceFreeformConfigTemplate) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This resource creates a Config Template in a Freeform Blueprint.",
		Attributes:          freeform.ConfigTemplate{}.ResourceAttributes(),
	}
}

func (o *resourceFreeformConfigTemplate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan freeform.ConfigTemplate
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
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

	// Convert the plan into an API Request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := bp.CreateConfigTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new ConfigTemplate", err.Error())
		return
	}

	// record the id and provisionally set the state
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set the CT system assignments, if any
	if !plan.AssignedTo.IsNull() {
		var assignments []apstra.ObjectId
		resp.Diagnostics.Append(plan.AssignedTo.ElementsAs(ctx, &assignments, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateRequest := make(map[apstra.ObjectId]*apstra.ObjectId, len(assignments))
		for _, assignment := range assignments {
			updateRequest[assignment] = &id
		}

		err = bp.UpdateConfigTemplateAssignments(ctx, updateRequest)
		if err != nil {
			resp.Diagnostics.AddError("error updating ConfigTemplate system Assignments", err.Error())
			return
		}
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformConfigTemplate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state freeform.ConfigTemplate
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	api, err := bp.GetConfigTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving ConfigTemplate", err.Error())
		return
	}

	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the system assignments
	assignments, err := bp.GetConfigTemplateAssignments(ctx, api.Id)
	if err != nil {
		resp.Diagnostics.AddError("error reading ConfigTemplate System Assignments", err.Error())
		return
	}

	state.AssignedTo = utils.SetValueOrNull(ctx, types.StringType, assignments, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceFreeformConfigTemplate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan freeform.ConfigTemplate
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get state values
	var state freeform.ConfigTemplate
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
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

	if plan.NeedsUpdate(state) {
		request := plan.Request(ctx, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// Update Config Template
		err = bp.UpdateConfigTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
		if err != nil {
			resp.Diagnostics.AddError("error updating Config Template", err.Error())
			return
		}
	}

	// update the assignments if necessary
	if !plan.AssignedTo.Equal(state.AssignedTo) {
		var planAssignments []apstra.ObjectId
		resp.Diagnostics.Append(plan.AssignedTo.ElementsAs(ctx, &planAssignments, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		err = bp.UpdateConfigTemplateAssignmentsByTemplate(ctx, apstra.ObjectId(plan.Id.ValueString()), planAssignments)
		if err != nil {
			resp.Diagnostics.AddError("error updating Resource Assignments", err.Error())
			return
		}
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformConfigTemplate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state freeform.ConfigTemplate
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
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

	// Delete Config Template by calling API
	err = bp.DeleteConfigTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Config Template", err.Error())
		return
	}
}

func (o *resourceFreeformConfigTemplate) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceFreeformConfigTemplate) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
