package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterConnectivityTemplatesStatus{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterConnectivityTemplatesStatus{}
)

type dataSourceDatacenterConnectivityTemplatesStatus struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterConnectivityTemplatesStatus) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_connectivity_templates_status"
}

func (o *dataSourceDatacenterConnectivityTemplatesStatus) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterConnectivityTemplatesStatus) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource returns a map detailing status of Connectivity Templates within a Datacenter Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Blueprint",
				Required:            true,
			},
			"connectivity_templates": schema.MapNestedAttribute{
				MarkdownDescription: "Connectivity Template status details, keyed by ID",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.ConnectivityTemplateStatus{}.DataSourceAttributes(),
				},
			},
		},
	}
}

func (o *dataSourceDatacenterConnectivityTemplatesStatus) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config struct {
		BlueprintId           types.String `tfsdk:"blueprint_id"`
		ConnectivityTemplates types.Map    `tfsdk:"connectivity_templates"`
	}
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

	// collect CT status info from the API
	apiData, err := bp.GetAllConnectivityTemplateStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Connectivity Template statuses", err.Error())
		return
	}

	// load CT status info
	var ctStatus blueprint.ConnectivityTemplateStatus
	ctStatusMap := make(map[string]blueprint.ConnectivityTemplateStatus)
	for k, v := range apiData {
		ctStatus.LoadApiData(ctx, v, &resp.Diagnostics)
		ctStatusMap[k.String()] = ctStatus
	}
	config.ConnectivityTemplates = value.MapOrNull(ctx, types.ObjectType{AttrTypes: ctStatus.AttrTypes()}, ctStatusMap, &resp.Diagnostics)

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterConnectivityTemplatesStatus) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
