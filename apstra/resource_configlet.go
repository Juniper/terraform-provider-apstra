package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceConfiglet{}
	_ resource.ResourceWithValidateConfig = &resourceConfiglet{}
	_ resourceWithSetClient               = &resourceConfiglet{}
)

type resourceConfiglet struct {
	client *apstra.Client
}

func (o *resourceConfiglet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlet"
}

func (o *resourceConfiglet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceConfiglet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a specific Configlet.\n\n",
		Attributes:          design.Configlet{}.ResourceAttributes(),
	}
}

func (o *resourceConfiglet) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config
	var config design.Configlet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delay Validation until the involved attributes have a known value.
	if config.Generators.IsUnknown() {
		return
	}

	// extract generators from config
	var generators []design.ConfigletGenerator
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
		err := utils.ApiStringerFromFriendlyString(&configletStyle, generator.ConfigStyle.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("generators").AtListIndex(i),
				fmt.Sprintf("failed to parse config_style %s", generator.ConfigStyle), err.Error(),
			)
		}

		// parse the config section
		var configletSection enum.ConfigletSection
		err = utils.ApiStringerFromFriendlyString(&configletSection, generator.Section.ValueString(), generator.ConfigStyle.ValueString())
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

func (o *resourceConfiglet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan design.Configlet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateConfiglet(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Configlet", err.Error())
		return
	}

	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceConfiglet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state design.Configlet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := o.client.GetConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read Configlet", err.Error())
		return
	}

	state.Id = types.StringValue(string(api.Id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceConfiglet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan design.Configlet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create request
	c := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Configlet
	err := o.client.UpdateConfiglet(ctx, apstra.ObjectId(plan.Id.ValueString()), c)
	if err != nil {
		resp.Diagnostics.AddError("error updating Configlet", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceConfiglet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state design.Configlet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Configlet by calling API
	err := o.client.DeleteConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Configlet", err.Error())
		return
	}
}

func (o *resourceConfiglet) setClient(client *apstra.Client) {
	o.client = client
}
