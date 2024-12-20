package primitives

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BgpPeeringIpEndpoint struct {
	Name            types.String        `tfsdk:"name"`
	NeighborAsn     types.Int64         `tfsdk:"neighbor_asn"`
	Ttl             types.Int64         `tfsdk:"ttl"`
	BfdEnabled      types.Bool          `tfsdk:"bfd_enabled"`
	Password        types.String        `tfsdk:"password"`
	KeepaliveTime   types.Int64         `tfsdk:"keepalive_time"`
	HoldTime        types.Int64         `tfsdk:"hold_time"`
	LocalAsn        types.Int64         `tfsdk:"local_asn"`
	Ipv4Address     iptypes.IPv4Address `tfsdk:"ipv4_address"`
	Ipv6Address     iptypes.IPv6Address `tfsdk:"ipv6_address"`
	RoutingPolicies types.Set           `tfsdk:"routing_policies"`
}

func (o BgpPeeringIpEndpoint) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":             types.StringType,
		"neighbor_asn":     types.Int64Type,
		"ttl":              types.Int64Type,
		"bfd_enabled":      types.BoolType,
		"password":         types.StringType,
		"keepalive_time":   types.Int64Type,
		"hold_time":        types.Int64Type,
		"local_asn":        types.Int64Type,
		"ipv4_address":     iptypes.IPv4AddressType{},
		"ipv6_address":     iptypes.IPv6AddressType{},
		"routing_policies": types.SetType{ElemType: types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}},
	}
}

func (o BgpPeeringIpEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used by the web UI on the Primitive \"block\" in the Connectivity Template.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"neighbor_asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "Neighbor ASN. Omit for *Neighbor ASN Type Dynamic*.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(constants.AsnMin, constants.AsnMax)},
		},
		"ttl": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP Time To Live. Omit to use device defaults.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(constants.TtlMin, constants.TtlMax)},
		},
		"bfd_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enable BFD.",
			Required:            true,
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "Password used to secure the BGP session.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"keepalive_time": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP keepalive time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(constants.BgpKeepaliveMin, constants.BgpKeepaliveMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("hold_time")),
			},
		},
		"hold_time": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(constants.BgpHoldMin, constants.BgpHoldMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("keepalive_time")),
				apstravalidator.AtLeastProductOf(3, path.MatchRelative().AtParent().AtName("keepalive_time")),
			},
		},
		"local_asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a router " +
				"to appear to be a member of a second autonomous system (AS) by prepending a local-as " +
				"AS number, in addition to its real AS number, announced to its eBGP peer, resulting " +
				"in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(1, math.MaxUint32)},
		},
		"ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address of peer.",
			CustomType:          iptypes.IPv4AddressType{},
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.AtLeastOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRelative().AtParent().AtName("ipv6_address"),
				}...),
			},
		},
		"ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address of peer.",
			CustomType:          iptypes.IPv6AddressType{},
			Optional:            true,
		},
		"routing_policies": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Routing Policy Primitives to be used with this *Protocol Endpoint*.",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: RoutingPolicy{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o BgpPeeringIpEndpoint) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt {
	var neighborAsn *uint32
	if !o.NeighborAsn.IsNull() {
		neighborAsn = utils.ToPtr(uint32(o.NeighborAsn.ValueInt64()))
	}

	var holdTime *uint16
	if !o.HoldTime.IsNull() {
		holdTime = utils.ToPtr(uint16(o.HoldTime.ValueInt64()))
	}

	var ipv4Addr net.IP
	if !o.Ipv4Address.IsNull() {
		ipv4Addr = net.ParseIP(o.Ipv4Address.ValueString())
	}

	var ipv6Addr net.IP
	if !o.Ipv6Address.IsNull() {
		ipv6Addr = net.ParseIP(o.Ipv6Address.ValueString())
	}

	var keepaliveTime *uint16
	if !o.KeepaliveTime.IsNull() {
		keepaliveTime = utils.ToPtr(uint16(o.KeepaliveTime.ValueInt64()))
	}

	var localAsn *uint32
	if !o.LocalAsn.IsNull() {
		localAsn = utils.ToPtr(uint32(o.LocalAsn.ValueInt64()))
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt{
		Label:              o.Name.ValueString(), // todo is this necessary?
		Asn:                neighborAsn,
		Bfd:                o.BfdEnabled.ValueBool(),
		Holdtime:           holdTime,
		Ipv4Addr:           ipv4Addr,
		Ipv6Addr:           ipv6Addr,
		Ipv4Safi:           !o.Ipv4Address.IsNull(),
		Ipv6Safi:           !o.Ipv6Address.IsNull(),
		Keepalive:          keepaliveTime,
		LocalAsn:           localAsn,
		NeighborAsnDynamic: o.NeighborAsn.IsNull(),
		Password:           o.Password.ValueStringPointer(),
		Ttl:                uint8(o.Ttl.ValueInt64()), // okay if null, then we get zero value
	}
}

func (o BgpPeeringIpEndpoint) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	result := apstra.ConnectivityTemplatePrimitive{
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(ctx, diags),
		// Subpolicies: // set below
	}

	result.Subpolicies = append(result.Subpolicies, RoutingPolicySubpolicies(ctx, o.RoutingPolicies, diags)...)

	return &result
}

func BgpPeeringIpEndpointSubpolicies(ctx context.Context, bgpPeeringIpEndpointSet types.Set, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var bgpPeeringIpEndpoints []BgpPeeringIpEndpoint
	diags.Append(bgpPeeringIpEndpointSet.ElementsAs(ctx, &bgpPeeringIpEndpoints, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(bgpPeeringIpEndpoints))
	for i, bgpPeeringIpEndpoint := range bgpPeeringIpEndpoints {
		subpolicies[i] = bgpPeeringIpEndpoint.primitive(ctx, diags)
	}

	return subpolicies
}

func newBgpPeeringIpEndpoint(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt, _ *diag.Diagnostics) BgpPeeringIpEndpoint {
	result := BgpPeeringIpEndpoint{
		// Name:       // handled by caller
		// Ttl:        // handled below due to 0 = null logic
		NeighborAsn:   utils.Int64PointerValue(in.Asn),
		BfdEnabled:    types.BoolValue(in.Bfd),
		Password:      types.StringPointerValue(in.Password),
		KeepaliveTime: utils.Int64PointerValue(in.Keepalive),
		HoldTime:      utils.Int64PointerValue(in.Holdtime),
		LocalAsn:      utils.Int64PointerValue(in.LocalAsn),
		Ipv4Address:   utils.Ipv4AddrValue(in.Ipv4Addr),
		Ipv6Address:   utils.Ipv6AddrValue(in.Ipv6Addr),
		// RoutingPolicies: handled by caller
	}

	if in.Ttl > 0 {
		result.Ttl = types.Int64Value(int64(in.Ttl))
	}

	return result
}

func BgpPeeringIpEndpointPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []BgpPeeringIpEndpoint

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting BgpPeeringIpEndpoint primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newBgpPeeringIpEndpoint(ctx, p, diags)
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			newPrimitive.RoutingPolicies = RoutingPolicyPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()}, result, diags)
}
