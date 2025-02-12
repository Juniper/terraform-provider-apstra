package blueprint

import (
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type RoutingZoneLoopback struct {
	Ipv4Addr cidrtypes.IPv4Prefix `tfsdk:"ipv4_addr"`
	Ipv6Addr cidrtypes.IPv6Prefix `tfsdk:"ipv6_addr"`
}

func (o RoutingZoneLoopback) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ipv4_addr": cidrtypes.IPv4PrefixType{},
		"ipv6_addr": cidrtypes.IPv6PrefixType{},
	}
}

func (o RoutingZoneLoopback) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"ipv4_addr": resourceSchema.StringAttribute{
			CustomType:          cidrtypes.IPv4PrefixType{},
			MarkdownDescription: "The IPv4 address to be assigned within the Routing Zone, in CIDR notation.",
			Optional:            true,
		},
		"ipv6_addr": resourceSchema.StringAttribute{
			CustomType:          cidrtypes.IPv6PrefixType{},
			MarkdownDescription: "The IPv6 address to be assigned within the Routing Zone, in CIDR notation.",
			Optional:            true,
		},
	}
}
