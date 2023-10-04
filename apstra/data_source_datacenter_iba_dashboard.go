package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterIbaDashboard{}

type dataSourceDatacenterIbaDashboard struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterIbaDashboard) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_iba_dashboard"
}

func (o *dataSourceDatacenterIbaDashboard) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterIbaDashboard) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific IBA Dashboard in a Blueprint." +
			"\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.IbaDashboard{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterIbaDashboard) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.IbaDashboard
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
	case !config.Label.IsNull():
		api, err = bpClient.GetIbaDashboardByLabel(ctx, config.Label.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"IBA dashboard not found",
					fmt.Sprintf("IBA Dashboard with label %s not found", config.Label))
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
					"IbaDashboard not found",
					fmt.Sprintf("IbaDashboard with ID %s not found", config.Id))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading IBA Dashboardt", err.Error(),
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
