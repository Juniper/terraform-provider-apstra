package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/analytics"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceTelemetryServiceRegistryEntry{}
	_ datasourceWithSetClient            = &dataSourceTelemetryServiceRegistryEntry{}
)

type dataSourceTelemetryServiceRegistryEntry struct {
	client *apstra.Client
}

func (o *dataSourceTelemetryServiceRegistryEntry) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_telemetry_service_registry_entry"
}

func (o *dataSourceTelemetryServiceRegistryEntry) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceTelemetryServiceRegistryEntry) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This data source provides details of a specific Telemetry Service Registry Entry.",
		Attributes:          analytics.TelemetryServiceRegistryEntry{}.DataSourceAttributes(),
	}
}

func (o *dataSourceTelemetryServiceRegistryEntry) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config analytics.TelemetryServiceRegistryEntry
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := o.client.GetTelemetryServiceRegistryEntry(ctx, config.ServiceName.ValueString())
	if utils.IsApstra404(err) {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"TelemetryServiceRegistryEntry not found",
			fmt.Sprintf("TelemetryServiceRegistryEntry with Name %q not found", config.ServiceName.ValueString()))
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving TelemetryServiceRegistryEntry", err.Error())
		return
	}

	// create new state object
	var state analytics.TelemetryServiceRegistryEntry
	state.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceTelemetryServiceRegistryEntry) setClient(client *apstra.Client) {
	o.client = client
}
