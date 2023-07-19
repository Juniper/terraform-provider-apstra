package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprints{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceBlueprints{}
var _ versionValidator = &dataSourceBlueprints{}

type dataSourceBlueprints struct {
	client           *apstra.Client
	minClientVersion *version.Version
	maxClientVersion *version.Version
}

func (o *dataSourceBlueprints) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprints"
}

func (o *dataSourceBlueprints) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceBlueprints) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the ID numbers of Blueprints.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"reference_design": schema.StringAttribute{
				MarkdownDescription: "Optional filter to select only Blueprints matching the specified Reference Design.",
				Optional:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					utils.StringersToFriendlyString(apstra.RefDesignDatacenter),
					apstra.RefDesignFreeform.String(),
				)},
			},
		},
	}
}

func (o *dataSourceBlueprints) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if o.client == nil { // cannot proceed without a client
		return
	}

	var config blueprints
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.RefDesign.ValueString() == apstra.RefDesignFreeform.String() {
		var err error
		o.minClientVersion, err = version.NewVersion("4.1.2")
		if err != nil {
			resp.Diagnostics.AddError(errProviderBug,
				fmt.Sprintf("error parsing min/max version - %s", err.Error()))

			return
		}
	}

	o.checkVersion(ctx, &resp.Diagnostics)
}

func (o *dataSourceBlueprints) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprints
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ids []apstra.ObjectId
	var err error
	if config.RefDesign.IsNull() {
		ids, err = o.client.ListAllBlueprintIds(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error listing Blueprint IDs", err.Error())
			return
		}
	} else {
		bpStatuses, err := o.client.GetAllBlueprintStatus(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving Blueprint statuses", err.Error())
			return
		}

		for _, bpStatus := range bpStatuses {
			if utils.StringersToFriendlyString(bpStatus.Design) == config.RefDesign.ValueString() {
				ids = append(ids, bpStatus.Id)
			}
		}
	}

	idSet, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// save the list of IDs in the config object
	config.Ids = idSet

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type blueprints struct {
	Ids       types.Set    `tfsdk:"ids"`
	RefDesign types.String `tfsdk:"reference_design"`
}

func (o *dataSourceBlueprints) apiVersion() (*version.Version, error) {
	if o.client == nil {
		return nil, nil
	}
	return version.NewVersion(o.client.ApiVersion())
}

func (o *dataSourceBlueprints) cfgVersionMin() (*version.Version, error) {
	return o.minClientVersion, nil
}

func (o *dataSourceBlueprints) cfgVersionMax() (*version.Version, error) {
	return o.maxClientVersion, nil
}

func (o *dataSourceBlueprints) checkVersion(ctx context.Context, diags *diag.Diagnostics) {
	checkVersionCompatibility(ctx, o, diags)
}
