package blueprint

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RenderedConfig struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	SystemId    types.String `tfsdk:"system_id"`
	NodeId      types.String `tfsdk:"node_id"`
	StagedCfg   types.String `tfsdk:"staged_config"`
	DeployedCfg types.String `tfsdk:"deployed_config"`
	Incremental types.String `tfsdk:"incremental_config"`
}

func (o RenderedConfig) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID (serial number) for the System (Managed Device), as found in " +
				"Devices -> Managed Devices in the GUI. Required when `node_id` is omitted.",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(
					path.MatchRoot("system_id"),
					path.MatchRoot("node_id"),
				),
			},
		},
		"node_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the System (spine, leaf, etc...) node. Required when `system_id` is omitted.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"staged_config": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Staged device configuration.",
			Computed:            true,
		},
		"deployed_config": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Deployed device configuration.",
			Computed:            true,
		},
		"incremental_config": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Incremental device configuration.",
			Computed:            true,
		},
	}
}
