package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
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
		Attributes:          freeform.FreeformResource{}.ResourceAttributes(),
	}
}

func (o *resourceFreeformResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config
	var config freeform.FreeformResource
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resourceType apstra.FFResourceType
	err := utils.ApiStringerFromFriendlyString(&resourceType, config.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("type"), "failed to parse 'type' attribute", err.Error())
		return
	}

	switch resourceType {
	case apstra.FFResourceTypeAsn,
		apstra.FFResourceTypeVni,
		apstra.FFResourceTypeVlan,
		apstra.FFResourceTypeInt:
		if !utils.HasValue(config.AllocatedFrom) && !utils.HasValue(config.IntValue) {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"Either `allocated_from` or `integer_value` must also be set when `type` is set to "+config.Type.String(),
			)
		}
	case apstra.FFResourceTypeHostIpv4:
		if !utils.HasValue(config.AllocatedFrom) && !utils.HasValue(config.Ipv4Value) {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"Either `allocated_from` or `ipv4_value` must also be set when `type` is set to "+config.Type.String(),
			)
		}
	case apstra.FFResourceTypeHostIpv6:
		if !utils.HasValue(config.AllocatedFrom) && !utils.HasValue(config.Ipv6Value) {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"Either `allocated_from` or `ipv6_value` must also be set when `type` is set to "+config.Type.String(),
			)
		}
	case apstra.FFResourceTypeIpv4:
		if !utils.HasValue(config.AllocatedFrom) && !utils.HasValue(config.Ipv4Value) {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"Either `allocated_from` or `ipv4_value` must also be set when `type` is set to "+config.Type.String(),
			)
		}
		if utils.HasValue(config.IntValue) && utils.HasValue(config.Ipv4Value) {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` must not be set when `ipv4_value` is set and `type` is set to "+config.Type.String(),
			)
		}
		if utils.HasValue(config.AllocatedFrom) && !utils.HasValue(config.IntValue) {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` is used to indicate the Subnet Prefix Length. It must be set when `allocated_from` is set and `type` is set to "+config.Type.String(),
			)
		}
	case apstra.FFResourceTypeIpv6:
		if !utils.HasValue(config.AllocatedFrom) && !utils.HasValue(config.Ipv6Value) {
			resp.Diagnostics.AddError(
				"Missing required attribute",
				"Either `allocated_from` or `ipv6_value` must also be set when `type` is set to "+config.Type.String(),
			)
		}
		if utils.HasValue(config.IntValue) && utils.HasValue(config.Ipv6Value) {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` must not be set when `ipv6_value` is set and `type` is set to "+config.Type.String(),
			)
		}
		if utils.HasValue(config.AllocatedFrom) && !utils.HasValue(config.IntValue) {
			resp.Diagnostics.AddError(
				"Conflicting Attributes",
				"`integer_value` is used to indicate the Subnet Prefix Length. It must be set when `allocated_from` is set and `type` is set to "+config.Type.String(),
			)
		}
	}
}

func (o *resourceFreeformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan freeform.FreeformResource
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

	// Read the resource back from Apstra to get computed values
	api, err := bp.GetRaResource(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error reading just created Resource", err.Error())
		return
	}

	// load state objects
	plan.Id = types.StringValue(id.String())
	plan.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state freeform.FreeformResource
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

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceFreeformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan freeform.FreeformResource
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

	// Update the Resource
	err = bp.UpdateRaResource(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Freeform Resource", err.Error())
		return
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
	var state freeform.FreeformResource
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
