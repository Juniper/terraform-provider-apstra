package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ip4PoolSubnet struct {
	Status         types.String  `tfsdk:"status"`
	CIDR           types.String  `tfsdk:"cidr"`
	Total          types.Number  `tfsdk:"total"`
	Used           types.Number  `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

func (o ip4PoolSubnet) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv4 resource pool.",
			Computed:            true,
		},
		"network": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"10.0.0.0/8\").",
			Required:            true,
		},
		"total": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of addresses in this IPv4 range.",
			Computed:            true,
		},
		"used": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used addresses in this IPv4 range.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in this IPv4 range.",
			Computed:            true,
		},
	}
}

func (o ip4PoolSubnet) resourceAttributesWrite() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv4 resource pool.",
			Computed:            true,
		},
		"cidr": resourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"10.0.0.0/8\").",
			Required:            true,
		},
		"total": resourceSchema.NumberAttribute{
			MarkdownDescription: "Total number of addresses in this IPv4 range.",
			Computed:            true,
		},
		"used": resourceSchema.NumberAttribute{
			MarkdownDescription: "Count of used addresses in this IPv4 range.",
			Computed:            true,
		},
		"used_percentage": resourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in this IPv4 range.",
			Computed:            true,
		},
	}
}

func (o ip4PoolSubnet) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":          types.StringType,
		"cidr":            types.StringType,
		"total":           types.NumberType,
		"used":            types.NumberType,
		"used_percentage": types.Float64Type,
	}
}

func (o *ip4PoolSubnet) loadApiData(_ context.Context, in *goapstra.IpSubnet, _ *diag.Diagnostics) {
	o.Status = types.StringValue(in.Status)
	o.CIDR = types.StringValue(in.Network.String())
	o.Total = types.NumberValue(bigIntToBigFloat(&in.Total))
	o.Used = types.NumberValue(bigIntToBigFloat(&in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
}

func (o *ip4PoolSubnet) request(_ context.Context, _ *diag.Diagnostics) *goapstra.NewIpSubnet {
	return &goapstra.NewIpSubnet{
		Network: o.CIDR.ValueString(),
	}
}
