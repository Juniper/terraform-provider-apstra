package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/iba"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprintIbaPredefinedProbe{}
var _ datasourceWithSetDcBpClientFunc = &dataSourceBlueprintIbaPredefinedProbe{}

type dataSourceBlueprintIbaPredefinedProbe struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceBlueprintIbaPredefinedProbe) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_predefined_probe"
}

func (o *dataSourceBlueprintIbaPredefinedProbe) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintIbaPredefinedProbe) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source provides details of a specific IBA Predefined Probe in a Blueprint.",
		Attributes:          iba.PredefinedProbe{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintIbaPredefinedProbe) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config iba.PredefinedProbe
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

	api, err := bp.GetIbaPredefinedProbeByName(ctx, config.Name.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"IBA Predefined Probe not found",
				fmt.Sprintf("IBA Predefined Probe with name %s not found", config.Name))
			return
		}
		resp.Diagnostics.AddAttributeError(
			path.Root("name"), "Failed reading IBA Predefined Probe", err.Error(),
		)
		return
	}
	config.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceBlueprintIbaPredefinedProbe) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
