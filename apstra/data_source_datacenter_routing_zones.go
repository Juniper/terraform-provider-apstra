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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterRoutingZones{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterRoutingZones{}
)

type dataSourceDatacenterRoutingZones struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterRoutingZones) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zones"
}

func (o *dataSourceDatacenterRoutingZones) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterRoutingZones) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns the IDs of Routing Zones within the specified Blueprint. " +
			"All of the `filter` attributes are optional.",
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
			"filter": schema.SingleNestedAttribute{
				DeprecationMessage: "The `filter` attribute is deprecated and will be removed in a future " +
					"release. Please migrate your configuration to use `filters` instead.",
				MarkdownDescription: "Routing Zone attributes used as a filter",
				Optional:            true,
				Attributes:          blueprint.DatacenterRoutingZone{}.DataSourceFilterAttributes(),
				Validators: []validator.Object{
					apstravalidator.AtMostNOf(1,
						path.MatchRelative(),
						path.MatchRoot("filters"),
					),
					apstravalidator.AtLeastNAttributes(
						1,
						"name", "vrf_name", "vlan_id", "vni", "dhcp_servers", "routing_policy_id",
						"import_route_targets", "export_route_targets", "junos_evpn_irb_mode",
					),
				},
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters used to select only desired node IDs. For a node " +
					"to match a filter, all specified attributes must match (each attribute within a " +
					"filter is AND-ed together). The returned node IDs represent the nodes matched by " +
					"all of the filters together (filters are OR-ed together).",
				Optional:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.DatacenterRoutingZone{}.DataSourceFilterAttributes(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(
							1,
							"name", "vrf_name", "vlan_id", "vni", "dhcp_servers", "routing_policy_id",
							"import_route_targets", "export_route_targets", "junos_evpn_irb_mode",
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

func (o *dataSourceDatacenterRoutingZones) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	type routingZones struct {
		BlueprintId  types.String `tfsdk:"blueprint_id"`
		IDs          types.Set    `tfsdk:"ids"`
		Filter       types.Object `tfsdk:"filter"`
		Filters      types.List   `tfsdk:"filters"`
		GraphQueries types.List   `tfsdk:"graph_queries"`
	}

	// get the configuration
	var config routingZones
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
	if config.Filter.IsNull() && config.Filters.IsNull() {
		securityZones, err := bp.GetSecurityZones(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to fetch Routing Zones", err.Error())
			return
		}

		// collect the IDs
		ids := make([]attr.Value, len(securityZones))
		for i, securityZone := range securityZones {
			if securityZone.ID() == nil {
				resp.Diagnostics.AddError("failed fetching Routing Zones", fmt.Sprintf("Routing Zone at index %d has no ID", i))
				return
			}
			ids[i] = types.StringValue(*securityZone.ID())
		}

		// set the state
		config.IDs = types.SetValueMust(types.StringType, ids)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// extract the supplied filters
	var filters []blueprint.DatacenterRoutingZone
	switch {
	case !config.Filter.IsNull():
		filters = make([]blueprint.DatacenterRoutingZone, 1)
		resp.Diagnostics.Append(config.Filter.As(ctx, &filters[0], basetypes.ObjectAsOptions{})...)
	case !config.Filters.IsNull():
		resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	idMap := make(map[string]struct{})               // collect IDs here
	graphQueries := make([]attr.Value, len(filters)) // collect graph query strings here
	for i, filter := range filters {
		// prep and save the query
		query := filter.Query("n_security_zone")
		graphQueries[i] = types.StringValue(query.String())

		// query response target
		queryResponse := new(struct {
			Items []struct {
				SecurityZone struct {
					Id string `json:"id"`
				} `json:"n_security_zone"`
			} `json:"items"`
		})

		// perform the query
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
			idMap[item.SecurityZone.Id] = struct{}{}
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

func (o *dataSourceDatacenterRoutingZones) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
