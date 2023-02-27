package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dcBlueprintStatus struct {
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

func (o dcBlueprintStatus) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Blueprint: Either as a result of a lookup, or user-specified.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Blueprint: Either as a result of a lookup, or user-specified.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the blueprint",
			Computed:            true,
		},
		"superspine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
			Computed:            true,
		},
		"spine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine devices in the topology.",
			Computed:            true,
		},
		"leaf_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of leaf switches in the topology.",
			Computed:            true,
		},
		"access_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of access switches in the topology.",
			Computed:            true,
		},
		"generic_system_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of generic systems in the topology.",
			Computed:            true,
		},
		"external_router_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of external routers attached to the topology.",
			Computed:            true,
		},
	}
}

func (o *dcBlueprintStatus) loadApiData(_ context.Context, in *goapstra.BlueprintStatus, _ *diag.Diagnostics) {
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
