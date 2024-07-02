package blueprint

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubinterfaceAddressing struct {
	BlueprintId    types.String         `tfsdk:"blueprint_id"`
	SubinterfaceId types.String         `tfsdk:"subinterface_id"`
	Ipv4Type       types.String         `tfsdk:"ipv4_address_type"`
	Ipv4Addr       cidrtypes.IPv4Prefix `tfsdk:"ipv4_address"`
	Ipv6Type       types.String         `tfsdk:"ipv6_address_type"`
	Ipv6Addr       cidrtypes.IPv6Prefix `tfsdk:"ipv6_address"`
}

func (o SubinterfaceAddressing) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"subinterface_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID of the node with which IP addresses should be associated.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ipv4_address_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Allowed values: [`%s`]", strings.Join(utils.AllInterfaceNumberingIpv4Types(), "`,`")),
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv4TypeNone)),
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllInterfaceNumberingIpv4Types()...)},
		},
		"ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address in CIDR notation.",
			Optional:            true,
			CustomType:          cidrtypes.IPv4PrefixType{},
			Validators: []validator.String{
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("ipv4_address_type"), types.StringValue(apstra.InterfaceNumberingIpv4TypeNumbered.String())),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ipv4_address_type"), types.StringValue(apstra.InterfaceNumberingIpv4TypeNone.String())),
				stringvalidator.AlsoRequires(path.MatchRoot("ipv4_address_type")),
			},
		},
		"ipv6_address_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Allowed values: [`%s`]", strings.Join(utils.AllInterfaceNumberingIpv6Types(), "`,`")),
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeNone)),
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllInterfaceNumberingIpv6Types()...)},
		},
		"ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address in CIDR notation.",
			Optional:            true,
			CustomType:          cidrtypes.IPv6PrefixType{},
			Validators: []validator.String{
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("ipv6_address_type"), types.StringValue(apstra.InterfaceNumberingIpv6TypeNumbered.String())),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ipv6_address_type"), types.StringValue(apstra.InterfaceNumberingIpv6TypeNone.String())),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ipv6_address_type"), types.StringValue(apstra.InterfaceNumberingIpv6TypeLinkLocal.String())),
				stringvalidator.AlsoRequires(path.MatchRoot("ipv6_address_type")),
			},
		},
	}
}

func (o SubinterfaceAddressing) Request(_ context.Context, diags *diag.Diagnostics) map[apstra.ObjectId]apstra.TwoStageL3ClosSubinterface {
	ipv4AddrType := apstra.InterfaceNumberingIpv4TypeNone
	if !o.Ipv4Type.IsNull() {
		err := utils.ApiStringerFromFriendlyString(&ipv4AddrType, o.Ipv4Type.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("ipv4_address_type"), "Cannot parse ipv4_address_type", err.Error())
		}
	}

	ipv6AddrType := apstra.InterfaceNumberingIpv6TypeNone
	if !o.Ipv6Type.IsNull() {
		err := utils.ApiStringerFromFriendlyString(&ipv6AddrType, o.Ipv6Type.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("ipv6_address_type"), "Cannot parse ipv6_address_type", err.Error())
		}
	}

	var err error
	var ipv4Addr, ipv6Addr *net.IPNet

	if !o.Ipv4Addr.IsNull() {
		_, ipv4Addr, err = net.ParseCIDR(o.Ipv4Addr.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("ipv4_address"), "Cannot parse ipv4_address", err.Error())
		}
	}

	if !o.Ipv6Addr.IsNull() {
		_, ipv6Addr, err = net.ParseCIDR(o.Ipv6Addr.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("ipv6_address"), "Cannot parse ipv6_address", err.Error())
		}
	}

	return map[apstra.ObjectId]apstra.TwoStageL3ClosSubinterface{
		apstra.ObjectId(o.SubinterfaceId.ValueString()): {
			Ipv4AddrType: ipv4AddrType,
			Ipv6AddrType: ipv6AddrType,
			Ipv4Addr:     ipv4Addr,
			Ipv6Addr:     ipv6Addr,
		},
	}
}

func (o SubinterfaceAddressing) LoadApiData(_ context.Context, in *apstra.TwoStageL3ClosSubinterface, diags *diag.Diagnostics) {
	o.Ipv4Type = types.StringValue(utils.StringersToFriendlyString(in.Ipv4AddrType))
	o.Ipv6Type = types.StringValue(utils.StringersToFriendlyString(in.Ipv6AddrType))

	o.Ipv4Addr = cidrtypes.NewIPv4PrefixNull()
	if in.Ipv4Addr != nil {
		o.Ipv4Addr = cidrtypes.NewIPv4PrefixValue(in.Ipv4Addr.String())
	}

	o.Ipv6Addr = cidrtypes.NewIPv6PrefixNull()
	if in.Ipv6Addr != nil {
		o.Ipv6Addr = cidrtypes.NewIPv6PrefixValue(in.Ipv6Addr.String())
	}
}
