package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterSvis{}

type dataSourceDatacenterSvis struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterSvis) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_svis_by_vn"
}

func (o *dataSourceDatacenterSvis) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterSvis) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns a map of Sets of SVI info keyed by Virtual Network ID.",
		Attributes:          blueprint.DatacenterSvis{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterSvis) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// fetch config
	var config blueprint.DatacenterSvis
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
	sviMap, query := config.RunQuery(ctx, bpClient, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// fill the required values
	config.GraphQuery = types.StringValue(query.String())
	config.SviMap = utils.MapValueOrNull(ctx, types.SetType{ElemType: types.ObjectType{AttrTypes: blueprint.SviMapEntry{}.AttrTypes()}}, sviMap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
