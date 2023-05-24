package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterRoutingZones{}

type dataSourceDatacenterRoutingZones struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterRoutingZones) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zones"
}

func (o *dataSourceDatacenterRoutingZones) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterRoutingZones) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the IDs of Routing Zones within the specified Blueprint",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of Routing Zone IDs",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"filters": schema.SingleNestedAttribute{
				MarkdownDescription: "Routing Zone attributes used as filters",
				Optional:            true,
				Attributes:          blueprint.DatacenterRoutingZone{}.DataSourceFilterAttributes(),
			},
			"graph_query": schema.StringAttribute{
				MarkdownDescription: "The graph datastore query used to perform the lookup.",
				Computed:            true,
			},
		},
	}
}

func (o *dataSourceDatacenterRoutingZones) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	type routingZones struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		IDs         types.Set    `tfsdk:"ids"`
		Filters     types.Object `tfsdk:"filters"`
		Query       types.String `tfsdk:"graph_query"`
	}

	var config routingZones
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := &blueprint.DatacenterRoutingZone{}
	d := config.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := filters.Query("n_security_zone", "n_policy")

	queryResponse := new(struct {
		Items []struct {
			SecurityZone struct {
				Id string `json:"id"`
			} `json:"n_security_zone""`
			//Policy struct {
			//	Id string `json:"id"`
			//} `json:"n_policy""`
		} `json:"items"`
	})

	query.
		SetClient(o.client).
		SetBlueprintId(apstra.ObjectId(config.BlueprintId.ValueString())).
		SetBlueprintType(apstra.BlueprintTypeStaging)

	err := query.Do(ctx, queryResponse)
	if err != nil {
		resp.Diagnostics.AddError("error querying graph datastore", err.Error())
		return
	}

	ids := make([]attr.Value, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		ids[i] = types.StringValue(item.SecurityZone.Id)
	}

	config.IDs = types.SetValueMust(types.StringType, ids)
	config.Query = types.StringValue(query.String())

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
