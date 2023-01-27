package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func (o *dataSourceIp4Pool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a single IPv4 Resource Pool. It is incumbent upon " +
			"the user to set enough optional criteria to match exactly one IPv4 Resource Pool. Matching zero or more " +
			"pools will produce an error.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the desired IPv4 Resource Pool.",
				Computed:            true,
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "(Non unique) name of the ASN resource pool.",
				Computed:            true,
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the IPv4 resource pool.",
				Computed:            true,
			},
			"total": schema.NumberAttribute{
				MarkdownDescription: "Total number of addresses in the IPv4 resource pool.",
				Computed:            true,
			},
			"used": schema.NumberAttribute{
				MarkdownDescription: "Count of used addresses in the IPv4 resource pool.",
				Computed:            true,
			},
			"used_percentage": schema.Float64Attribute{
				MarkdownDescription: "Percent of used addresses in the IPv4 resource pool.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation time.",
				Computed:            true,
			},
			"last_modified_at": schema.StringAttribute{
				MarkdownDescription: "Last modification time.",
				Computed:            true,
			},
			"subnets": schema.SetNestedAttribute{
				MarkdownDescription: "Detailed info about individual IPv4 CIDR allocations within the IPv4 Resource Pool.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"status": schema.StringAttribute{
							MarkdownDescription: "Status of the IPv4 resource pool.",
							Computed:            true,
						},
						"network": schema.StringAttribute{
							MarkdownDescription: "Network specification in CIDR syntax (\"10.0.0.0/8\").",
							Required:            true,
						},
						"total": schema.Int64Attribute{
							MarkdownDescription: "Total number of addresses in this IPv4 range.",
							Computed:            true,
						},
						"used": schema.Int64Attribute{
							MarkdownDescription: "Count of used addresses in this IPv4 range.",
							Computed:            true,
						},
						"used_percentage": schema.Float64Attribute{
							MarkdownDescription: "Percent of used addresses in this IPv4 range.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
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
	var ip4Pool *goapstra.IpPool
	switch {
	case !config.Name.IsNull():
		ip4Pool, err = o.client.GetIp4PoolByName(ctx, config.Name.ValueString())
	case !config.Id.IsNull():
		ip4Pool, err = o.client.GetIp4Pool(ctx, goapstra.ObjectId(config.Id.ValueString()))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving IPv4 pool",
			fmt.Sprintf("cannot retrieve IPv4 pool - %s", err),
		)
		return
	}

	// create new state object
	var state dIp4Pool
	state.loadApiResponse(ctx, ip4Pool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dump, _ := json.MarshalIndent(&state, "", "  ")
	resp.Diagnostics.AddWarning("dump", string(dump))

	// Set state
	diags = resp.State.Set(ctx, &state)
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

func (o *dIp4Pool) loadApiResponse(_ context.Context, in *goapstra.IpPool, _ *diag.Diagnostics) {
	subnets := make([]dIp4PoolSubnet, len(in.Subnets))
	for i, s := range in.Subnets {
		subnets[i] = dIp4PoolSubnet{
			Status:         types.StringValue(s.Status),
			Network:        types.StringValue(s.Network.String()),
			Total:          types.NumberValue(bigIntToBigFloat(&s.Total)),
			Used:           types.NumberValue(bigIntToBigFloat(&s.Used)),
			UsedPercentage: types.Float64Value(float64(s.UsedPercentage)),
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.DisplayName)
	o.Status = types.StringValue(in.Status)
	o.UsedPercent = types.Float64Value(float64(in.UsedPercentage))
	o.CreatedAt = types.StringValue(in.CreatedAt.String())
	o.LastModifiedAt = types.StringValue(in.LastModifiedAt.String())
	o.Used = types.NumberValue(bigIntToBigFloat(&in.Used))
	o.Total = types.NumberValue(bigIntToBigFloat(&in.Total))
	o.Subnets = make([]dIp4PoolSubnet, len(in.Subnets))
}

type dIp4PoolSubnet struct {
	Status         types.String  `tfsdk:"status"`
	Network        types.String  `tfsdk:"network"`
	Total          types.Number  `tfsdk:"total"`
	Used           types.Number  `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}
