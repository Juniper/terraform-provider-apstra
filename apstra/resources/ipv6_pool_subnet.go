package resources

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/utils"
)

type Ipv6PoolSubnet struct {
	Status         types.String         `tfsdk:"status"`
	Network        cidrtypes.IPv6Prefix `tfsdk:"network"`
	Total          types.Number         `tfsdk:"total"`
	Used           types.Number         `tfsdk:"used"`
	UsedPercentage types.Float64        `tfsdk:"used_percentage"`
}

func (o Ipv6PoolSubnet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv6 resource pool.",
			Computed:            true,
		},
		"network": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"2001:db8::/32\").",
			CustomType:          new(cidrtypes.IPv6PrefixType),
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

func (o Ipv6PoolSubnet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv6 resource pool.",
			Computed:            true,
		},
		"network": resourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"2001:db8::/64\").",
			CustomType:          new(cidrtypes.IPv6PrefixType),
			Required:            true,
			Validators:          []validator.String{apstravalidator.ParseCidr(false, true)},
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

func (o Ipv6PoolSubnet) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":          types.StringType,
		"network":         types.StringType,
		"total":           types.NumberType,
		"used":            types.NumberType,
		"used_percentage": types.Float64Type,
	}
}

func (o *Ipv6PoolSubnet) LoadApiData(_ context.Context, in *apstra.IpSubnet, _ *diag.Diagnostics) {
	o.Status = types.StringValue(in.Status)
	o.Network = cidrtypes.NewIPv6PrefixValue(in.Network.String())
	o.Total = types.NumberValue(utils.BigIntToBigFloat(&in.Total))
	o.Used = types.NumberValue(utils.BigIntToBigFloat(&in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
}

func (o *Ipv6PoolSubnet) Request(_ context.Context, _ *diag.Diagnostics) *apstra.NewIpSubnet {
	return &apstra.NewIpSubnet{
		Network: o.Network.ValueString(),
	}
}
