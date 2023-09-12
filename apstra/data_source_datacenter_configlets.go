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

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterConfiglets{}

type dataSourceDatacenterConfiglets struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterConfiglets) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_configlets"
}

func (o *dataSourceDatacenterConfiglets) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterConfiglets) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the ID numbers of all Configlets in a Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the Configlet belongs to.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
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

func (o *dataSourceDatacenterConfiglets) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		BlueprintId        types.String `tfsdk:"blueprint_id"`
		Ids                types.Set    `tfsdk:"ids"`
		SupportedPlatforms types.Set    `tfsdk:"supported_platforms"`
	}

	// get the configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(config.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found",
				config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	var ids []apstra.ObjectId
	if config.SupportedPlatforms.IsNull() {
		// no required platform filters
		ids, err = bpClient.GetAllConfigletIds(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving Configlet IDs", err.Error())
			return
		}
	} else {
		// extract required platform filters
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

		configlets, err := bpClient.GetAllConfiglets(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving Configlets", err.Error())
			return
		}

		ids = make([]apstra.ObjectId, len(configlets))
		var count int
		for i := range configlets {
			if utils.ConfigletSupportsPlatforms(configlets[i].Data.Data, platforms) {
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
		BlueprintId        types.String `tfsdk:"blueprint_id"`
		Ids                types.Set    `tfsdk:"ids"`
		SupportedPlatforms types.Set    `tfsdk:"supported_platforms"`
	}{
		BlueprintId:        config.BlueprintId,
		Ids:                idSet,
		SupportedPlatforms: config.SupportedPlatforms,
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
