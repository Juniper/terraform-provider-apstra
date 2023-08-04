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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

	query := filters.Query("n_virtual_network")
	queryResponse := new(struct {
		Items []struct {
			VirtualNetwork struct {
				Id          string `json:"id"`
				Ipv6Subnet  string `json:"ipv6_subnet"`          // todo verify match against filter
				Ipv6Gateway string `json:"virtual_gateway_ipv6"` // todo verify match against filter
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

	ids := make([]attr.Value, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		ids[i] = types.StringValue(item.VirtualNetwork.Id)
	}

	config.IDs = types.SetValueMust(types.StringType, ids)
	config.Query = types.StringValue(query.String())

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
