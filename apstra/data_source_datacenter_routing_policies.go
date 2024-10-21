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

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterRoutingPolicies{}
var _ datasourceWithSetDcBpClientFunc = &dataSourceDatacenterRoutingPolicies{}

type dataSourceDatacenterRoutingPolicies struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterRoutingPolicies) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_policies"
}

func (o *dataSourceDatacenterRoutingPolicies) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterRoutingPolicies) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns Graph DB node IDs of *routing_policy* nodes within a Blueprint.\n\n" +
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
					Attributes: blueprint.DatacenterRoutingPolicy{}.DataSourceAttributesAsFilter(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(
							1,
							"id", "name", "description", "import_policy",
							"export_policy", "expect_default_ipv4", "expect_default_ipv6",
							"aggregate_prefixes", "extra_imports", "extra_exports",
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

func (o *dataSourceDatacenterRoutingPolicies) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Filters     types.List   `tfsdk:"filters"`
		Ids         types.Set    `tfsdk:"ids"`
	}

	// retrieve values from config
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// extract routing policy filters from the config
	var filters []blueprint.DatacenterRoutingPolicy
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

	// collect all routing policies in the blueprint
	apiResponse, err := bp.GetAllRoutingPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to retrieve routing policies", err.Error())
		return
	}

	// Did the user send any filters?
	if len(filters) == 0 { // no filter shortcut! return all routing policy IDs without inspection

		// collect the IDs into config.Ids
		ids := make([]attr.Value, len(apiResponse))
		for i, routingPolicy := range apiResponse {
			ids[i] = types.StringValue(routingPolicy.Id.String())
		}
		config.Ids = types.SetValueMust(types.StringType, ids)

		// set the state
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// extract the API response items so that they can be filtered
	candidates := make([]blueprint.DatacenterRoutingPolicy, len(apiResponse))
	for i := range apiResponse {
		routingPolicy := blueprint.DatacenterRoutingPolicy{Id: types.StringValue(apiResponse[i].Id.String())}
		routingPolicy.LoadApiData(ctx, apiResponse[i].Data, &resp.Diagnostics)
		candidates[i] = routingPolicy
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

func (o *dataSourceDatacenterRoutingPolicies) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
