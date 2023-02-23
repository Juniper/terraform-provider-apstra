package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceTwoStageL3ClosBlueprint{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceTwoStageL3ClosBlueprint{}

type dataSourceTwoStageL3ClosBlueprint struct {
	client *goapstra.Client
}

func (o *dataSourceTwoStageL3ClosBlueprint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *dataSourceTwoStageL3ClosBlueprint) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = dataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceTwoStageL3ClosBlueprint) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides high-level details of a single Datacenter Blueprint. It is " +
			"incumbent upon the user to set enough optional criteria to match exactly one Blueprint. Matching zero " +
			"Blueprints or more than one Blueprint will produce an error.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Blueprint: Either as a result of a lookup, or user-specified.",
				Computed:            true,
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Blueprint: Either as a result of a lookup, or user-specified.",
				Computed:            true,
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Deployment status of the blueprint",
				Computed:            true,
			},
			"superspine_count": schema.Int64Attribute{
				MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
				Computed:            true,
			},
			"spine_count": schema.Int64Attribute{
				MarkdownDescription: "The count of spine devices in the topology.",
				Computed:            true,
			},
			"leaf_switch_count": schema.Int64Attribute{
				MarkdownDescription: "The count of leaf switches in the topology.",
				Computed:            true,
			},
			"access_switch_count": schema.Int64Attribute{
				MarkdownDescription: "The count of access switches in the topology.",
				Computed:            true,
			},
			"generic_system_count": schema.Int64Attribute{
				MarkdownDescription: "The count of generic systems in the topology.",
				Computed:            true,
			},
			"external_router_count": schema.Int64Attribute{
				MarkdownDescription: "The count of external routers attached to the topology.",
				Computed:            true,
			},
		},
	}
}

func (o *dataSourceTwoStageL3ClosBlueprint) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dBlueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError(
			"cannot search for Blueprint",
			"exactly one of 'name' or 'id' must be specified",
		)
	}
}

func (o *dataSourceTwoStageL3ClosBlueprint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dBlueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var status *goapstra.BlueprintStatus
	var err error
	switch {
	case !config.Name.IsNull():
		status, err = o.client.GetBlueprintStatusByName(ctx, config.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("error getting blueprint status by name", err.Error())
			return
		}
	case !config.Id.IsNull():
		status, err = o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("error getting blueprint status by name", err.Error())
			return
		}
	default:
		resp.Diagnostics.AddError("both name and id seem to be null", "this should not have happened please report to provider maintainers")
		return
	}

	// create new state object
	var state dBlueprint
	state.loadApiResponse(ctx, status, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type dBlueprint struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Status          types.String `tfsdk:"status"`
	SuperspineCount types.Int64  `tfsdk:"superspine_count"`
	SpineCount      types.Int64  `tfsdk:"spine_count"`
	LeafCount       types.Int64  `tfsdk:"leaf_switch_count"`
	AccessCount     types.Int64  `tfsdk:"access_switch_count"`
	GenericCount    types.Int64  `tfsdk:"generic_system_count"`
	ExternalCount   types.Int64  `tfsdk:"external_router_count"`
}

func (o *dBlueprint) loadApiResponse(ctx context.Context, in *goapstra.BlueprintStatus, diags *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.Status = types.StringValue(in.Status)
	o.SuperspineCount = types.Int64Value(int64(in.SuperspineCount))
	o.SpineCount = types.Int64Value(int64(in.SpineCount))
	o.LeafCount = types.Int64Value(int64(in.LeafCount))
	o.AccessCount = types.Int64Value(int64(in.AccessCount))
	o.GenericCount = types.Int64Value(int64(in.GenericCount))
	o.ExternalCount = types.Int64Value(int64(in.ExternalRouterCount))
}
