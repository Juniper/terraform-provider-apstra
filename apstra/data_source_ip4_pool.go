package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceIp4Pool{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceIp4Pool{}

type dataSourceIp4Pool struct {
	client *goapstra.Client
}

func (o *dataSourceIp4Pool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_pool"
}

func (o *dataSourceIp4Pool) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataDetail,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *dataSourceIp4Pool) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details of a single IPv4 Resource Pool. It is incumbent upon " +
			"the user to set enough optional criteria to match exactly one IPv4 Resource Pool. Matching zero or more " +
			"pools will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the desired IPv4 Resource Pool.",
				Computed:            true,
				Optional:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "(Non unique) name of the ASN resource pool.",
				Computed:            true,
				Optional:            true,
				Type:                types.StringType,
			},
			"status": {
				MarkdownDescription: "Status of the IPv4 resource pool.",
				Computed:            true,
				Type:                types.StringType,
			},
			"total": {
				MarkdownDescription: "Total number of addresses in the IPv4 resource pool.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"used": {
				MarkdownDescription: "Count of used addresses in the IPv4 resource pool.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"used_percentage": {
				MarkdownDescription: "Percent of used addresses in the IPv4 resource pool.",
				Computed:            true,
				Type:                types.Float64Type,
			},
			"created_at": {
				MarkdownDescription: "Creation time.",
				Computed:            true,
				Type:                types.StringType,
			},
			"last_modified_at": {
				MarkdownDescription: "Last modification time.",
				Computed:            true,
				Type:                types.StringType,
			},
			"subnets": {
				MarkdownDescription: "Detailed info about individual IPv4 CIDR allocations within the IPv4 Resource Pool.",
				Computed:            true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"status": {
						MarkdownDescription: "Status of the IPv4 resource pool.",
						Type:                types.StringType,
						Computed:            true,
					},
					"network": {
						MarkdownDescription: "Network specification in CIDR syntax (\"10.0.0.0/8\").",
						Type:                types.StringType,
						Required:            true,
					},
					"total": {
						MarkdownDescription: "Total number of addresses in this IPv4 range.",
						Type:                types.Int64Type,
						Computed:            true,
					},
					"used": {
						MarkdownDescription: "Count of used addresses in this IPv4 range.",
						Type:                types.Int64Type,
						Computed:            true,
					},
					"used_percentage": {
						MarkdownDescription: "Percent of used addresses in this IPv4 range.",
						Type:                types.Float64Type,
						Computed:            true,
					},
				}),
			},
		},
	}, nil
}

func (o *dataSourceIp4Pool) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dIp4Pool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError(
			"cannot search for ASN Pool",
			"exactly one of 'name' or 'id' must be specified",
		)
	}
}

func (o *dataSourceIp4Pool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dIp4Pool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var ipPool *goapstra.IpPool
	switch {
	case !config.Name.IsNull():
		ipPool, err = o.client.GetIp4PoolByName(ctx, config.Name.ValueString())
	case !config.Id.IsNull():
		ipPool, err = o.client.GetIp4Pool(ctx, goapstra.ObjectId(config.Id.ValueString()))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving IPv4 pool",
			fmt.Sprintf("cannot retrieve IPv4 pool - %s", err),
		)
		return
	}

	config.Id = types.StringValue(string(ipPool.Id))
	config.Name = types.StringValue(ipPool.DisplayName)
	config.Status = types.StringValue(ipPool.Status)
	config.UsedPercent = types.Float64Value(float64(ipPool.UsedPercentage))
	config.CreatedAt = types.StringValue(ipPool.CreatedAt.String())
	config.LastModifiedAt = types.StringValue(ipPool.LastModifiedAt.String())
	config.Used = types.NumberValue(bigIntToBigFloat(&ipPool.Used))
	config.Total = types.NumberValue(bigIntToBigFloat(&ipPool.Total))
	config.Subnets = make([]dIp4PoolSubnet, len(ipPool.Subnets))

	for i, subnet := range ipPool.Subnets {
		config.Subnets[i] = dIp4PoolSubnet{
			Status:         types.StringValue(subnet.Status),
			Network:        types.StringValue(subnet.Network.String()),
			Total:          types.NumberValue(bigIntToBigFloat(&subnet.Total)),
			Used:           types.NumberValue(bigIntToBigFloat(&subnet.Used)),
			UsedPercentage: types.Float64Value(float64(subnet.UsedPercentage)),
		}
	}

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

type dIp4Pool struct {
	Id             types.String     `tfsdk:"id"`
	Name           types.String     `tfsdk:"name"`
	Status         types.String     `tfsdk:"status"`
	Used           types.Number     `tfsdk:"used"`
	UsedPercent    types.Float64    `tfsdk:"used_percentage"`
	CreatedAt      types.String     `tfsdk:"created_at"`
	LastModifiedAt types.String     `tfsdk:"last_modified_at"`
	Total          types.Number     `tfsdk:"total"`
	Subnets        []dIp4PoolSubnet `tfsdk:"subnets"`
}

type dIp4PoolSubnet struct {
	Status         types.String  `tfsdk:"status"`
	Network        types.String  `tfsdk:"network"`
	Total          types.Number  `tfsdk:"total"`
	Used           types.Number  `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}
