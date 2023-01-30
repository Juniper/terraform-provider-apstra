package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAsnPool{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceAsnPool{}

type dataSourceAsnPool struct {
	client *goapstra.Client
}

func (o *dataSourceAsnPool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pool"
}

func (o *dataSourceAsnPool) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceAsnPool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a single ASN Resource Pool. It is incumbent upon " +
			"the user to set enough optional criteria to match exactly one ASN Resource Pool. Matching zero or more " +
			"pools will produce an error.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the desired ASN Resource Pool.",
				Computed:            true,
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name of the ASN Resource Pool.",
				Computed:            true,
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the ASN Resource Pool.",
				Computed:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of ASNs in the ASN Resource Pool.",
				Computed:            true,
			},
			"used": schema.Int64Attribute{
				MarkdownDescription: "Count of used ASNs in the ASN Resource Pool.",
				Computed:            true,
			},
			"used_percentage": schema.Float64Attribute{
				MarkdownDescription: "Percent of used ASNs in the ASN Resource Pool.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation time of the ASN Resource Pool.",
				Computed:            true,
			},
			"last_modified_at": schema.StringAttribute{
				MarkdownDescription: "Modification time of the ASN Resource Pool.",
				Computed:            true,
			},
			"ranges": schema.ListNestedAttribute{
				MarkdownDescription: "Detailed info about individual ASN Pool Ranges within the ASN Resource Pool.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"status": schema.StringAttribute{
							MarkdownDescription: "Status of the ASN Pool Range, as reported by Apstra.",
							Computed:            true,
						},
						"first": schema.Int64Attribute{
							MarkdownDescription: "Lowest numbered AS in this ASN Pool Range.",
							Computed:            true,
						},
						"last": schema.Int64Attribute{
							MarkdownDescription: "Highest numbered AS in this ASN Pool Range.",
							Computed:            true,
						},
						"total": schema.Int64Attribute{
							MarkdownDescription: "Total number of ASNs in the ASN Pool Range.",
							Computed:            true,
						},
						"used": schema.Int64Attribute{
							MarkdownDescription: "Count of used ASNs in the ASN Pool Range.",
							Computed:            true,
						},
						"used_percentage": schema.Float64Attribute{
							MarkdownDescription: "Percent of used ASNs in the ASN Pool Range.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (o *dataSourceAsnPool) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dAsnPool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError(
			errInvalidConfig,
			"exactly one of 'name' or 'id' must be specified",
		)
	}
}

func (o *dataSourceAsnPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dAsnPool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var asnPool *goapstra.AsnPool
	switch {
	case !config.Name.IsNull():
		asnPool, err = o.client.GetAsnPoolByName(ctx, config.Name.ValueString())
	case !config.Id.IsNull():
		asnPool, err = o.client.GetAsnPool(ctx, goapstra.ObjectId(config.Id.ValueString()))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving ASN pool",
			fmt.Sprintf("cannot retrieve ASN pool - %s", err),
		)
		return
	}

	// create new state object
	var state dAsnPool
	state.loadApiResponse(ctx, asnPool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

type dAsnPool struct {
	Id             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Status         types.String  `tfsdk:"status"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercent    types.Float64 `tfsdk:"used_percentage"`
	CreatedAt      types.String  `tfsdk:"created_at"`
	LastModifiedAt types.String  `tfsdk:"last_modified_at"`
	Total          types.Int64   `tfsdk:"total"`
	Ranges         []dAsnRange   `tfsdk:"ranges"`
}

func (o *dAsnPool) loadApiResponse(_ context.Context, in *goapstra.AsnPool, _ *diag.Diagnostics) {
	ranges := make([]dAsnRange, len(in.Ranges))
	for i, r := range in.Ranges {
		ranges[i] = dAsnRange{
			Status:         types.StringValue(r.Status),
			First:          types.Int64Value(int64(r.First)),
			Last:           types.Int64Value(int64(r.Last)),
			Total:          types.Int64Value(int64(r.Total)),
			Used:           types.Int64Value(int64(r.Used)),
			UsedPercentage: types.Float64Value(float64(r.UsedPercentage)),
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.DisplayName)
	o.Status = types.StringValue(in.Status)
	o.Used = types.Int64Value(int64(in.Used))
	o.UsedPercent = types.Float64Value(float64(in.UsedPercentage))
	o.CreatedAt = types.StringValue(in.CreatedAt.String())
	o.LastModifiedAt = types.StringValue(in.LastModifiedAt.String())
	o.Total = types.Int64Value(int64(in.Total))
	o.Ranges = ranges
}

type dAsnRange struct {
	Status         types.String  `tfsdk:"status"`
	First          types.Int64   `tfsdk:"first"`
	Last           types.Int64   `tfsdk:"last"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}
