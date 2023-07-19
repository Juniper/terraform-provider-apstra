package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterRoutingPolicy{}

type dataSourceDatacenterRoutingPolicy struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterRoutingPolicy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_policy"
}

func (o *dataSourceDatacenterRoutingPolicy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterRoutingPolicy) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource returns details of a Datacenter Routing Policy.",
		Attributes:          blueprint.DatacenterRoutingPolicy{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterRoutingPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	// Retrieve values from config.
	var config blueprint.DatacenterRoutingPolicy
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

	var api *apstra.DcRoutingPolicy
	switch {
	case !config.Id.IsNull():
		api, err = bpClient.GetRoutingPolicy(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"Routing Policy not found",
					fmt.Sprintf("Routing Policy with ID %s not found", config.Id))
				return
			}
			resp.Diagnostics.AddError(
				"Failed reading Routing Policy", err.Error(),
			)
		}
	case !config.Name.IsNull():
		api, err = bpClient.GetRoutingPolicyByName(ctx, config.Name.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"Routing Policy not found",
					fmt.Sprintf("Routing Policy with Name %s not found", config.Name))
				return
			}
			resp.Diagnostics.AddError(
				"Failed reading Routing Policy", err.Error(),
			)
		}
	}

	config.Id = types.StringValue(api.Id.String())
	config.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}