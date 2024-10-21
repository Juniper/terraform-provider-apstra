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
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterSecurityPolicies{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterSecurityPolicies{}
)

type dataSourceDatacenterSecurityPolicies struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterSecurityPolicies) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_security_policies"
}

func (o *dataSourceDatacenterSecurityPolicies) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterSecurityPolicies) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns the IDs of Security Policies within the " +
			"specified Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of Virtual Network IDs",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters used to select only desired node IDs. For a node " +
					"to match a filter, all specified attributes must match (each attribute within a " +
					"filter is AND-ed together). The returned node IDs represent the nodes matched by " +
					"all of the filters together (filters are OR-ed together).",
				Optional:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.DatacenterSecurityPolicy{}.DataSourceFilterAttributes(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(
							1,
							"name", "description", "enabled", "tags",
							"source_application_point_id", "destination_application_point_id"),
					},
				},
			},
			"graph_queries": schema.ListAttribute{
				MarkdownDescription: "The graph datastore query based on `filter` used to perform the lookup.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (o *dataSourceDatacenterSecurityPolicies) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	type virtualNetworks struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		IDs         types.Set    `tfsdk:"ids"`
		Filters     types.List   `tfsdk:"filters"`
		Queries     types.List   `tfsdk:"graph_queries"`
	}

	var config virtualNetworks
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
		// just pull the VN IDs via API when no filter is specified
		ids := o.getAllSpIds(ctx, bp, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// set the state
		config.IDs = types.SetValueMust(types.StringType, ids)
		config.Queries = types.ListNull(types.StringType)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// populate a list of filters from the `filters` configuration attribute
	var filters []blueprint.DatacenterSecurityPolicy
	resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ids, queries := o.getSpIdsWithFilters(ctx, bp, filters, &resp.Diagnostics)
	config.IDs = types.SetValueMust(types.StringType, ids)
	config.Queries = types.ListValueMust(types.StringType, queries)

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterSecurityPolicies) getAllSpIds(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) []attr.Value {
	policies, err := bp.GetAllPolicies(ctx)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("failed to retrieve security policies in blueprint %s", bp.Id()), err.Error())
		return nil
	}

	result := make([]attr.Value, len(policies))
	for i, policy := range policies {
		result[i] = types.StringValue(policy.Id.String())
	}

	return result
}

func (o *dataSourceDatacenterSecurityPolicies) getSpIdsWithFilters(ctx context.Context, bp *apstra.TwoStageL3ClosClient, filters []blueprint.DatacenterSecurityPolicy, diags *diag.Diagnostics) ([]attr.Value, []attr.Value) {
	queries := make([]attr.Value, len(filters))
	resultMap := make(map[string]bool)
	for i, filter := range filters {
		ids, query := o.getSpIdsWithFilter(ctx, bp, filter, diags)
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

func (o *dataSourceDatacenterSecurityPolicies) getSpIdsWithFilter(ctx context.Context, bp *apstra.TwoStageL3ClosClient, filter blueprint.DatacenterSecurityPolicy, diags *diag.Diagnostics) ([]string, apstra.QEQuery) {
	query := filter.Query("n_security_policy")
	queryResponse := new(struct {
		Items []struct {
			VirtualNetwork struct {
				Id string `json:"id"`
			} `json:"n_security_policy"`
		} `json:"items"`
	})

	// todo remove this type assertion when QEQuery is extended with new methods used below
	query.(*apstra.MatchQuery).SetClient(bp.Client())
	query.(*apstra.MatchQuery).SetBlueprintId(bp.Id())
	query.(*apstra.MatchQuery).SetBlueprintType(apstra.BlueprintTypeStaging)
	err := query.Do(ctx, queryResponse)
	if err != nil {
		diags.AddError("error querying graph datastore", err.Error())
		return nil, nil
	}

	result := make([]string, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		result[i] = item.VirtualNetwork.Id
	}

	return result, query
}

func (o *dataSourceDatacenterSecurityPolicies) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
