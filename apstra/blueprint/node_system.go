package blueprint

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

//type System struct {
//	BlueprintId types.String `tfsdk:"blueprint_id"`
//	Id          types.String `tfsdk:"id"`
//	Attributes  types.Object `tfsdk:"attributes"`
//}

//func (o System) DataSourceSchemaNested() map[string]dataSourceSchema.Attribute {
//	return map[string]dataSourceSchema.Attribute{
//		"blueprint_id": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Blueprint ID",
//			Computed:            true,
//		},
//		"id": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `ID`",
//			Computed:            true,
//		},
//		"attributes": dataSourceSchema.SingleNestedAttribute{
//			MarkdownDescription: "Attributes of a `system` Graph DB node.",
//			Attributes:          SystemNode{}.DataSourceAttributes(),
//		},
//	}
//}

type SystemNode struct {
	Hostname   types.String `tfsdk:"hostname"`
	Label      types.String `tfsdk:"label"`
	Role       types.String `tfsdk:"role"`
	SystemId   types.String `tfsdk:"system_id"`
	SystemType types.String `tfsdk:"system_type"`
	TagIds     types.Set    `tfsdk:"tag_ids"`
}

//func (o SystemNode) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
//	return map[string]dataSourceSchema.Attribute{
//		"hostname": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `hostname`",
//			Computed:            true,
//		},
//		"label": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `label`",
//			Computed:            true,
//		},
//		"role": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `role`",
//			Computed:            true,
//		},
//		"system_type": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `system_type`",
//			Computed:            true,
//		},
//		"type": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `type`",
//			Computed:            true,
//		},
//	}
//}

func (o SystemNode) DataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"hostname": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `hostname`",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"label": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `label`",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"role": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `role`",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the physical system (not to be confused with its fabric role)",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `system_type`",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag IDs (labels) - only nodes with all tags will match this filter",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				setvalidator.AtLeastOneOf(
					path.MatchRoot("filters").AtName("hostname"),
					path.MatchRoot("filters").AtName("label"),
					path.MatchRoot("filters").AtName("role"),
					path.MatchRoot("filters").AtName("system_id"),
					path.MatchRoot("filters").AtName("system_type"),
					path.MatchRoot("filters").AtName("tag_ids"),
				),
			},
		},
	}
}

//func (o SystemNode) ResourceAttributes() map[string]resourceSchema.Attribute {
//	return map[string]resourceSchema.Attribute{
//		"hostname": resourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `hostname`",
//			Optional:            true,
//			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
//		},
//		"label": resourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `label`",
//			Optional:            true,
//			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
//		},
//		"role": resourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `role`",
//			Optional:            true,
//			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
//		},
//		"system_type": resourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `system_type`",
//			Optional:            true,
//			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
//		},
//		"type": resourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra Graph DB node `type`",
//			Optional:            true,
//			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
//		},
//	}
//}

func (o SystemNode) QEEAttributes() []apstra.QEEAttribute {
	var result []apstra.QEEAttribute

	if utils.Known(o.Hostname) {
		result = append(result, apstra.QEEAttribute{Key: "hostname", Value: apstra.QEStringVal(o.Hostname.ValueString())})
	}

	if utils.Known(o.Label) {
		result = append(result, apstra.QEEAttribute{Key: "label", Value: apstra.QEStringVal(o.Label.ValueString())})
	}

	if utils.Known(o.Role) {
		result = append(result, apstra.QEEAttribute{Key: "role", Value: apstra.QEStringVal(o.Role.ValueString())})
	}

	if utils.Known(o.SystemId) {
		result = append(result, apstra.QEEAttribute{Key: "system_id", Value: apstra.QEStringVal(o.SystemId.ValueString())})
	}

	if utils.Known(o.SystemType) {
		result = append(result, apstra.QEEAttribute{Key: "system_type", Value: apstra.QEStringVal(o.SystemType.ValueString())})
	}

	return result
}
