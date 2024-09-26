package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/iba"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithConfigure      = &dataSourceBlueprintIbaWidget{}
	_ datasource.DataSourceWithValidateConfig = &dataSourceBlueprintIbaWidget{}
	_ datasourceWithSetDcBpClientFunc         = &dataSourceBlueprintIbaWidget{}
	_ datasourceWithSetClient                 = &dataSourceBlueprintIbaWidget{}
)

type dataSourceBlueprintIbaWidget struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	client          *apstra.Client
}

func (o *dataSourceBlueprintIbaWidget) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_widget"
}

func (o *dataSourceBlueprintIbaWidget) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintIbaWidget) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source provides details of a specific IBA Widget in a Blueprint." +
			"\n\n" +
			"At least one optional attribute is required.\n\n" +
			"*Note: Compatible only with Apstra " + compatibility.BpIbaWidgetOk.String() + "*",

		Attributes: iba.Widget{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintIbaWidget) ValidateConfig(_ context.Context, _ datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	// cannot proceed to api version validation if the provider has not been configured
	if o.client == nil {
		return
	}

	// only supported with Apstra 4.x
	if !compatibility.BpIbaWidgetOk.Check(version.Must(version.NewVersion(o.client.ApiVersion()))) {
		resp.Diagnostics.AddError(
			"Incompatible API version",
			"This data source is compatible only with Apstra "+compatibility.BpIbaWidgetOk.String(),
		)
	}
}

func (o *dataSourceBlueprintIbaWidget) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config iba.Widget
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

	var api *apstra.IbaWidget
	switch {
	case !config.Name.IsNull():
		api, err = bp.GetIbaWidgetByLabel(ctx, config.Name.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"IBA widget not found",
					fmt.Sprintf("IBA Widget with name %s not found", config.Name))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading IBA Widget", err.Error(),
			)
			return
		}
	case !config.Id.IsNull():
		api, err = bp.GetIbaWidget(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"Widget not found",
					fmt.Sprintf("Widget with ID %s not found", config.Id))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading IBA Widget", err.Error(),
			)
			return
		}
	}

	config.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceBlueprintIbaWidget) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

// setClient is used for API version compatibility check only
func (o *dataSourceBlueprintIbaWidget) setClient(client *apstra.Client) {
	o.client = client
}
