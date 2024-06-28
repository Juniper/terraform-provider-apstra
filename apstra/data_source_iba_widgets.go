package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprintIbaWidgets{}
var _ datasourceWithSetDcBpClientFunc = &dataSourceBlueprintIbaWidgets{}

type dataSourceBlueprintIbaWidgets struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceBlueprintIbaWidgets) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_widgets"
}

func (o *dataSourceBlueprintIbaWidgets) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintIbaWidgets) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source returns the IDs of the IBA Widgets in a Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID. " +
					"Used to identify the Blueprint that the IBA Widgets belongs to.",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers representing the IBA Widgets in the blueprint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceBlueprintIbaWidgets) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	ws, err := bp.GetAllIbaWidgets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving IBA Widgets", err.Error())
		return
	}

	ids := make([]attr.Value, len(ws))
	for i, j := range ws {
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

func (o *dataSourceBlueprintIbaWidgets) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
