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

type ipv6PoolSubnet struct {
	Status         types.String  `tfsdk:"status"`
	Network        types.String  `tfsdk:"network"`
	Total          types.Number  `tfsdk:"total"`
	Used           types.Number  `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

func (o ipv6PoolSubnet) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv6 resource pool.",
			Computed:            true,
		},
		"network": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"2001:db8::/32\").",
			Required:            true,
		},
		"total": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Total number of addresses in this IPv6 range.",
			Computed:            true,
		},
		"used": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Count of used addresses in this IPv6 range.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in this IPv6 range.",
			Computed:            true,
		},
	}
}

func (o ipv6PoolSubnet) resourceAttributesWrite() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv6 resource pool.",
			Computed:            true,
		},
		"network": resourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"10.0.0.0/8\").",
			Required:            true,
		},
		"total": resourceSchema.NumberAttribute{
			MarkdownDescription: "Total number of addresses in this IPv6 range.",
			Computed:            true,
		},
		"used": resourceSchema.NumberAttribute{
			MarkdownDescription: "Count of used addresses in this IPv6 range.",
			Computed:            true,
		},
		"used_percentage": resourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in this IPv6 range.",
			Computed:            true,
		},
	}
}

func (o ipv6PoolSubnet) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":          types.StringType,
		"network":         types.StringType,
		"total":           types.NumberType,
		"used":            types.NumberType,
		"used_percentage": types.Float64Type,
	}
}

func (o *ipv6PoolSubnet) loadApiData(_ context.Context, in *goapstra.IpSubnet, _ *diag.Diagnostics) {
	o.Status = types.StringValue(in.Status)
	o.Network = types.StringValue(in.Network.String())
	o.Total = types.NumberValue(bigIntToBigFloat(&in.Total))
	o.Used = types.NumberValue(bigIntToBigFloat(&in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
}

func (o *ipv6PoolSubnet) request(_ context.Context, _ *diag.Diagnostics) *goapstra.NewIpSubnet {
	return &goapstra.NewIpSubnet{
		Network: o.Network.ValueString(),
	}
}
