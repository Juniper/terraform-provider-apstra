package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceDatacenterConfiglet{}
	_ resource.ResourceWithValidateConfig = &resourceDatacenterConfiglet{}
	_ resourceWithSetDcBpClientFunc       = &resourceDatacenterConfiglet{}
	_ resourceWithSetBpLockFunc           = &resourceDatacenterConfiglet{}
)

type resourceDatacenterConfiglet struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterConfiglet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_configlet"
}

func (o *resourceDatacenterConfiglet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterConfiglet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource adds a Configlet to a Blueprint, either by " +
			"importing from the Global Catalog, or by creating one from scratch.",
		Attributes: blueprint.DatacenterConfiglet{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterConfiglet) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config
	var config blueprint.DatacenterConfiglet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delay Validation until the involved attributes have a known value.
	if config.Generators.IsUnknown() {
		return
	}

	// extract generators from config
	var generators []blueprint.ConfigletGenerator
	resp.Diagnostics.Append(config.Generators.ElementsAs(ctx, &generators, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validate each generator
	for i, generator := range generators {
		if generator.ConfigStyle.IsUnknown() || generator.Section.IsUnknown() {
			continue // cannot validate with unknown value
		}

		// parse the config style
		var configletStyle enum.ConfigletStyle
		err := rosetta.ApiStringerFromFriendlyString(&configletStyle, generator.ConfigStyle.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("generators").AtListIndex(i),
				fmt.Sprintf("failed to parse config_style %s", generator.ConfigStyle), err.Error(),
			)
		}

		// parse the config section
		var configletSection enum.ConfigletSection
		err = rosetta.ApiStringerFromFriendlyString(&configletSection, generator.Section.ValueString(), generator.ConfigStyle.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("generators").AtListIndex(i),
				fmt.Sprintf("failed to parse section %s",
					generator.Section), err.Error(),
			)
		}

		if resp.Diagnostics.HasError() {
			continue
		}

		if !utils.ItemInSlice(configletSection, apstra.ValidConfigletSections(configletStyle)) {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
				path.Root("generators").AtListIndex(i),
				fmt.Sprintf("Section %s not valid with config_style %s", generator.Section, generator.ConfigStyle),
			))
		}
	}
}

func (o *resourceDatacenterConfiglet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan blueprint.DatacenterConfiglet
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

	// create a plan based on the catalog configlet if a catalog ID was supplied
	if !plan.CatalogConfigletID.IsNull() {
		catalogConfiglet, err := bp.Client().GetConfiglet(ctx, apstra.ObjectId(plan.CatalogConfigletID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading Configlet from catalog", err.Error())
			return
		}

		plan.LoadCatalogConfigletData(ctx, catalogConfiglet.Data, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// create a datacenter configlet request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the datacenter configlet
	id, err := bp.CreateConfiglet(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Datacenter Configlet", err.Error())
		return
	}

	// update the plan with the configlet ID and set the state
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConfiglet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state blueprint.DatacenterConfiglet
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

	api, err := bp.GetConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to read imported Configlet %s", state.Id), err.Error())
		return
	}

	// Set state
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterConfiglet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan blueprint.DatacenterConfiglet
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

	// generate a request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Configlet
	err = bp.UpdateConfiglet(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error updating Blueprint %s Configlet %s", plan.BlueprintId, plan.Id),
			err.Error())
		return
	}

	// create new state object
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterConfiglet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state blueprint.DatacenterConfiglet
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

	// Delete Configlet by calling API
	err = bp.DeleteConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("unable to delete Configlet %s from Blueprint %s", state.Id, state.BlueprintId),
			err.Error())
	}
}

func (o *resourceDatacenterConfiglet) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterConfiglet) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
