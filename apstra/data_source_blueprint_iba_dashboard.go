package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/iba"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprintIbaDashboard{}

type dataSourceBlueprintIbaDashboard struct {
	client *apstra.Client
}

func (o *dataSourceBlueprintIbaDashboard) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_dashboard"
}

func (o *dataSourceBlueprintIbaDashboard) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceBlueprintIbaDashboard) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source provides details of a specific IBA Dashboard in a Blueprint." +
			"\n\n" +
			"At least one optional attribute is required.",
		Attributes: iba.Dashboard{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintIbaDashboard) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config iba.Dashboard
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

	var api *apstra.IbaDashboard
	switch {
	case !config.Name.IsNull():
		api, err = bpClient.GetIbaDashboardByLabel(ctx, config.Name.ValueString())
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
		api, err = bpClient.GetIbaDashboard(ctx, apstra.ObjectId(config.Id.ValueString()))
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
