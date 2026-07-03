package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure      = (*dataSourceDatacenterSwitchingZones)(nil)
	_ datasource.DataSourceWithValidateConfig = (*dataSourceDatacenterSwitchingZones)(nil)
	_ datasourceWithSetDcBpClientFunc         = (*dataSourceDatacenterSwitchingZones)(nil)
	_ datasourceWithSetClient                 = (*dataSourceDatacenterSwitchingZones)(nil)
)

type dataSourceDatacenterSwitchingZones struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterSwitchingZones) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_switching_zones"
}

func (o *dataSourceDatacenterSwitchingZones) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterSwitchingZones) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns the IDs of Switching Zones within the specified Blueprint. " +
			"All of the `filter` attributes are optional. Requires Apstra " + compatibility.SwitchingZoneOK.String() + ".",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of Switching Zone IDs",
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
					Attributes: blueprint.DatacenterSwitchingZone{}.DataSourceFilterAttributes(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(
							1,
							"name", "mac_vrf_name", "description", "service_type", "route_target", "tags",
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

func (o *dataSourceDatacenterSwitchingZones) ValidateConfig(_ context.Context, _ datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	// cannot proceed if the resource has not been configured
	if o.client == nil {
		return
	}

	apiVersion, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse API version %q", o.client.ApiVersion()), err.Error())
		return
	}

	if !compatibility.SwitchingZoneOK.Check(apiVersion) {
		resp.Diagnostics.AddError(
			"Resource requires Apstra "+compatibility.SwitchingZoneOK.String(),
			"Resource requires Apstra "+compatibility.SwitchingZoneOK.String(),
		)
		return
	}
}

func (o *dataSourceDatacenterSwitchingZones) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	type switchingZones struct {
		BlueprintId  types.String `tfsdk:"blueprint_id"`
		IDs          types.Set    `tfsdk:"ids"`
		Filters      types.List   `tfsdk:"filters"`
		GraphQueries types.List   `tfsdk:"graph_queries"`
	}

	// get the configuration
	var config switchingZones
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
		switchingZones, err := bp.GetSwitchingZones(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to fetch Switching Zones", err.Error())
			return
		}

		// collect the IDs
		ids := make([]attr.Value, len(switchingZones))
		for i, switchingZone := range switchingZones {
			if switchingZone.ID() == nil {
				resp.Diagnostics.AddError("failed fetching Switching Zones", fmt.Sprintf("Switching Zone at index %d has no ID", i))
				return
			}
			ids[i] = types.StringValue(*switchingZone.ID())
		}

		// set the state
		config.IDs = types.SetValueMust(types.StringType, ids)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// extract the supplied filters
	var filters []blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idMap := make(map[string]struct{})               // collect IDs here
	graphQueries := make([]attr.Value, len(filters)) // collect graph query strings here
	for i, filter := range filters {
		// prep and save the query
		query := filter.Query("n_switching_zone")
		graphQueries[i] = types.StringValue(query.String())

		// query response target
		queryResponse := new(struct {
			Items []struct {
				SwitchingZone struct {
					Id string `json:"id"`
				} `json:"n_switching_zone"`
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
			idMap[item.SwitchingZone.Id] = struct{}{}
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

func (o *dataSourceDatacenterSwitchingZones) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *dataSourceDatacenterSwitchingZones) setClient(client *apstra.Client) {
	o.client = client
}
