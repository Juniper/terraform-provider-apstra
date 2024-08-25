package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/Juniper/terraform-provider-apstra/apstra/constants"

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

type IpLinkAddressing struct {
	BlueprintId     types.String         `tfsdk:"blueprint_id"`
	LinkId          types.String         `tfsdk:"link_id"`
	SwitchIntfId    types.String         `tfsdk:"switch_interface_id"`
	GenericIntfId   types.String         `tfsdk:"generic_interface_id"`
	SwitchIpv4Type  types.String         `tfsdk:"switch_ipv4_address_type"`
	SwitchIpv4Addr  cidrtypes.IPv4Prefix `tfsdk:"switch_ipv4_address"`
	SwitchIpv6Type  types.String         `tfsdk:"switch_ipv6_address_type"`
	SwitchIpv6Addr  cidrtypes.IPv6Prefix `tfsdk:"switch_ipv6_address"`
	GenericIpv4Type types.String         `tfsdk:"generic_ipv4_address_type"`
	GenericIpv4Addr cidrtypes.IPv4Prefix `tfsdk:"generic_ipv4_address"`
	GenericIpv6Type types.String         `tfsdk:"generic_ipv6_address_type"`
	GenericIpv6Addr cidrtypes.IPv6Prefix `tfsdk:"generic_ipv6_address"`
}

func (o IpLinkAddressing) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"link_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID of the logical Link created by assigning a Connectivity " +
				"Template with an IP Link primitive to a switch port. Note that CT assignment will create a logical " +
				"link for each IP Link primitive. This resource is concerned with a single logical link. CTs which " +
				"include multiple IP Link primitives may require multiple instances of this resource.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"switch_interface_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID of the node to which `switch` IP information will be associated.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"generic_interface_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID of the node to which `generic` IP information will be associated.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"switch_ipv4_address_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Allowed values: [`%s`]", strings.Join(utils.AllInterfaceNumberingIpv4Types(), "`,`")),
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv4TypeNone)),
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllInterfaceNumberingIpv4Types()...)},
		},
		"switch_ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address in CIDR notation.",
			Optional:            true,
			CustomType:          cidrtypes.IPv4PrefixType{},
			Validators: []validator.String{
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("switch_ipv4_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv4TypeNumbered))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("switch_ipv4_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv4TypeNone))),
				stringvalidator.AlsoRequires(path.MatchRoot("switch_ipv4_address_type")),
			},
		},
		"switch_ipv6_address_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Allowed values: [`%s`]", strings.Join(utils.AllInterfaceNumberingIpv6Types(), "`,`")),
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeNone)),
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllInterfaceNumberingIpv6Types()...)},
		},
		"switch_ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address in CIDR notation.",
			Optional:            true,
			CustomType:          cidrtypes.IPv6PrefixType{},
			Validators: []validator.String{
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("switch_ipv6_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeNumbered))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("switch_ipv6_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeNone))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("switch_ipv6_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeLinkLocal))),
				stringvalidator.AlsoRequires(path.MatchRoot("switch_ipv6_address_type")),
			},
		},
		"generic_ipv4_address_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Allowed values: [`%s`]", strings.Join(utils.AllInterfaceNumberingIpv4Types(), "`,`")),
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv4TypeNone)),
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllInterfaceNumberingIpv4Types()...)},
		},
		"generic_ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address in CIDR notation.",
			Optional:            true,
			CustomType:          cidrtypes.IPv4PrefixType{},
			Validators: []validator.String{
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("generic_ipv4_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv4TypeNumbered))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("generic_ipv4_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv4TypeNone))),
				stringvalidator.AlsoRequires(path.MatchRoot("generic_ipv4_address_type")),
			},
		},
		"generic_ipv6_address_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Allowed values: [`%s`]", strings.Join(utils.AllInterfaceNumberingIpv6Types(), "`,`")),
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeNone)),
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllInterfaceNumberingIpv6Types()...)},
		},
		"generic_ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address in CIDR notation.",
			Optional:            true,
			CustomType:          cidrtypes.IPv6PrefixType{},
			Validators: []validator.String{
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("generic_ipv6_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeNumbered))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("generic_ipv6_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeNone))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("generic_ipv6_address_type"), types.StringValue(utils.StringersToFriendlyString(apstra.InterfaceNumberingIpv6TypeLinkLocal))),
				stringvalidator.AlsoRequires(path.MatchRoot("generic_ipv6_address_type")),
			},
		},
	}
}

func requestEndpoint(v4type, v6type types.String, v4addr cidrtypes.IPv4Prefix, v6addr cidrtypes.IPv6Prefix, attrPrefix string, diags *diag.Diagnostics) apstra.TwoStageL3ClosSubinterface {
	var result apstra.TwoStageL3ClosSubinterface

	if !v4type.IsNull() {
		err := utils.ApiStringerFromFriendlyString(&result.Ipv4AddrType, v4type.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root(attrPrefix+"_ipv4_address_type"), "Cannot parse ipv4 address type", err.Error())
		}
	}

	if !v6type.IsNull() {
		err := utils.ApiStringerFromFriendlyString(&result.Ipv6AddrType, v6type.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root(attrPrefix+"_ipv6_address_type"), "Cannot parse ipv6 address type", err.Error())
		}
	}

	var err error

	if !v4addr.IsNull() {
		var ip net.IP
		ip, result.Ipv4Addr, err = net.ParseCIDR(v4addr.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root(attrPrefix+"_ipv4_address"), "Cannot parse ipv4 address", err.Error())
		}
		result.Ipv4Addr.IP = ip
	}

	if !v6addr.IsNull() {
		var ip net.IP
		ip, result.Ipv6Addr, err = net.ParseCIDR(v6addr.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root(attrPrefix+"_ipv6_address"), "Cannot parse ipv6 address", err.Error())
		}
		result.Ipv6Addr.IP = ip
	}

	return result
}

func (o IpLinkAddressing) Request(_ context.Context, diags *diag.Diagnostics) map[apstra.ObjectId]apstra.TwoStageL3ClosSubinterface {
	if !utils.HasValue(o.SwitchIntfId) || !utils.HasValue(o.GenericIntfId) {
		diags.AddError(
			constants.ErrProviderBug,
			fmt.Sprintf("attempt to generate ip link addressing with unknown interface ID\n"+
				"switch_interface_id: %s\n generic_interface_id: %s", o.SwitchIntfId, o.GenericIntfId),
		)
	}

	result := make(map[apstra.ObjectId]apstra.TwoStageL3ClosSubinterface, 2)
	result[apstra.ObjectId(o.SwitchIntfId.ValueString())] = requestEndpoint(o.SwitchIpv4Type, o.SwitchIpv6Type, o.SwitchIpv4Addr, o.SwitchIpv6Addr, "switch", diags)
	result[apstra.ObjectId(o.GenericIntfId.ValueString())] = requestEndpoint(o.GenericIpv4Type, o.GenericIpv6Type, o.GenericIpv4Addr, o.GenericIpv6Addr, "generic", diags)

	return result
}

func epBySubinterfaceId(siId apstra.ObjectId, eps []apstra.TwoStageL3ClosSubinterfaceLinkEndpoint, diags *diag.Diagnostics) *apstra.TwoStageL3ClosSubinterfaceLinkEndpoint {
	var result *apstra.TwoStageL3ClosSubinterfaceLinkEndpoint
	for _, ep := range eps {
		ep := ep
		if ep.SubinterfaceId == siId {
			if result != nil {
				diags.AddError(
					"Unexpected API response",
					fmt.Sprintf("Logical Link has multiple endpoints on with subinterface ID %q", siId),
				)
				return nil
			}

			result = &ep
		}
	}

	if result == nil {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("Link previously had a subinterface with ID %q, but that interface has gone missing", siId),
		)
	}

	return result
}

func epBySystemType(sysType apstra.SystemType, eps []apstra.TwoStageL3ClosSubinterfaceLinkEndpoint, diags *diag.Diagnostics) *apstra.TwoStageL3ClosSubinterfaceLinkEndpoint {
	var systemRoles []apstra.SystemRole

	switch sysType {
	case apstra.SystemTypeSwitch:
		systemRoles = []apstra.SystemRole{apstra.SystemRoleSuperSpine, apstra.SystemRoleSpine, apstra.SystemRoleLeaf, apstra.SystemRoleAccess}
	case apstra.SystemTypeServer:
		systemRoles = []apstra.SystemRole{apstra.SystemRoleGeneric}
	default:
		diags.AddError(constants.ErrProviderBug, fmt.Sprintf("unexpected system type %q", sysType))
		return nil
	}

	var result *apstra.TwoStageL3ClosSubinterfaceLinkEndpoint
	for _, ep := range eps {
		ep := ep
		if utils.SliceContains(ep.System.Role, systemRoles) {
			if result != nil {
				diags.AddError(
					"Unexpected API response",
					fmt.Sprintf("Logical link has multiple endpoints on systems with %q roles", sysType),
				)
				return nil
			}

			result = &ep
		}
	}

	if result == nil {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("Logical link has no endpoints on systems with %q roles", sysType),
		)
	}

	return result
}

func (o *IpLinkAddressing) LoadApiData(_ context.Context, in *apstra.TwoStageL3ClosSubinterfaceLink, diags *diag.Diagnostics) {
	// ensure 2 endpoints
	if len(in.Endpoints) != 2 {
		diags.AddError("Unexpected API response", fmt.Sprintf("Logical links should have 2 endpoints, got %d", len(in.Endpoints)))
		return
	}

	// extract the endpoint subinterface IDs
	siIds := make([]apstra.ObjectId, 2)
	for i, ep := range in.Endpoints {
		siIds[i] = ep.SubinterfaceId
	}

	// ensure endpoint subinterface IDs are different
	if siIds[0] == siIds[1] {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("Logical link %q has two endpoints with identical subinterface ID %q", in.Id, siIds[0]),
		)
		return
	}

	// ensure the subinterface IDs have the expected values
	if !utils.ItemInSlice(apstra.ObjectId(o.GenericIntfId.ValueString()), siIds) ||
		!utils.ItemInSlice(apstra.ObjectId(o.SwitchIntfId.ValueString()), siIds) {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("Logical link %s previously had subinterface IDs %s and %s.\n"+
				"Now it has IDs %q and %q. Endpoint IDs are not expected to change.",
				o.LinkId, o.SwitchIntfId, o.GenericIntfId, siIds[0], siIds[1]),
		)
		return
	}

	// extract the endpoints by subinterface ID
	switchEp := epBySubinterfaceId(apstra.ObjectId(o.SwitchIntfId.ValueString()), in.Endpoints, diags)
	genericEp := epBySubinterfaceId(apstra.ObjectId(o.GenericIntfId.ValueString()), in.Endpoints, diags)
	if diags.HasError() {
		return
	}

	// load the API data from each endpoint
	o.SwitchIpv4Type = types.StringValue(utils.StringersToFriendlyString(switchEp.Subinterface.Ipv4AddrType))
	o.SwitchIpv4Addr = cidrtypes.NewIPv4PrefixNull()
	if switchEp.Subinterface.Ipv4Addr != nil {
		o.SwitchIpv4Addr = cidrtypes.NewIPv4PrefixValue(switchEp.Subinterface.Ipv4Addr.String())
	}
	o.SwitchIpv6Type = types.StringValue(utils.StringersToFriendlyString(switchEp.Subinterface.Ipv6AddrType))
	o.SwitchIpv6Addr = cidrtypes.NewIPv6PrefixNull()
	if switchEp.Subinterface.Ipv6Addr != nil {
		o.SwitchIpv6Addr = cidrtypes.NewIPv6PrefixValue(switchEp.Subinterface.Ipv6Addr.String())
	}

	o.GenericIpv4Type = types.StringValue(utils.StringersToFriendlyString(genericEp.Subinterface.Ipv4AddrType))
	o.GenericIpv4Addr = cidrtypes.NewIPv4PrefixNull()
	if genericEp.Subinterface.Ipv4Addr != nil {
		o.GenericIpv4Addr = cidrtypes.NewIPv4PrefixValue(genericEp.Subinterface.Ipv4Addr.String())
	}
	o.GenericIpv6Type = types.StringValue(utils.StringersToFriendlyString(genericEp.Subinterface.Ipv6AddrType))
	o.GenericIpv6Addr = cidrtypes.NewIPv6PrefixNull()
	if genericEp.Subinterface.Ipv6Addr != nil {
		o.GenericIpv6Addr = cidrtypes.NewIPv6PrefixValue(genericEp.Subinterface.Ipv6Addr.String())
	}
}

// LoadImmutableData sets the switch and generic subinterface ID elements within o and saves the
// initial switch/generic v4/v6 address types (probably those indicated in the connectivity template).
// The user supplies the link ID, so we only need to do this once: at the beginning of Create(). The
// subinterface nodes associated with a given link node should never change.
func (o *IpLinkAddressing) LoadImmutableData(ctx context.Context, in *apstra.TwoStageL3ClosSubinterfaceLink, resp *resource.CreateResponse) {
	switchEp := epBySystemType(apstra.SystemTypeSwitch, in.Endpoints, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	genericEp := epBySystemType(apstra.SystemTypeServer, in.Endpoints, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	o.SwitchIntfId = types.StringValue(switchEp.SubinterfaceId.String())
	o.GenericIntfId = types.StringValue(genericEp.SubinterfaceId.String())

	private, err := json.Marshal(struct {
		SwitchIpv4AddressType  string `json:"switch_ipv4_address_type"`
		SwitchIpv6AddressType  string `json:"switch_ipv6_address_type"`
		GenericIpv4AddressType string `json:"generic_ipv4_address_type"`
		GenericIpv6AddressType string `json:"generic_ipv6_address_type"`
	}{
		SwitchIpv4AddressType:  utils.StringersToFriendlyString(switchEp.Subinterface.Ipv4AddrType),
		SwitchIpv6AddressType:  utils.StringersToFriendlyString(switchEp.Subinterface.Ipv6AddrType),
		GenericIpv4AddressType: utils.StringersToFriendlyString(genericEp.Subinterface.Ipv4AddrType),
		GenericIpv6AddressType: utils.StringersToFriendlyString(genericEp.Subinterface.Ipv6AddrType),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed marshaling private data", err.Error())
		return
	}

	resp.Private.SetKey(ctx, "ep_addr_types", private)
}
