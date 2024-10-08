package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure      = &dataSourceBlueprintIbaDashboards{}
	_ datasource.DataSourceWithValidateConfig = &dataSourceBlueprintIbaDashboards{}
	_ datasourceWithSetDcBpClientFunc         = &dataSourceBlueprintIbaDashboards{}
	_ datasourceWithSetClient                 = &dataSourceBlueprintIbaDashboards{}
)

type dataSourceBlueprintIbaDashboards struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	client          *apstra.Client
}

func (o *dataSourceBlueprintIbaDashboards) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_dashboards"
}

func (o *dataSourceBlueprintIbaDashboards) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintIbaDashboards) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source returns the ID numbers of all IBA Dashboards in a Blueprint.\n\n" +
			"*Note: Compatible only with Apstra " + compatibility.BpIbaDashboardOk.String() + "*",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID. " +
					"Used to identify the Blueprint that the IBA Dashboards belongs to.",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers of the IBA Dashboards in the Blueprint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceBlueprintIbaDashboards) ValidateConfig(_ context.Context, _ datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	// cannot proceed to config + api version validation if the provider has not been configured
	if o.client == nil {
		return
	}

	// only supported with Apstra 4.x
	if !compatibility.BpIbaDashboardOk.Check(version.Must(version.NewVersion(o.client.ApiVersion()))) {
		resp.Diagnostics.AddError(
			"Incompatible API version",
			"This data source is compatible only with Apstra "+compatibility.BpIbaDashboardOk.String(),
		)
	}
}

func (o *dataSourceBlueprintIbaDashboards) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Ids         types.Set    `tfsdk:"ids"`
	}

	// get the configuration
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

	ds, err := bp.GetAllIbaDashboards(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving IBA Dashboards", err.Error())
		return
	}

	ids := make([]attr.Value, len(ds))
	for i, j := range ds {
		ids[i] = types.StringValue(j.Id.String())
	}
	idSet := types.SetValueMust(types.StringType, ids)

	// create new state object
	state := struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Ids         types.Set    `tfsdk:"ids"`
	}{
		BlueprintId: config.BlueprintId,
		Ids:         idSet,
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceBlueprintIbaDashboards) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

// setClient is used for API version compatibility check only
func (o *dataSourceBlueprintIbaDashboards) setClient(client *apstra.Client) {
	o.client = client
}
