package blueprint

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodeTypeSystemAttributes struct {
	Id         types.String `tfsdk:"id"`
	Hostname   types.String `tfsdk:"hostname"`
	Label      types.String `tfsdk:"label"`
	Role       types.String `tfsdk:"role"`
	SystemId   types.String `tfsdk:"system_id"`
	SystemType types.String `tfsdk:"system_type"`
	TagIds     types.Set    `tfsdk:"tag_ids"`
}

func (o NodeTypeSystemAttributes) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"hostname":    types.StringType,
		"label":       types.StringType,
		"role":        types.StringType,
		"system_id":   types.StringType,
		"system_type": types.StringType,
		"tag_ids":     types.SetType{ElemType: types.StringType},
	}
}

func (o NodeTypeSystemAttributes) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"hostname": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `hostname`",
			Computed:            true,
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node ID",
			Computed:            true,
		},
		"label": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `label`",
			Computed:            true,
		},
		"role": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `role`",
			Computed:            true,
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the physical system (not to be confused with its fabric role)",
			Computed:            true,
		},
		"system_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `system_type`",
			Computed:            true,
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Apstra Graph DB tags associated with this system",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o NodeTypeSystemAttributes) DataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"hostname": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `hostname`",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node ID",
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
					path.MatchRoot("filters").AtName("id"),
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

func (o NodeTypeSystemAttributes) QEEAttributes() []apstra.QEEAttribute {
	var result []apstra.QEEAttribute

	if utils.Known(o.Hostname) {
		result = append(result, apstra.QEEAttribute{Key: "hostname", Value: apstra.QEStringVal(o.Hostname.ValueString())})
	}

	if utils.Known(o.Id) {
		result = append(result, apstra.QEEAttribute{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())})
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
