package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterExternalGateways{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterExternalGateways{}
)

type dataSourceDatacenterExternalGateways struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterExternalGateways) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_external_gateways"
}

func (o *dataSourceDatacenterExternalGateways) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterExternalGateways) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns Graph DB node IDs of DCI External Gateways within a Blueprint.\n\n" +
			"Optional `filters` can be used to select only interesting External Gateways.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint to search.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters used to select only desired External Gateways. " +
					"To match a filter, all specified attributes must match (each attribute within a " +
					"filter is AND-ed together). The returned IDs represent the gateways matched by " +
					"all of the filters together (filters are OR-ed together).",
				Optional:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.ExternalGateway{}.DataSourceAttributesAsFilter(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(
							1,
							"id", "name", "ip_address", "asn", "ttl",
							"keepalive_time", "hold_time", "evpn_route_types",
							"local_gateway_nodes", "password",
						),
					},
				},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "IDs of matching `routing_policy` Graph DB nodes.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceDatacenterExternalGateways) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Filters     types.List   `tfsdk:"filters"`
		Ids         types.Set    `tfsdk:"ids"`
	}

	// Retrieve values from config.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// extract filters from the config
	var filters []blueprint.ExternalGateway
	resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
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

	// collect all external gateways in the blueprint
	remoteGateways, err := bp.GetAllRemoteGateways(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch external gateways", err.Error())
		return
	}

	// Did the user send any filters?
	if len(filters) == 0 { // no filter shortcut! return all IDs without inspection

		// collect the IDs into config.Ids
		ids := make([]attr.Value, 0, len(remoteGateways))
		for _, remoteGateway := range remoteGateways {
			ids = append(ids, types.StringValue(remoteGateway.Id.String()))
		}
		config.Ids = types.SetValueMust(types.StringType, ids)

		// set the state
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// extract the API response items so that they can be filtered
	candidates := make([]blueprint.ExternalGateway, 0, len(remoteGateways))
	for _, remoteGateway := range remoteGateways {
		if remoteGateway.Data.EvpnInterconnectGroupId != nil {
			continue // this is an interconnect domain gateway, not an external gateway
		}

		externalGateway := blueprint.ExternalGateway{Id: types.StringValue(remoteGateway.Id.String())}
		externalGateway.LoadApiData(ctx, remoteGateway.Data, &resp.Diagnostics)
		externalGateway.ReadProtocolPassword(ctx, bp, &resp.Diagnostics)
		candidates = append(candidates, externalGateway)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// collect ids of candidates which match any filter
	var ids []attr.Value
	for _, candidate := range candidates { // loop over candidates
		for _, filter := range filters { // loop over filters
			if filter.FilterMatch(ctx, &candidate, &resp.Diagnostics) {
				ids = append(ids, candidate.Id)
				break
			}
		}
	}

	// pack the IDs into config.Ids
	config.Ids = utils.SetValueOrNull(ctx, types.StringType, ids, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterExternalGateways) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
