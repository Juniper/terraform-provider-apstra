package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"net"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterVirtualNetworks{}

type dataSourceDatacenterVirtualNetworks struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterVirtualNetworks) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_networks"
}

func (o *dataSourceDatacenterVirtualNetworks) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterVirtualNetworks) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the IDs of Virtual Networks within the specified Blueprint. " +
			"All of the `filter` attributes are optional.",
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
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Virtual Network attributes used as filter. At least " +
					"one filter attribute must be included when this attribute is used.",
				Optional:   true,
				Attributes: blueprint.DatacenterVirtualNetwork{}.DataSourceFilterAttributes(),
			},
			"graph_query": schema.StringAttribute{
				MarkdownDescription: "The graph datastore query based on `filter` used to " +
					"perform the lookup. Note that the `ipv6_subnet` and `ipv6_gateway` " +
					"attributes are never part of the graph query because IPv6 zero " +
					"compression rules make string matches unreliable.",
				Computed: true,
			},
		},
	}
}

func (o *dataSourceDatacenterVirtualNetworks) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	type virtualNetworks struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		IDs         types.Set    `tfsdk:"ids"`
		Filter      types.Object `tfsdk:"filter"`
		Query       types.String `tfsdk:"graph_query"`
	}

	var config virtualNetworks
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ids []attr.Value
	var query *apstra.MatchQuery // todo change to interface after SDK update
	bpId := apstra.ObjectId(config.BlueprintId.ValueString())
	if config.Filter.IsNull() {
		// just pull the VN IDs via API when no filter is specified
		ids = o.getAllVnIds(ctx, bpId, &resp.Diagnostics)
		config.IDs = types.SetValueMust(types.StringType, ids)
	} else {
		// use a graph query (and some IPv6 value matching)
		filter := blueprint.DatacenterVirtualNetwork{}
		resp.Diagnostics.Append(config.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		ids, query = o.getFilteredVnIds(ctx, bpId, filter, &resp.Diagnostics)
		config.IDs = types.SetValueMust(types.StringType, ids)
		config.Query = types.StringValue(query.String())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterVirtualNetworks) getAllVnIds(ctx context.Context, bpId apstra.ObjectId, diags *diag.Diagnostics) []attr.Value {
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, bpId)
	if err != nil {
		if utils.IsApstra404(err) {
			diags.AddError(fmt.Sprintf("blueprint %s not found", bpId), err.Error())
			return nil
		}
		diags.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, bpId), err.Error())
		return nil
	}

	ids, err := bpClient.ListAllVirtualNetworkIds(ctx)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("failed to list virtual networks in blueprint %s", bpId), err.Error())
		return nil
	}

	result := make([]attr.Value, len(ids))
	for i, id := range ids {
		result[i] = types.StringValue(id.String())
	}

	return result
}

// todo change returned query to interface after SDK update
func (o *dataSourceDatacenterVirtualNetworks) getFilteredVnIds(ctx context.Context, bpId apstra.ObjectId, filter blueprint.DatacenterVirtualNetwork, diags *diag.Diagnostics) ([]attr.Value, *apstra.MatchQuery) {
	query := filter.Query("n_virtual_network")
	queryResponse := new(struct {
		Items []struct {
			VirtualNetwork struct {
				Id          string  `json:"id"`
				Ipv6Subnet  *string `json:"ipv6_subnet"`
				Ipv6Gateway *string `json:"virtual_gateway_ipv6"`
			} `json:"n_virtual_network"`
		} `json:"items"`
	})

	// todo remove this type assertion when QEQuery is extended with new methods used below
	query2 := query.(*apstra.MatchQuery)
	query2.
		SetClient(o.client).
		SetBlueprintId(bpId).
		SetBlueprintType(apstra.BlueprintTypeStaging)

	err := query2.Do(ctx, queryResponse)
	if err != nil {
		diags.AddError("error querying graph datastore", err.Error())
		return nil, nil
	}

	// eliminate duplicate results
	idMap := make(map[string]bool)
	for i := len(queryResponse.Items) - 1; i >= 0; i-- {
		id := queryResponse.Items[i].VirtualNetwork.Id
		if idMap[id] {
			utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
			continue
		}
		idMap[id] = true
	}

	// extract the v6 subnet and gateway filters (if any)
	filterPath := path.Root("filter")
	v6SubnetFilter := filter.Ipv6Subnet(ctx, filterPath.AtName("ipv6_subnet"), diags)
	if diags.HasError() {
		return nil, nil
	}
	v6GatewayFilter := filter.Ipv6Gateway(ctx, filterPath.AtName("ipv6_virtual_gateway"), diags)
	if diags.HasError() {
		return nil, nil
	}

	if v6GatewayFilter != nil {
		// remove results which don't match the filter
		for i := len(queryResponse.Items) - 1; i >= 0; i-- {
			if queryResponse.Items[i].VirtualNetwork.Ipv6Gateway == nil {
				// Item has no gateway, so not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
			g := net.ParseIP(*queryResponse.Items[i].VirtualNetwork.Ipv6Gateway)
			if !v6GatewayFilter.Equal(g) {
				// Item's gateway is not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
		}
	}

	if v6SubnetFilter != nil {
		// remove results which don't match the filter
		for i := len(queryResponse.Items) - 1; i >= 0; i-- {
			if queryResponse.Items[i].VirtualNetwork.Ipv6Subnet == nil {
				// Item has no subnet, so not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
			_, s, err := net.ParseCIDR(*queryResponse.Items[i].VirtualNetwork.Ipv6Subnet)
			if err != nil {
				diags.AddError(fmt.Sprintf("failed parsing API response %q as CIDR",
					*queryResponse.Items[i].VirtualNetwork.Ipv6Subnet), err.Error())
				return nil, nil
			}
			if v6SubnetFilter.String() != s.String() {
				// Item's subnet is not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
		}
	}

	result := make([]attr.Value, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		result[i] = types.StringValue(item.VirtualNetwork.Id)
	}

	return result, query2
}
