package device

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CfgInfo struct {
	SystemId                      types.String `tfsdk:"system_id"`
	LastBootTime                  types.String `tfsdk:"last_boot_time"`
	Deviated                      types.Bool   `tfsdk:"deviated"`
	ErrorMessage                  types.String `tfsdk:"error_message"`
	ContiguousFailures            types.Int64  `tfsdk:"contiguous_failures"`
	UserGoldenConfigUpdateVersion types.Int64  `tfsdk:"user_golden_config_update_version"`
	UserFullConfigDeployVersion   types.Int64  `tfsdk:"user_full_config_deploy_version"`
	AosConfigVersion              types.Int64  `tfsdk:"aos_config_version"`
	Expected                      types.String `tfsdk:"config_expected"`
	Actual                        types.String `tfsdk:"config_actual"`
}

func (o CfgInfo) DataSourceAttributes() map[string]dataSourceSchema.Attribute {

	return map[string]dataSourceSchema.Attribute{
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the System, as found in Devices -> Managed Devices in the GUI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"last_boot_time": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Last boot time of the system.",
			Computed:            true,
		},
		"deviated": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Boolean `true` if the configuration has deviated.",
			Computed:            true,
		},
		"error_message": dataSourceSchema.StringAttribute{
			MarkdownDescription: "System error message",
			Computed:            true,
		},
		"contiguous_failures": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Indicates the system's contiguous error count.",
			Computed:            true,
		},
		"user_golden_config_update_version": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Version number of the system's golden configuration",
			Computed:            true,
		},
		"user_full_config_deploy_version": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Version number of the system's full configuration deployed",
			Computed:            true,
		},
		"aos_config_version": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Version number of the aos config",
			Computed:            true,
		},
		"config_expected": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Expected system configuration",
			Computed:            true,
		},
		"config_actual": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Actual system configuration",
			Computed:            true,
		},
	}
}

func (o *CfgInfo) LoadApiData(_ context.Context, in apstra.SystemConfig, _ *diag.Diagnostics) {
	o.SystemId = types.StringValue(string(in.SystemId))
	o.LastBootTime = types.StringValue(in.LastBootTime.String())
	o.Deviated = types.BoolValue(in.Deviated)
	o.ErrorMessage = types.StringPointerValue(in.ErrorMessage)
	o.ContiguousFailures = types.Int64Value(int64(in.ContiguousFailures))
	o.UserGoldenConfigUpdateVersion = types.Int64Value(int64(in.UserGoldenConfigUpdateVersion))
	o.UserFullConfigDeployVersion = types.Int64Value(int64(in.UserFullConfigDeployVersion))
	o.AosConfigVersion = types.Int64Value(int64(in.AosConfigVersion))
	o.Expected = types.StringValue(in.ExpectedConfig)
	o.Actual = types.StringValue(in.ActualConfig)
}
