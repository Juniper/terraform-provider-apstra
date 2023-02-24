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

type ip4Pool struct {
	Id             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Status         types.String  `tfsdk:"status"`
	Used           types.Number  `tfsdk:"used"`
	UsedPercent    types.Float64 `tfsdk:"used_percentage"`
	CreatedAt      types.String  `tfsdk:"created_at"`
	LastModifiedAt types.String  `tfsdk:"last_modified_at"`
	Total          types.Number  `tfsdk:"total"`
	Subnets        types.Set     `tfsdk:"subnets"`
}

func (o ip4Pool) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the desired IPv4 Resource Pool.",
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
			MarkdownDescription: "(Non unique) name of the ASN resource pool.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv4 resource pool.",
			Computed:            true,
		},
		"total": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Total number of addresses in the IPv4 resource pool.",
			Computed:            true,
		},
		"used": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Count of used addresses in the IPv4 resource pool.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in the IPv4 resource pool.",
			Computed:            true,
		},
		"created_at": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Creation time.",
			Computed:            true,
		},
		"last_modified_at": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Last modification time.",
			Computed:            true,
		},
		"subnets": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Detailed info about individual IPv4 CIDR allocations within the IPv4 Resource Pool.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: dIp4PoolSubnet{}.attributes(),
			},
		},
	}
}

func (o *ip4Pool) loadApiData(ctx context.Context, in *goapstra.IpPool, diags *diag.Diagnostics) {
	subnets := make([]ip4PoolSubnet, len(in.Subnets))
	for i, subnet := range in.Subnets {
		subnets[i].loadApiData(ctx, &subnet, diags)
		if diags.HasError() {
			return
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.DisplayName)
	o.Status = types.StringValue(in.Status)
	o.UsedPercent = types.Float64Value(float64(in.UsedPercentage))
	o.CreatedAt = types.StringValue(in.CreatedAt.String())
	o.LastModifiedAt = types.StringValue(in.LastModifiedAt.String())
	o.Used = types.NumberValue(bigIntToBigFloat(&in.Used))
	o.Total = types.NumberValue(bigIntToBigFloat(&in.Total))
	o.Subnets = setValueOrNull(ctx, types.ObjectType{AttrTypes: ip4PoolSubnet{}.attrTypes()}, subnets, diags)
}
