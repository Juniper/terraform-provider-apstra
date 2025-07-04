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
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterInterconnectDomainGateways{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterInterconnectDomainGateways{}
)

type dataSourceDatacenterInterconnectDomainGateways struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterInterconnectDomainGateways) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain_gateways"
}

func (o *dataSourceDatacenterInterconnectDomainGateways) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterInterconnectDomainGateways) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns Graph DB node IDs of Interconnect Domain Gateways within a Blueprint.\n\n" +
			"Optional `filters` can be used to select only interesting Interconnect Domain Gateways.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint to search.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters used to select only desired Interconnect Domain Gateways. " +
					"To match a filter, all specified attributes must match (each attribute within a " +
					"filter is AND-ed together). The returned IDs represent the gateways matched by " +
					"all of the filters together (filters are OR-ed together).",
				Optional:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.InterconnectDomainGateway{}.DataSourceAttributesAsFilter(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(
							1,
							"id", "name", "ip_address", "asn", "ttl",
							"keepalive_time", "hold_time", "interconnect_domain_id",
							"local_gateway_nodes", "password",
						),
					},
				},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "IDs of matching `interconnect_domain_gateway` Graph DB nodes.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceDatacenterInterconnectDomainGateways) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
	var filters []blueprint.InterconnectDomainGateway
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

	// collect all remote gateways in the blueprint
	remoteGateways, err := bp.GetAllRemoteGateways(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch external gateways", err.Error())
		return
	}

	// eliminate remote gateways which do not belong to an interconnect group
	for i := len(remoteGateways) - 1; i >= 0; i-- { // loop backwards
		if remoteGateways[i].Data.EvpnInterconnectGroupId == nil {
			remoteGateways[i] = remoteGateways[len(remoteGateways)-1] // copy last item to position i
			remoteGateways = remoteGateways[:len(remoteGateways)-1]   // delete last item
		}
	}

	// Did the user send any filters?
	if len(filters) == 0 { // no filter shortcut! return all IDs without inspection

		// collect the IDs into config.Ids
		ids := make([]attr.Value, len(remoteGateways))
		for i, remoteGateway := range remoteGateways {
			ids[i] = types.StringValue(remoteGateway.Id.String())
		}
		config.Ids = types.SetValueMust(types.StringType, ids)

		// set the state
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// determine whether we need to extract protocol passwords from the graph
	filterOnProtocolPassword := false
	for _, filter := range filters {
		if !filter.Password.IsNull() {
			filterOnProtocolPassword = true // we need to extract protocol passwords from the graph
			break
		}
	}

	// extract the API response items so that they can be filtered
	var candidates []blueprint.InterconnectDomainGateway
	for _, remoteGateway := range remoteGateways {
		interconnectDomainGateway := blueprint.InterconnectDomainGateway{Id: types.StringValue(remoteGateway.Id.String())}
		interconnectDomainGateway.LoadApiData(ctx, remoteGateway.Data, &resp.Diagnostics)
		if filterOnProtocolPassword {
			interconnectDomainGateway.ReadProtocolPassword(ctx, bp, &resp.Diagnostics)
		}
		candidates = append(candidates, interconnectDomainGateway)
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
				break // we found a match - don't need to check remaining filters
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

func (o *dataSourceDatacenterInterconnectDomainGateways) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
