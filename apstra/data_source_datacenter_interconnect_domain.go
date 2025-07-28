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
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterInterconnectDomain{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterInterconnectDomain{}
)

type dataSourceDatacenterInterconnectDomain struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterInterconnectDomain) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain"
}

func (o *dataSourceDatacenterInterconnectDomain) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterInterconnectDomain) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source provides details of a specific Interconnect " +
			"Domain within the given Blueprint.\n\nAt least one optional attribute is required.",
		Attributes: blueprint.InterconnectDomain{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterInterconnectDomain) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.InterconnectDomain
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, config.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, config.BlueprintId), err.Error())
		return
	}

	var api *apstra.EvpnInterconnectGroup
	switch {
	case !config.Name.IsNull():
		api, err = bp.GetEvpnInterconnectGroupByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Interconnect Domain not found",
				fmt.Sprintf("Interconnect Domain with label %s not found", config.Name))
			return
		}
	case !config.Id.IsNull():
		api, err = bp.GetEvpnInterconnectGroup(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Interconnect Domain not found",
				fmt.Sprintf("Interconnect Domain with ID %s not found", config.Id))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed reading Interconnect Domain", err.Error())
		return
	}

	// add ID value in case the caller passed us the name instead
	if !utils.HasValue(config.Id) {
		config.Id = types.StringValue(api.Id.String())
	}

	// load the remaining details
	config.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterInterconnectDomain) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
