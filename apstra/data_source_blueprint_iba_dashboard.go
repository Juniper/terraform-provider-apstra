package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/iba"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithConfigure      = &dataSourceBlueprintIbaDashboard{}
	_ datasource.DataSourceWithValidateConfig = &dataSourceBlueprintIbaDashboard{}
	_ datasourceWithSetDcBpClientFunc         = &dataSourceBlueprintIbaDashboard{}
	_ datasourceWithSetClient                 = &dataSourceBlueprintIbaDashboard{} // needed for API version compatibility check only
)

type dataSourceBlueprintIbaDashboard struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceBlueprintIbaDashboard) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_dashboard"
}

func (o *dataSourceBlueprintIbaDashboard) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintIbaDashboard) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source provides details of a specific IBA Dashboard in a Blueprint." +
			"\n\n" +
			"At least one optional attribute is required.",
		Attributes: iba.Dashboard{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintIbaDashboard) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	// Retrieve values from config.
	var config iba.Dashboard
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// cannot proceed to config + api version validation if the provider has not been configured
	if o.client == nil {
		return
	}

	// only supported with Apstra 4.x
	if !compatibility.BpIbaDashboardOk.Check(version.Must(version.NewVersion(o.client.ApiVersion()))) {
		resp.Diagnostics.AddError(
			"Incompatible API version",
			"This data source is compatible only with Apstra "+compatibility.BpIbaDashboardOk.String(),
		)
		return
	}
}

func (o *dataSourceBlueprintIbaDashboard) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config iba.Dashboard
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, config.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, config.BlueprintId), err.Error())
		return
	}

	var api *apstra.IbaDashboard
	switch {
	case !config.Name.IsNull():
		api, err = bp.GetIbaDashboardByLabel(ctx, config.Name.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"IBA dashboard not found",
					fmt.Sprintf("IBA Dashboard with name %s not found: Error : %q", config.Name, err))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading IBA Dashboard", err.Error(),
			)
			return
		}
	case !config.Id.IsNull():
		api, err = bp.GetIbaDashboard(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"Dashboard not found",
					fmt.Sprintf("Dashboard with ID %s not found", config.Id))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading IBA Dashboard", err.Error(),
			)
			return
		}
	}

	config.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// setClient is used for API version compatibility check only
func (o *dataSourceBlueprintIbaDashboard) setClient(client *apstra.Client) {
	o.client = client
}

func (o *dataSourceBlueprintIbaDashboard) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
