package blueprint

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SviMapEntry struct {
	SystemId  types.String `tfsdk:"system_id"`
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Ipv4Addr  types.String `tfsdk:"ipv4_addr"`
	Ipv6Addr  types.String `tfsdk:"ipv6_addr"`
	Ipv4Mode  types.String `tfsdk:"ipv4_mode"`
	Ipv6Mode  types.String `tfsdk:"ipv6_mode"`
	NetworkId types.String `tfsdk:"virtual_network_id"`
}

func (o SviMapEntry) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"system_id":          types.StringType,
		"id":                 types.StringType,
		"name":               types.StringType,
		"ipv4_addr":          types.StringType,
		"ipv6_addr":          types.StringType,
		"ipv4_mode":          types.StringType,
		"ipv6_mode":          types.StringType,
		"virtual_network_id": types.StringType,
	}
}

//func (o SviMapEntry) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
//	return map[string]dataSourceSchema.Attribute{
//		"system_id": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Graph node ID of the System which owns this SVI",
//			Computed:            true,
//		},
//		"svi_id": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Graph node ID of the SVI interface",
//			Computed:            true,
//		},
//		"name": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Interface name",
//			Computed:            true,
//		},
//		"ipv4_addr": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "IPv4 address/mask of the SVI",
//			Computed:            true,
//		},
//		"ipv6_addr": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "IPv6 address/mask of the SVI",
//			Computed:            true,
//		},
//		"ipv4_mode": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Secondary IPv4 allocation mode",
//			Computed:            true,
//		},
//		"ipv6_mode": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "Secondary IPv6 allocation mode",
//			Computed:            true,
//		},
//		"virtual_network_id": dataSourceSchema.StringAttribute{
//			MarkdownDescription: "ID of the Virtual Network associated with this SVI",
//			Computed:            true,
//		},
//	}
//}
