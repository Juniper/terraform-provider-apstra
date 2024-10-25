package tfapstra

import (
	"context"
	"fmt"
	"math"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/freeform"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceFreeformResource{}
	_ resource.ResourceWithValidateConfig = &resourceFreeformResource{}
	_ resourceWithSetFfBpClientFunc       = &resourceFreeformResource{}
	_ resourceWithSetBpLockFunc           = &resourceFreeformResource{}
)

type resourceFreeformResource struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceFreeformResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_resource"
}

func (o *resourceFreeformResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceFreeformResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This resource creates a Resource in a Freeform Blueprint.",
		Attributes:          freeform.Resource{}.ResourceAttributes(),
	}
}

func (o *resourceFreeformResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config
	var config freeform.Resource
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// cannot proceed with unknown values
	if config.Type.IsUnknown() ||
		config.AllocatedFrom.IsUnknown() ||
		config.IntValue.IsUnknown() ||
		config.Ipv4Value.IsUnknown() ||
		config.Ipv6Value.IsUnknown() {
		return
	}

	var resourceType enum.FFResourceType
	err := utils.ApiStringerFromFriendlyString(&resourceType, config.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("type"), "failed to parse 'type' attribute", err.Error())
		return
	}

	// Reminder Logic: the unknown state in this function means a value was passed by reference
	switch resourceType {
	case enum.FFResourceTypeAsn:
		if (config.AllocatedFrom.IsNull() && config.IntValue.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.IntValue.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` and `integer_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && (config.IntValue.ValueInt64() < constants.AsnMin || config.IntValue.ValueInt64() > constants.AsnMax) {
			resp.Diagnostics.AddAttributeError(
				path.Root("integer_value"),
				errInvalidConfig,
				fmt.Sprintf("When type is %s, value must be between %d and %d, got %s", config.Type, constants.AsnMin, uint32(constants.AsnMax), config.IntValue.String()),
			)
		}
	case enum.FFResourceTypeVni:
		if (config.AllocatedFrom.IsNull() && config.IntValue.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.IntValue.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` and `integer_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && (config.IntValue.ValueInt64() < constants.VniMin || config.IntValue.ValueInt64() > constants.VniMax) {
			resp.Diagnostics.AddAttributeError(
				path.Root("integer_value"),
				errInvalidConfig,
				fmt.Sprintf("When type is %s, value must be between %d and %d, got %s", config.Type, constants.VniMin, constants.VniMax, config.IntValue.String()),
			)
		}
	case enum.FFResourceTypeVlan:
		if (config.AllocatedFrom.IsNull() && config.IntValue.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.IntValue.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` and `integer_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && (config.IntValue.ValueInt64() < constants.VlanMinUsable || config.IntValue.ValueInt64() > constants.VlanMaxUsable) {
			resp.Diagnostics.AddAttributeError(
				path.Root("integer_value"),
				errInvalidConfig,
				fmt.Sprintf("When type is %s, value must be between %d and %d, got %s", config.Type, constants.VlanMinUsable, constants.VlanMaxUsable, config.IntValue.String()),
			)
		}
	case enum.FFResourceTypeInt:
		if (config.AllocatedFrom.IsNull() && config.IntValue.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.IntValue.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` and `integer_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && (config.IntValue.ValueInt64() < 1 || config.IntValue.ValueInt64() > math.MaxUint32) {
			resp.Diagnostics.AddAttributeError(
				path.Root("integer_value"),
				errInvalidConfig,
				fmt.Sprintf("When type is %s, value must be between %d and %d, got %s", config.Type, 1, uint32(math.MaxUint32), config.IntValue.String()),
			)
		}
	case enum.FFResourceTypeHostIpv4:
		if (config.AllocatedFrom.IsNull() && config.Ipv4Value.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.Ipv4Value.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` and `ipv4_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
	case enum.FFResourceTypeHostIpv6:
		if (config.AllocatedFrom.IsNull() && config.Ipv6Value.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.Ipv6Value.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` or `ipv6_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
	case enum.FFResourceTypeIpv4:
		if (config.AllocatedFrom.IsNull() && config.Ipv4Value.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.Ipv4Value.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` and `ipv4_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && !config.Ipv4Value.IsNull() {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` is used to indicate the Subnet Prefix Length. It must not be set when `ipv4_value` is set and `type` is set to "+config.Type.String(),
			)
		}
		if !config.AllocatedFrom.IsNull() && config.IntValue.IsNull() {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` is used to indicate the Subnet Prefix Length. It must be set when `allocated_from` is set and `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && (config.IntValue.ValueInt64() < 1 || config.IntValue.ValueInt64() > 32) {
			resp.Diagnostics.AddAttributeError(
				path.Root("integer_value"),
				errInvalidConfig,
				fmt.Sprintf("When type is %s, value must be between %d and %d, got %s", config.Type, 1, 32, config.IntValue.String()),
			)
		}
	case enum.FFResourceTypeIpv6:
		if (config.AllocatedFrom.IsNull() && config.Ipv6Value.IsNull()) ||
			(!config.AllocatedFrom.IsNull() && !config.Ipv6Value.IsNull()) {
			resp.Diagnostics.AddError(
				errInvalidConfig,
				"Exactly one of `allocated_from` and `ipv6_value` must be set when `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && !config.Ipv6Value.IsNull() {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` is used to indicate the Subnet Prefix Length. It must not be set when `ipv6_value` is set and `type` is set to "+config.Type.String(),
			)
		}
		if !config.AllocatedFrom.IsNull() && config.IntValue.IsNull() {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` is used to indicate the Subnet Prefix Length. It must be set when `allocated_from` is set and `type` is set to "+config.Type.String(),
			)
		}
		if !config.IntValue.IsNull() && (config.IntValue.ValueInt64() < 1 || config.IntValue.ValueInt64() > 128) {
			resp.Diagnostics.AddAttributeError(
				path.Root("integer_value"),
				errInvalidConfig,
				fmt.Sprintf("When type is %s, value must be between %d and %d, got %s", config.Type, 1, 128, config.IntValue.String()),
			)
		}
	}
}

func (o *resourceFreeformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan freeform.Resource
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

	// Create the resource
	id, err := bp.CreateRaResource(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new Resource", err.Error())
		return
	}

	// record the id and provisionally set the state
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set the resource assignments, if any
	if !plan.AssignedTo.IsNull() {
		var assignments []apstra.ObjectId
		resp.Diagnostics.Append(plan.AssignedTo.ElementsAs(ctx, &assignments, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		err = bp.UpdateResourceAssignments(ctx, id, assignments)
		if err != nil {
			resp.Diagnostics.AddError("error updating Resource Assignments", err.Error())
			return
		}
	}

	// Read the resource back from Apstra to get computed values
	api, err := bp.GetRaResource(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error reading just created Resource", err.Error())
		return
	}

	// load state objects
	plan.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state freeform.Resource
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

	api, err := bp.GetRaResource(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Freeform Resource", err.Error())
		return
	}

	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the resource assignments
	assignedTo, err := bp.ListResourceAssignments(ctx, api.Id)
	if err != nil {
		resp.Diagnostics.AddError("error reading Resource Assignments", err.Error())
		return
	}

	state.AssignedTo = utils.SetValueOrNull(ctx, types.StringType, assignedTo, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceFreeformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan freeform.Resource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get state values
	var state freeform.Resource
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

	// Update the resource if necessary
	if plan.NeedsUpdate(state) {
		// Convert the plan into an API Request
		request := plan.Request(ctx, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// Update the Resource
		err = bp.UpdateRaResource(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
		if err != nil {
			resp.Diagnostics.AddError("error updating Freeform Resource", err.Error())
			return
		}
	}

	var planAssignments, stateAssignments []apstra.ObjectId
	resp.Diagnostics.Append(plan.AssignedTo.ElementsAs(ctx, &planAssignments, false)...)
	resp.Diagnostics.Append(state.AssignedTo.ElementsAs(ctx, &stateAssignments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the assignments if necessary
	if !utils.SlicesAreEqualSets(planAssignments, stateAssignments) {
		err = bp.UpdateResourceAssignments(ctx, apstra.ObjectId(plan.Id.ValueString()), planAssignments)
		if err != nil {
			resp.Diagnostics.AddError("error updating Resource Assignments", err.Error())
			return
		}
	}

	// Read the resource back from Apstra to get computed values
	api, err := bp.GetRaResource(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error reading just updated Resource", err.Error())
		return
	}

	plan.LoadApiData(ctx, api.Data, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state freeform.Resource
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
	err = bp.DeleteRaResource(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Freeform Resource", err.Error())
		return
	}
}

func (o *resourceFreeformResource) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceFreeformResource) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
