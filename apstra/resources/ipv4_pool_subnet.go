package resources

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Ipv4PoolSubnet struct {
	Status         types.String         `tfsdk:"status"`
	Network        cidrtypes.IPv4Prefix `tfsdk:"network"`
	Total          types.Number         `tfsdk:"total"`
	Used           types.Number         `tfsdk:"used"`
	UsedPercentage types.Float64        `tfsdk:"used_percentage"`
}

func (o Ipv4PoolSubnet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv4 resource pool.",
			Computed:            true,
		},
		"network": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"10.0.0.0/8\").",
			CustomType:          cidrtypes.IPv4PrefixType{},
			Required:            true,
		},
		"total": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Total number of addresses in this IPv4 range.",
			Computed:            true,
		},
		"used": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Count of used addresses in this IPv4 range.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in this IPv4 range.",
			Computed:            true,
		},
	}
}

func (o Ipv4PoolSubnet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Mutable read-only is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
		},
		"network": resourceSchema.StringAttribute{
			MarkdownDescription: "Network specification in CIDR syntax (\"192.0.2.0/24\").",
			CustomType:          cidrtypes.IPv4PrefixType{},
			Required:            true,
			Validators:          []validator.String{apstravalidator.ParseCidr(true, false)},
			// ParseCidr is still required because the IPv4PrefixType doesn't enforce the zero address.
		},
		"total": resourceSchema.NumberAttribute{
			MarkdownDescription: "Mutable read-only is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
		},
		"used": resourceSchema.NumberAttribute{
			MarkdownDescription: "Mutable read-only is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
		},
		"used_percentage": resourceSchema.Float64Attribute{
			MarkdownDescription: "Mutable read-only is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
		},
	}
}

func (o Ipv4PoolSubnet) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":          types.StringType,
		"network":         types.StringType,
		"total":           types.NumberType,
		"used":            types.NumberType,
		"used_percentage": types.Float64Type,
	}
}

func (o *Ipv4PoolSubnet) LoadApiData(_ context.Context, in *apstra.IpSubnet, _ *diag.Diagnostics) {
	o.Status = types.StringValue(in.Status)
	o.Network = cidrtypes.NewIPv4PrefixValue(in.Network.String())
	o.Total = types.NumberValue(utils.BigIntToBigFloat(&in.Total))
	o.Used = types.NumberValue(utils.BigIntToBigFloat(&in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
}

func (o *Ipv4PoolSubnet) Request(_ context.Context, _ *diag.Diagnostics) *apstra.NewIpSubnet {
	return &apstra.NewIpSubnet{
		Network: o.Network.ValueString(),
	}
}

func (o *Ipv4PoolSubnet) setMutablesToNull() {
	o.Status = types.StringNull()
	o.Total = types.NumberNull()
	o.Used = types.NumberNull()
	o.UsedPercentage = types.Float64Null()
}
