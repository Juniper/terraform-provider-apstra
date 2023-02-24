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

type asnPool struct {
	Id             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Status         types.String  `tfsdk:"status"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercent    types.Float64 `tfsdk:"used_percentage"`
	CreatedAt      types.String  `tfsdk:"created_at"`
	LastModifiedAt types.String  `tfsdk:"last_modified_at"`
	Total          types.Int64   `tfsdk:"total"`
	Ranges         types.Set     `tfsdk:"ranges"`
}

func (o asnPool) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the desired ASN Resource Pool.",
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
			MarkdownDescription: "Display name of the ASN Resource Pool.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the ASN Resource Pool.",
			Computed:            true,
		},
		"total": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of ASNs in the ASN Resource Pool.",
			Computed:            true,
		},
		"used": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used ASNs in the ASN Resource Pool.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used ASNs in the ASN Resource Pool.",
			Computed:            true,
		},
		"created_at": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Creation time of the ASN Resource Pool.",
			Computed:            true,
		},
		"last_modified_at": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Modification time of the ASN Resource Pool.",
			Computed:            true,
		},
		"ranges": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Detailed info about individual ASN Pool Ranges within the ASN Resource Pool.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: asnPoolRange{}.dataSourceAttributes(),
			},
		},
	}
}

func (o *asnPool) loadApiData(ctx context.Context, in *goapstra.AsnPool, diags *diag.Diagnostics) {
	ranges := make([]asnPoolRange, len(in.Ranges))
	for i, r := range in.Ranges {
		ranges[i].loadApiData(ctx, &r, diags)
		if diags.HasError() {
			return
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
	o.Ranges = setValueOrNull(ctx, types.ObjectType{AttrTypes: asnPoolRange{}.attrTypes()}, ranges, diags)
}
