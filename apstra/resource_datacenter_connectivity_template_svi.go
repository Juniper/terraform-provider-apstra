package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint/connectivity_templates"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint/connectivity_templates/primitives"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceDatacenterConnectivityTemplateSvi{}
	_ resource.ResourceWithValidateConfig = &resourceDatacenterConnectivityTemplateSvi{}
	_ resourceWithSetDcBpClientFunc       = &resourceDatacenterConnectivityTemplateSvi{}
	_ resourceWithSetBpLockFunc           = &resourceDatacenterConnectivityTemplateSvi{}
)

type resourceDatacenterConnectivityTemplateSvi struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterConnectivityTemplateSvi) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_connectivity_template_svi"
}

func (o *resourceDatacenterConnectivityTemplateSvi) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterConnectivityTemplateSvi) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates a Connectivity Template suitable for use " +
			"with Application Points of type *SVI* within a Datacenter Blueprint.",
		Attributes: connectivitytemplates.ConnectivityTemplateSvi{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterConnectivityTemplateSvi) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from plan.
	var config connectivitytemplates.ConnectivityTemplateSvi
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validation not possible when DynamicBgpPeerings is unknown
	if config.DynamicBgpPeerings.IsUnknown() {
		return
	}

	// validation not possible when any individual DynamicBgpPeering is unknown
	for _, v := range config.DynamicBgpPeerings.Elements() {
		if v.IsUnknown() {
			return
		}
	}

	// extract DynamicBgpPeerings
	var dynamicBgpPeerings []primitives.DynamicBgpPeering
	resp.Diagnostics.Append(config.DynamicBgpPeerings.ElementsAs(ctx, &dynamicBgpPeerings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, dynamicBgpPeering := range dynamicBgpPeerings {
		// extract set value for use in error pathing.
		// Note this doesn't currently work. https://github.com/hashicorp/terraform/issues/33491
		setVal, d := types.ObjectValueFrom(ctx, dynamicBgpPeering.AttrTypes(), &dynamicBgpPeering)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		dynamicBgpPeering.ValidateConfig(ctx, path.Root("dynamic_bgp_peerings").AtSetValue(setVal), &resp.Diagnostics)
	}
}

func (o *resourceDatacenterConnectivityTemplateSvi) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan connectivitytemplates.ConnectivityTemplateSvi
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

	// create an API request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// send the request to Apstra
	err = bp.CreateConnectivityTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Connectivity Template", err.Error())
		return
	}

	// set the state
	plan.Id = types.StringValue(string(*request.Id))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConnectivityTemplateSvi) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state connectivitytemplates.ConnectivityTemplateSvi
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

	api, err := bp.GetConnectivityTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to retrieve Connectivity Template %s", state.Id), err.Error())
		return
	}

	state.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterConnectivityTemplateSvi) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan connectivitytemplates.ConnectivityTemplateSvi
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	// create an API request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// send the request to Apstra
	err = bp.UpdateConnectivityTemplate(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Connectivity Template", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConnectivityTemplateSvi) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state connectivitytemplates.ConnectivityTemplateSvi
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
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Delete connectivity template
	err = bp.DeleteConnectivityTemplate(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed while deleting Connectivity Template %s", state.Id), err.Error())
	}
}

func (o *resourceDatacenterConnectivityTemplateSvi) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterConnectivityTemplateSvi) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
