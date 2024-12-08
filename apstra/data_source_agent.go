package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	systemAgents "github.com/Juniper/terraform-provider-apstra/apstra/system_agents"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceAgent{}
	_ datasourceWithSetClient            = &dataSourceAgent{}
)

type dataSourceAgent struct {
	client *apstra.Client
}

func (o *dataSourceAgent) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (o *dataSourceAgent) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceAgent) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDevices + "This resource returns details of a Managed Device Agent.",
		Attributes:          systemAgents.ManagedDevice{}.DataSourceAttributes(),
	}
}

func (o *dataSourceAgent) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config systemAgents.ManagedDevice
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agent, err := o.client.GetSystemAgent(ctx, apstra.ObjectId(config.AgentId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("agent %s not found",
				config.AgentId), err.Error())
			return
		}
		resp.Diagnostics.AddError("error retrieving Agent", err.Error())
		return
	}

	config.LoadApiData(ctx, agent, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	config.GetDeviceKey(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get System info from Api
	systemInfo, err := o.client.GetSystemInfo(ctx, agent.Status.SystemId)
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching system info",
			fmt.Sprintf("Could not Read system info for %q - %s", agent.Status.SystemId, err.Error()),
		)
		return
	}

	config.LoadUserConfig(ctx, systemInfo.UserConfig, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceAgent) setClient(client *apstra.Client) {
	o.client = client
}
