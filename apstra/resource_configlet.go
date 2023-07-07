package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceConfiglet{}
var _ resource.ResourceWithValidateConfig = &resourceConfiglet{}

type resourceConfiglet struct {
	client *apstra.Client
}

func (o *resourceConfiglet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlet"
}

func (o *resourceConfiglet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceConfiglet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a specific Configlet.\n\n",
		Attributes:          design.Configlet{}.ResourceAttributes(),
	}
}
func (o *resourceConfiglet) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// create a map of each friendly (aligned with the web UI) config section names keyed by platform
	platformToAllowedSectionsMap := map[apstra.PlatformOS][]string{
		apstra.PlatformOSJunos: {
			utils.StringersToFriendlyString(apstra.ConfigletSectionSystem, apstra.PlatformOSJunos),
			utils.StringersToFriendlyString(apstra.ConfigletSectionSetBasedSystem, apstra.PlatformOSJunos),
			utils.StringersToFriendlyString(apstra.ConfigletSectionSetBasedInterface, apstra.PlatformOSJunos),
			utils.StringersToFriendlyString(apstra.ConfigletSectionDeleteBasedInterface, apstra.PlatformOSJunos),
			utils.StringersToFriendlyString(apstra.ConfigletSectionInterface, apstra.PlatformOSJunos),
		},
		apstra.PlatformOSCumulus: {
			apstra.ConfigletSectionFRR.String(),
			apstra.ConfigletSectionInterface.String(),
			apstra.ConfigletSectionFile.String(),
			apstra.ConfigletSectionOSPF.String(),
		},
		apstra.PlatformOSNxos: {
			apstra.ConfigletSectionSystem.String(),
			apstra.ConfigletSectionInterface.String(),
			apstra.ConfigletSectionSystemTop.String(),
			apstra.ConfigletSectionOSPF.String(),
		},
		apstra.PlatformOSEos: {
			apstra.ConfigletSectionSystem.String(),
			apstra.ConfigletSectionInterface.String(),
			apstra.ConfigletSectionSystemTop.String(),
			apstra.ConfigletSectionOSPF.String(),
		},
		apstra.PlatformOSSonic: {
			apstra.ConfigletSectionSystem.String(),
			apstra.ConfigletSectionFile.String(),
			apstra.ConfigletSectionOSPF.String(),
			apstra.ConfigletSectionFRR.String(),
		},
	}

	// Retrieve values from config
	var config design.Configlet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
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
		// extract the platform/config_style from the generator object as an SDK iota type
		var platform apstra.PlatformOS
		err := platform.FromString(generator.ConfigStyle.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("generators").AtListIndex(i),
				fmt.Sprintf("unknown config style %q validation should have caught this", platform),
				err.Error())
			return
		}

		// ensure that the validation map has an entry for this platform
		var ok bool
		if _, ok = platformToAllowedSectionsMap[platform]; !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("generators").AtListIndex(i),
				fmt.Sprintf("cannot validate config style %q config sections - this is a provider issue", platform),
				fmt.Sprintf("cannot validate config style %q config sections - this is a provider issue", platform))
			return
		}

		// ensure that the configured section is valid for the specified platform
		if !utils.SliceContains(generator.Section.ValueString(), platformToAllowedSectionsMap[platform]) {
			resp.Diagnostics.Append(
				validatordiag.InvalidAttributeCombinationDiagnostic(
					path.Root("generators").AtListIndex(i),
					fmt.Sprintf("config style %q allows sections \"%s\", got %s",
						platform, strings.Join(platformToAllowedSectionsMap[platform], "\", \""), generator.Section),
				),
			)
		}
	}
}

func (o *resourceConfiglet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

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
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	var state design.Configlet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.Configlet
	var ace apstra.ApstraClientErr
	api, err = o.client.GetConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
		resp.State.RemoveResource(ctx)
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
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan design.Configlet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	c := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Update Configlet
	err := o.client.UpdateConfiglet(ctx, apstra.ObjectId(plan.Id.ValueString()), c)
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error updating Configlet", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceConfiglet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}
	var state design.Configlet

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Delete Configlet by calling API
	err := o.client.DeleteConfiglet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != apstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Configlet", err.Error())
			return
		}
	}
}
