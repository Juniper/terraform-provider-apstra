package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprint{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceBlueprint{}

type dataSourceBlueprint struct {
	client *goapstra.Client
}

func (o *dataSourceBlueprint) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *dataSourceBlueprint) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceBlueprint) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides high-level details of a single Datacenter Blueprint. It is " +
			"incumbent upon the user to set enough optional criteria to match exactly one Blueprint. Matching zero " +
			"Blueprints or more than one Blueprint will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the Blueprint: Either as a result of a lookup, or user-specified.",
				Computed:            true,
				Optional:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Name of the Blueprint: Either as a result of a lookup, or user-specified.",
				Computed:            true,
				Optional:            true,
				Type:                types.StringType,
			},
			"status": {
				MarkdownDescription: "Deployment status of the blueprint",
				Computed:            true,
				Type:                types.StringType,
			},
			"superspine_count": {
				MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"spine_count": {
				MarkdownDescription: "The count of spine devices in the topology.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"leaf_switch_count": {
				MarkdownDescription: "The count of leaf switches in the topology.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"access_switch_count": {
				MarkdownDescription: "The count of access switches in the topology.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"generic_system_count": {
				MarkdownDescription: "The count of generic systems in the topology.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"external_router_count": {
				MarkdownDescription: "The count of external routers attached to the topology.",
				Computed:            true,
				Type:                types.Int64Type,
			},
		},
	}, nil
}

func (o *dataSourceBlueprint) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dBlueprint
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.Null && config.Id.Null) || (!config.Name.Null && !config.Id.Null) { // XOR
		resp.Diagnostics.AddError(
			"cannot search for Blueprint",
			"exactly one of 'name' or 'id' must be specified",
		)
	}
}

func (o *dataSourceBlueprint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dBlueprint
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var status *goapstra.BlueprintStatus
	var err error
	switch {
	case !config.Name.IsNull():
		status, err = o.client.GetBlueprintStatusByName(ctx, config.Name.Value)
		if err != nil {
			resp.Diagnostics.AddError("error getting blueprint status by name", err.Error())
			return
		}
	case !config.Id.IsNull():
		status, err = o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(config.Id.Value))
		if err != nil {
			resp.Diagnostics.AddError("error getting blueprint status by name", err.Error())
			return
		}
	default:
		resp.Diagnostics.AddError("both name and id seem to be null", "this should not have happened please report to provider maintainers")
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &dBlueprint{
		Id:              types.String{Value: string(status.Id)},
		Name:            types.String{Value: status.Label},
		Status:          types.String{Value: status.Status},
		SuperspineCount: types.Int64{Value: int64(status.SuperspineCount)},
		SpineCount:      types.Int64{Value: int64(status.SpineCount)},
		LeafCount:       types.Int64{Value: int64(status.LeafCount)},
		AccessCount:     types.Int64{Value: int64(status.AccessCount)},
		GenericCount:    types.Int64{Value: int64(status.GenericCount)},
		ExternalCount:   types.Int64{Value: int64(status.ExternalRouterCount)},
	})
	resp.Diagnostics.Append(diags...)
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
