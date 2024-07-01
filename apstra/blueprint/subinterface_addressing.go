package blueprint

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubinterfaceAddressing struct {
	BlueprintId types.String         `tfsdk:"blueprint_id"`
	Ipv4Type    types.String         `tfsdk:"ipv4_address_type"`
	Ipv4Addr    cidrtypes.IPv4Prefix `tfsdk:"ipv4_addr"`
	Ipv6Type    types.String         `tfsdk:"ipv6_address_type"`
	Ipv6Addr    cidrtypes.IPv6Prefix `tfsdk:"ipv6_addr"`
}

func (o SubinterfaceAddressing) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required: true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ipv4_address_type": resourceSchema.StringAttribute{
			MarkdownDescription: "",
			Validators: []validator.String{stringvalidator.OneOf(apstra.InterfaceNumberingIpv4TypeNone.String())},
		},
		"ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "",
			CustomType: cidrtypes.IPv4PrefixType{},
			Validators: []validator.String,
		},
	}

}
