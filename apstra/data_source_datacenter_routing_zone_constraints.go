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
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterRoutingZoneConstraints{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterRoutingZoneConstraints{}
)

type dataSourceDatacenterRoutingZoneConstraints struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterRoutingZoneConstraints) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zone_constraints"
}

func (o *dataSourceDatacenterRoutingZoneConstraints) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterRoutingZoneConstraints) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns the IDs of Routing Zone Constraints within the specified Blueprint. " +
			"All of the `filter` attributes are optional.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of Routing Zone Constraint IDs",
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
					Attributes: blueprint.DatacenterRoutingZoneConstraint{}.DataSourceFilterAttributes(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(
							1,
							"name", "max_count_constraint", "routing_zones_list_constraint", "constraints",
						),
					},
				},
			},
			"graph_queries": schema.ListAttribute{
				MarkdownDescription: "Graph datastore queries which performed the lookup based on supplied filters.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceDatacenterRoutingZoneConstraints) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	type routingZoneConstraints struct {
		BlueprintId  types.String `tfsdk:"blueprint_id"`
		IDs          types.Set    `tfsdk:"ids"`
		Filters      types.List   `tfsdk:"filters"`
		GraphQueries types.List   `tfsdk:"graph_queries"`
	}

	// get the configuration
	var config routingZoneConstraints
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

	// If no filters supplied, we can just fetch IDs via the API
	if config.Filters.IsNull() {
		allRoutingZoneConstraints, err := bp.GetAllRoutingZoneConstraints(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to fetch routing zone constraints", err.Error())
			return
		}

		// collect the IDs
		ids := make([]attr.Value, len(allRoutingZoneConstraints))
		for i, routingZoneConstraint := range allRoutingZoneConstraints {
			ids[i] = types.StringValue(routingZoneConstraint.Id.String())
		}

		// set the state
		config.IDs = types.SetValueMust(types.StringType, ids)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// extract the supplied filters
	var filters []blueprint.DatacenterRoutingZoneConstraint
	resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idMap := make(map[string]struct{})               // collect IDs here
	graphQueries := make([]attr.Value, len(filters)) // collect graph query strings here
	for i, filter := range filters {
		// prep a query
		query := filter.Query(ctx, "n_routing_zone_constraint", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// save the query
		graphQueries[i] = types.StringValue(query.String())

		// query response target
		queryResponse := new(struct {
			Items []struct {
				RoutingZoneConstraint struct {
					Id string `json:"id"`
				} `json:"n_routing_zone_constraint"`
			} `json:"items"`
		})

		// run the query
		query.
			SetClient(bp.Client()).
			SetBlueprintId(apstra.ObjectId(config.BlueprintId.ValueString())).
			SetBlueprintType(apstra.BlueprintTypeStaging)
		err = query.Do(ctx, queryResponse)
		if err != nil {
			resp.Diagnostics.AddError("error querying graph datastore", err.Error())
			return
		}

		// save the IDs into idMap
		for _, item := range queryResponse.Items {
			idMap[item.RoutingZoneConstraint.Id] = struct{}{}
		}
	}

	// pull the IDs out of the map
	ids := make([]attr.Value, len(idMap))
	var i int
	for id := range idMap {
		ids[i] = types.StringValue(id)
		i++
	}

	// set the state
	config.IDs = types.SetValueMust(types.StringType, ids)
	config.GraphQueries = types.ListValueMust(types.StringType, graphQueries)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterRoutingZoneConstraints) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
