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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterBlueprint{}
	_ datasourceWithSetClient            = &dataSourceDatacenterBlueprint{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterBlueprint{}
)

type dataSourceDatacenterBlueprint struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterBlueprint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *dataSourceDatacenterBlueprint) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterBlueprint) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source looks up summary details of a Datacenter Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.Blueprint{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterBlueprint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.Blueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiData *apstra.BlueprintStatus

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetBlueprintStatusByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Blueprint not found",
				fmt.Sprintf("Blueprint with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetBlueprintStatus(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Blueprint not found",
				fmt.Sprintf("Blueprint with ID %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Blueprint Status", err.Error())
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, apiData.Id.String())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// create new state object
	var state blueprint.Blueprint

	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.GetFabricSettings(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// -1 is a special case in the resource, not relevant
	// to the data source. Set these to null instead.
	if state.MaxEvpnRoutesCount.ValueInt64() < 0 {
		state.MaxEvpnRoutesCount = types.Int64Null()
	}
	if state.MaxExternalRoutesCount.ValueInt64() < 0 {
		state.MaxExternalRoutesCount = types.Int64Null()
	}
	if state.MaxFabricRoutesCount.ValueInt64() < 0 {
		state.MaxFabricRoutesCount = types.Int64Null()
	}
	if state.MaxMlagRoutesCount.ValueInt64() < 0 {
		state.MaxMlagRoutesCount = types.Int64Null()
	}

	state.GetDefaultRZParams(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceDatacenterBlueprint) setClient(client *apstra.Client) {
	o.client = client
}

func (o *dataSourceDatacenterBlueprint) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
