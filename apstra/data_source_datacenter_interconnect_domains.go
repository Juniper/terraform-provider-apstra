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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterInterconnectDomains{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterInterconnectDomains{}
)

type dataSourceDatacenterInterconnectDomains struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterInterconnectDomains) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domains"
}

func (o *dataSourceDatacenterInterconnectDomains) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterInterconnectDomains) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns Graph DB node IDs of " +
			"`evpn_interconnect_group` (Interconnect Domain) nodes within the given Blueprint. Note that creation of " +
			"multiple Interconnect Domain resources is not currently supported.\n\n" +
			"Optional `filters` can be used to select only interesting nodes.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint to search.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters used to select only desired node IDs. For a node" +
					"to match a filter, all specified attributes must match (each attribute within a " +
					"filter is AND-ed together). The returned node IDs represent the nodes matched by " +
					"all of the filters together (filters are OR-ed together).",
				Optional:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.InterconnectDomain{}.DataSourceAttributesAsFilter(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(1, "id", "name", "route_target", "esi_mac"),
					},
				},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "IDs of matching `evpn_interconnect_group` Graph DB nodes.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"graph_queries": schema.ListAttribute{
				MarkdownDescription: "The graph datastore query based on `filter` used to perform the lookup.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceDatacenterInterconnectDomains) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Ids         types.Set    `tfsdk:"ids"`
		Filters     types.List   `tfsdk:"filters"`
		Queries     types.List   `tfsdk:"graph_queries"`
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

	if config.Filters.IsNull() {
		// just pull the IDs via API when no filter is specified
		ids := o.getAllIds(ctx, bp, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// set the state
		config.Ids = types.SetValueMust(types.StringType, ids)
		config.Queries = types.ListNull(types.StringType)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// populate a list of filters from the `filters` configuration attribute
	var filters []blueprint.InterconnectDomain
	resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ids, queries := o.getIdsWithFilters(ctx, bp, filters, &resp.Diagnostics)
	config.Ids = types.SetValueMust(types.StringType, ids)
	config.Queries = types.ListValueMust(types.StringType, queries)

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterInterconnectDomains) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *dataSourceDatacenterInterconnectDomains) getAllIds(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) []attr.Value {
	all, err := bp.GetAllEvpnInterconnectGroups(ctx)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("failed to retrieve Interconnect Domains in Blueprint %s", bp.Id()), err.Error())
		return nil
	}

	result := make([]attr.Value, len(all))
	for i, each := range all {
		result[i] = types.StringValue(each.Id.String())
	}

	return result
}

func (o *dataSourceDatacenterInterconnectDomains) getIdsWithFilters(ctx context.Context, bp *apstra.TwoStageL3ClosClient, filters []blueprint.InterconnectDomain, diags *diag.Diagnostics) ([]attr.Value, []attr.Value) {
	queries := make([]attr.Value, len(filters))
	resultMap := make(map[string]bool)
	for i, filter := range filters {
		ids, query := o.getIdsWithFilter(ctx, bp, filter, diags)
		if diags.HasError() {
			return nil, nil
		}

		queries[i] = types.StringValue(query.String())
		for _, id := range ids {
			resultMap[id] = true
		}
	}

	ids := make([]attr.Value, len(resultMap))
	var i int
	for id := range resultMap {
		ids[i] = types.StringValue(id)
		i++
	}

	return ids, queries
}

func (o *dataSourceDatacenterInterconnectDomains) getIdsWithFilter(ctx context.Context, bp *apstra.TwoStageL3ClosClient, filter blueprint.InterconnectDomain, diags *diag.Diagnostics) ([]string, apstra.QEQuery) {
	query := filter.Query("n_evpn_interconnect_group")
	queryResponse := new(struct {
		Items []struct {
			EvpnInterconnectGroup struct {
				Id string `json:"id"`
			} `json:"n_evpn_interconnect_group"`
		} `json:"items"`
	})

	// todo remove this type assertion when QEQuery is extended with new methods used below
	query.(*apstra.PathQuery).SetClient(bp.Client())
	query.(*apstra.PathQuery).SetBlueprintId(bp.Id())
	query.(*apstra.PathQuery).SetBlueprintType(apstra.BlueprintTypeStaging)
	err := query.Do(ctx, queryResponse)
	if err != nil {
		diags.AddError("error querying graph datastore", err.Error())
		return nil, nil
	}

	result := make([]string, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		result[i] = item.EvpnInterconnectGroup.Id
	}

	return result, query
}
