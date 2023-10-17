package iba

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PredefinedProbe struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Schema      types.String `tfsdk:"schema"`
}

func (o PredefinedProbe) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the IBA Predefined Probe belongs to.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a IBA Predefined Probe.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Description of the IBA Predefined Probe",
			Computed:            true,
		},
		"schema": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Schema of the IBA Predefined Probe's parameters",
			Computed:            true,
		},
	}
}

func (o *PredefinedProbe) LoadApiData(_ context.Context, in *apstra.IbaPredefinedProbe, d *diag.Diagnostics) {
	o.Name = types.StringValue(in.Name)
	o.Description = types.StringValue(in.Description)
	o.Schema = types.StringValue(string(in.Schema))
}
