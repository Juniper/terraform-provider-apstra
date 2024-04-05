package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/device"
	systemAgents "github.com/Juniper/terraform-provider-apstra/apstra/system_agents"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDeviceConfig{}
var _ datasourceWithSetClient = &dataSourceDeviceConfig{}

type dataSourceDeviceConfig struct {
	client *apstra.Client
}

func (o *dataSourceDeviceConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_config"
}

func (o *dataSourceDeviceConfig) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDeviceConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDevices + "This data source gets the device configuration information.",
		Attributes:          systemAgents.ManagedDevice{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDeviceConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config device.CfgInfo
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	systemConfig, err := o.client.GetSystemConfig(ctx, apstra.ObjectId(config.SystemId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("System %s not found",
				config.SystemId), err.Error())
			return
		}
		resp.Diagnostics.AddError("error retrieving Config Info", err.Error())
		return
	}

	config.LoadApiData(ctx, systemConfig, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDeviceConfig) setClient(client *apstra.Client) {
	o.client = client
}
