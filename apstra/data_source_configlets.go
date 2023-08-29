package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceConfiglets{}

type dataSourceConfiglets struct {
	client *apstra.Client
}

func (o *dataSourceConfiglets) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlets"
}

func (o *dataSourceConfiglets) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceConfiglets) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the ID numbers of all Configlets.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"supported_platforms": schema.SetAttribute{
				MarkdownDescription: "Configlets which do not support each of the specified platforms will be " +
					"filtered out of the results.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{setvalidator.ValueStringsAre(
					stringvalidator.OneOf(utils.AllPlatformOSNames()...),
				)},
			},
		},
	}
}

func (o *dataSourceConfiglets) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		Ids                types.Set `tfsdk:"ids"`
		SupportedPlatforms types.Set `tfsdk:"supported_platforms"`
	}

	// get the configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var ids []apstra.ObjectId
	if config.SupportedPlatforms.IsNull() {
		// no required platform filter
		ids, err = o.client.ListAllConfiglets(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving Configlet IDs", err.Error())
			return
		}
	} else {
		// extract required platform filter
		platformStrings := make([]string, len(config.SupportedPlatforms.Elements()))
		d := config.SupportedPlatforms.ElementsAs(ctx, &platformStrings, false)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		platforms := make([]apstra.PlatformOS, len(platformStrings))
		for i := range platformStrings {
			err := platforms[i].FromString(platformStrings[i])
			if err != nil {
				resp.Diagnostics.AddError("error parsing platform",
					fmt.Sprintf("unable to parse platform %q - %s", platformStrings[i], err.Error()))
			}
		}
		if resp.Diagnostics.HasError() {
			return
		}

		configlets, err := o.client.GetAllConfiglets(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving Configlets", err.Error())
			return
		}

		ids = make([]apstra.ObjectId, len(configlets))
		var count int
		for i := range configlets {
			if utils.ConfigletSupportsPlatforms(&configlets[i], platforms) {
				ids[count] = configlets[i].Id
				count++
			}
		}
		ids = ids[:count]
	}

	idSet, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create new state object
	state := struct {
		Ids                types.Set `tfsdk:"ids"`
		SupportedPlatforms types.Set `tfsdk:"supported_platforms"`
	}{
		Ids:                idSet,
		SupportedPlatforms: config.SupportedPlatforms,
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
