package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"net"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
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
			"All of the `filters` attributes are optional.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of Virtual Neteork IDs",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"filters": schema.SingleNestedAttribute{
				MarkdownDescription: "Virtual Network attributes used as filters",
				Optional:            true,
				Attributes:          blueprint.DatacenterVirtualNetwork{}.DataSourceFilterAttributes(),
			},
			"graph_query": schema.StringAttribute{
				MarkdownDescription: "The graph datastore query used to perform the lookup.",
				Computed:            true,
			},
		},
	}
}

func (o *dataSourceDatacenterVirtualNetworks) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	type virtualNetworks struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		IDs         types.Set    `tfsdk:"ids"`
		Filters     types.Object `tfsdk:"filters"`
		Query       types.String `tfsdk:"graph_query"`
	}

	var config virtualNetworks
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Filters.IsNull() {
		// just pull the VN IDs via API when no filters are specified
		bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(config.BlueprintId.ValueString()))
		if err != nil {
			var ace apstra.ApstraClientErr
			if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
				resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found",
					config.BlueprintId), err.Error())
				return
			}
			resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, config.BlueprintId), err.Error())
			return
		}

		ids, err := bpClient.ListAllVirtualNetworkIds(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("failed to list virtual networks in blueprint %s", config.BlueprintId), err.Error())
			return
		}

		// load the result
		config.IDs = utils.SetValueOrNull(ctx, types.StringType, ids, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// set state
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)

		return
	}

	// if we got here, the user specified some filter attributes
	filters := blueprint.DatacenterVirtualNetwork{}
	if !config.Filters.IsNull() {
		resp.Diagnostics.Append(config.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// pull out IPv6 network objects for later use (we can't let the graph db do
	// string compare on these because of possible "::" expansion weirdness.
	var v6Gateway net.IP
	if !filters.IPv6Gateway.IsNull() {
		v6Gateway = net.ParseIP(filters.IPv6Gateway.ValueString())
		if v6Gateway == nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("filters").AtMapKey("ipv6_virtual_gateway"),
				fmt.Sprintf("failed to parse 'ipv6_virtual_gateway' value %s", filters.IPv4Gateway),
				"result: `nil`")
		}
	}

	// pull out IPv6 network objects for later use (we can't let the graph db do
	// string compare on these because of possible "::" expansion weirdness.
	var v6Subnet *net.IPNet
	if !filters.IPv6Subnet.IsNull() {
		var err error
		_, v6Subnet, err = net.ParseCIDR(filters.IPv6Subnet.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("filters").AtMapKey("ipv6_subnet"),
				fmt.Sprintf("failed to parse 'ipv6_subnet' value %s", filters.IPv6Subnet),
				err.Error())
		}
	}

	query := filters.Query("n_virtual_network")
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
		SetBlueprintId(apstra.ObjectId(config.BlueprintId.ValueString())).
		SetBlueprintType(apstra.BlueprintTypeStaging)

	err := query2.Do(ctx, queryResponse)
	if err != nil {
		resp.Diagnostics.AddError("error querying graph datastore", err.Error())
		return
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

	// check IPv6 fields because these cannot be compared by Apstra's query
	// feature (string match might have unexpected results)
	if v6Gateway != nil {
		for i := len(queryResponse.Items) - 1; i >= 0; i-- {
			itemGateway := queryResponse.Items[i].VirtualNetwork.Ipv6Gateway
			if itemGateway == nil {
				// Item has no gateway, so not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
			g := net.ParseIP(*queryResponse.Items[i].VirtualNetwork.Ipv6Gateway)
			if !v6Gateway.Equal(g) {
				// Item's gateway is not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
		}
	}

	// check IPv6 fields because these cannot be compared by Apstra's query
	// feature (string match might have unexpected results)
	if v6Subnet != nil {
		for i := len(queryResponse.Items) - 1; i >= 0; i-- {
			itemSubnet := queryResponse.Items[i].VirtualNetwork.Ipv6Subnet
			if itemSubnet == nil {
				// Item has no subnet, so not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
			_, s, err := net.ParseCIDR(*queryResponse.Items[i].VirtualNetwork.Ipv6Subnet)
			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("failed parsing API response %q as CIDR", itemSubnet), err.Error())
				return
			}
			if v6Subnet.String() != s.String() {
				// Item's subnet is not a match. Drop it.
				utils.SliceDeleteUnOrdered(i, &queryResponse.Items)
				continue
			}
		}
	}

	ids := make([]attr.Value, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		ids[i] = types.StringValue(item.VirtualNetwork.Id)
	}

	config.IDs = types.SetValueMust(types.StringType, ids)
	config.Query = types.StringValue(query.String())

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
