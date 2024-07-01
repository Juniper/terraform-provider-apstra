package blueprint

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type IpLinkIps struct {
	//Vlan       types.Int64          `tfsdk:"vlan"`
	SwitchIpv4 cidrtypes.IPv4Prefix `tfsdk:"switch_ipv4"`
	PeerIpv4   cidrtypes.IPv4Prefix `tfsdk:"peer_ipv4"`
	SwitchIpv6 cidrtypes.IPv6Prefix `tfsdk:"switch_ipv6"`
	PeerIpv6   cidrtypes.IPv6Prefix `tfsdk:"peer_ipv6"`
}

func (o IpLinkIps) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		//"vlan":        types.Int64Type,
		"switch_ipv4": cidrtypes.IPv4PrefixType{},
		"peer_ipv4":   cidrtypes.IPv4PrefixType{},
		"switch_ipv6": cidrtypes.IPv6PrefixType{},
		"peer_ipv6":   cidrtypes.IPv6PrefixType{},
	}
}

func (o *IpLinkIps) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		//"vlan": resourceSchema.Int64Attribute{
		//	MarkdownDescription: fmt.Sprintf("VLAN ID in the range %d-%d. Used to select between IP Link "+
		//		"primitives found in the Connectivity Template. VLAN %d represents an *untagged* IP Link. ",
		//		design.VlanMin-1, design.VlanMax, design.VlanMin-1),
		//	Required:   true,
		//	Validators: []validator.Int64{int64validator.Between(design.VlanMin-1, design.VlanMax)},
		//},
		"switch_ipv4": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address to be applied to the Leaf Switch subinterface created by assigning a " +
				"Connectivity Template which includes an IP Link primitive",
			CustomType: cidrtypes.IPv4PrefixType{},
			Optional:   true,
		},
		"peer_ipv4": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address to be applied to the Generic System subinterface created by assigning a " +
				"Connectivity Template which includes an IP Link primitive",
			CustomType: cidrtypes.IPv4PrefixType{},
			Optional:   true,
		},
		"switch_ipv6": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address to be applied to the Leaf Switch subinterface created by assigning a " +
				"Connectivity Template which includes an IP Link primitive",
			CustomType: cidrtypes.IPv6PrefixType{},
			Optional:   true,
		},
		"peer_ipv6": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address to be applied to the Generic System subinterface created by assigning a " +
				"Connectivity Template which includes an IP Link primitive",
			CustomType: cidrtypes.IPv6PrefixType{},
			Optional:   true,
		},
	}
}

//func (o *IpLinkIps) setSubinterfaces(ctx context.Context, state *IpLinkIps, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	add := make(map[int64])
//
//}

func (o *IpLinkIps) request() *apstra.TwoStageL3ClosSubinterface{
	var result apstra.TwoStageL3ClosSubinterface
	result.
}