package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceTelemetryServiceRegistryEntries{}
var _ datasourceWithSetClient = &dataSourceTelemetryServiceRegistryEntries{}

type dataSourceTelemetryServiceRegistryEntries struct {
	client *apstra.Client
}

func (o *dataSourceTelemetryServiceRegistryEntries) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_telemetry_service_registry_entries"
}

func (o *dataSourceTelemetryServiceRegistryEntries) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceTelemetryServiceRegistryEntries) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This data source returns the Service Names of all Telemetry Service Registry Entries.",
		Attributes: map[string]schema.Attribute{
			"service_names": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"built_in": schema.BoolAttribute{
				MarkdownDescription: "Return only built_in if true, not built_in if false, all otherwise",
				Optional:            true,
			},
		},
	}
}

func (o *dataSourceTelemetryServiceRegistryEntries) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		ServiceNames types.Set  `tfsdk:"service_names"`
		BuiltIn      types.Bool `tfsdk:"built_in"`
	}

	// get the configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	tses, err := o.client.GetAllTelemetryServiceRegistryEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving Property Set IDs", err.Error())
		return
	}
	var snames []string
	for _, t := range tses {
		if config.BuiltIn.IsUnknown() || config.BuiltIn.IsNull() || config.BuiltIn.ValueBool() == t.Builtin {
			snames = append(snames, t.ServiceName)
		}
	}

	snameSet, diags := types.SetValueFrom(ctx, types.StringType, snames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create new state object
	var state struct {
		ServiceNames types.Set  `tfsdk:"service_names"`
		BuiltIn      types.Bool `tfsdk:"built_in"`
	}
	state.ServiceNames = snameSet
	state.BuiltIn = config.BuiltIn
	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceTelemetryServiceRegistryEntries) setClient(client *apstra.Client) {
	o.client = client
}
