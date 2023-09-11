package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceInterfacesBySystem{}

type dataSourceInterfacesBySystem struct {
	client *apstra.Client
}

func (o *dataSourceInterfacesBySystem) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interfaces_by_system"
}

func (o *dataSourceInterfacesBySystem) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceInterfacesBySystem) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns a map of Interface IDs keyed by Interface Name for the given System ID.",
		Attributes:          blueprint.InterfacesBySystem{}.DataSourceAttributes(),
	}
}

func (o *dataSourceInterfacesBySystem) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// fetch config
	var config blueprint.InterfacesBySystem
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a blueprint client
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

	// prepare and execute a graph query
	interfaces, query := config.RunQuery(ctx, bpClient, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// fill the required values
	config.GraphQuery = types.StringValue(query.String())
	config.IfMap = utils.MapValueOrNull(ctx, types.StringType, interfaces, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
