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
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
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

type DynamicBgpPeering struct {
	Name            types.String         `tfsdk:"name"`
	Ttl             types.Int64          `tfsdk:"ttl"`
	BfdEnabled      types.Bool           `tfsdk:"bfd_enabled"`
	Password        types.String         `tfsdk:"password"`
	KeepaliveTime   types.Int64          `tfsdk:"keepalive_time"`
	HoldTime        types.Int64          `tfsdk:"hold_time"`
	Ipv4Enabled     types.Bool           `tfsdk:"ipv4_enabled"`
	Ipv6Enabled     types.Bool           `tfsdk:"ipv6_enabled"`
	LocalAsn        types.Int64          `tfsdk:"local_asn"`
	Ipv4PeerPrefix  cidrtypes.IPv4Prefix `tfsdk:"ipv4_peer_prefix"`
	Ipv6PeerPrefix  cidrtypes.IPv6Prefix `tfsdk:"ipv6_peer_prefix"`
	RoutingPolicies types.Set            `tfsdk:"routing_policies"`
}

func (o DynamicBgpPeering) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":             types.StringType,
		"ttl":              types.Int64Type,
		"bfd_enabled":      types.BoolType,
		"password":         types.StringType,
		"keepalive_time":   types.Int64Type,
		"hold_time":        types.Int64Type,
		"ipv4_enabled":     types.BoolType,
		"ipv6_enabled":     types.BoolType,
		"local_asn":        types.Int64Type,
		"ipv4_peer_prefix": cidrtypes.IPv4PrefixType{},
		"ipv6_peer_prefix": cidrtypes.IPv6PrefixType{},
		"routing_policies": types.SetType{ElemType: types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}},
	}
}

func (o DynamicBgpPeering) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used by the web UI on the Primitive \"block\" in the Connectivity Template.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
		"ipv4_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables peering with IPv4 neighbors.",
			Required:            true,
		},
		"ipv6_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables peering with IPv6 neighbors.",
			Required:            true,
		},
		"local_asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a router " +
				"to appear to be a member of a second autonomous system (AS) by prepending a local-as " +
				"AS number, in addition to its real AS number, announced to its eBGP peer, resulting " +
				"in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(1, math.MaxUint32)},
		},
		"ipv4_peer_prefix": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 Subnet for BGP Prefix Dynamic Neighbors. Leave blank to derive subnet from application point.",
			CustomType:          cidrtypes.IPv4PrefixType{},
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv4_enabled"), types.BoolNull()),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv4_enabled"), types.BoolValue(false)),
			},
		},
		"ipv6_peer_prefix": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 Subnet for BGP Prefix Dynamic Neighbors. Leave blank to derive subnet from application point.",
			CustomType:          cidrtypes.IPv6PrefixType{},
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv6_enabled"), types.BoolNull()),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv6_enabled"), types.BoolValue(false)),
			},
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

func (o DynamicBgpPeering) ValidateConfig(_ context.Context, path path.Path, diags *diag.Diagnostics) {
	if !o.Ipv4Enabled.ValueBool() && !o.Ipv6Enabled.ValueBool() {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			path, "at least one of 'ipv4_enabled' and 'ipv6_enabled' must be true.",
		))
	}
}

func (o DynamicBgpPeering) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface {
	var holdTime *uint16
	if !o.HoldTime.IsNull() {
		holdTime = utils.ToPtr(uint16(o.HoldTime.ValueInt64()))
	}

	var ipv4PeerPrefix *net.IPNet
	if !o.Ipv4PeerPrefix.IsNull() {
		_, ipv4PeerPrefix, _ = net.ParseCIDR(o.Ipv4PeerPrefix.ValueString())
	}

	var ipv6PeerPrefix *net.IPNet
	if !o.Ipv6PeerPrefix.IsNull() {
		_, ipv6PeerPrefix, _ = net.ParseCIDR(o.Ipv6PeerPrefix.ValueString())
	}

	var keepaliveTime *uint16
	if !o.KeepaliveTime.IsNull() {
		keepaliveTime = utils.ToPtr(uint16(o.KeepaliveTime.ValueInt64()))
	}

	var localAsn *uint32
	if !o.LocalAsn.IsNull() {
		localAsn = utils.ToPtr(uint32(o.LocalAsn.ValueInt64()))
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface{
		Label:                 o.Name.ValueString(), // todo is this necessary?
		Bfd:                   o.BfdEnabled.ValueBool(),
		Holdtime:              holdTime,
		Ipv4Safi:              o.Ipv4Enabled.ValueBool(),
		Ipv6Safi:              o.Ipv6Enabled.ValueBool(),
		Keepalive:             keepaliveTime,
		LocalAsn:              localAsn,
		Password:              o.Password.ValueStringPointer(),
		PrefixNeighborIpv4:    ipv4PeerPrefix,
		PrefixNeighborIpv6:    ipv6PeerPrefix,
		SessionAddressingIpv4: o.Ipv4Enabled.ValueBool(),
		SessionAddressingIpv6: o.Ipv6Enabled.ValueBool(),
		Ttl:                   uint8(o.Ttl.ValueInt64()), // okay if null, then we get zero value
	}
}

func (o DynamicBgpPeering) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	result := apstra.ConnectivityTemplatePrimitive{
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(ctx, diags),
		// Subpolicies: // set below
	}

	result.Subpolicies = append(result.Subpolicies, RoutingPolicySubpolicies(ctx, o.RoutingPolicies, diags)...)

	return &result
}

func DynamicBgpPeeringSubpolicies(ctx context.Context, dynamicBgpPeeringSet types.Set, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var dynamicBgpPeerings []DynamicBgpPeering
	diags.Append(dynamicBgpPeeringSet.ElementsAs(ctx, &dynamicBgpPeerings, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(dynamicBgpPeerings))
	for i, dynamicBgpPeering := range dynamicBgpPeerings {
		subpolicies[i] = dynamicBgpPeering.primitive(ctx, diags)
	}

	return subpolicies
}

func newDynamicBgpPeering(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface, _ *diag.Diagnostics) DynamicBgpPeering {
	result := DynamicBgpPeering{
		// Name:        // handled by caller
		// Ttl:         // handled below due to 0 = null logic
		BfdEnabled:     types.BoolValue(in.Bfd),
		Password:       types.StringPointerValue(in.Password),
		KeepaliveTime:  utils.Int64PointerValue(in.Keepalive),
		HoldTime:       utils.Int64PointerValue(in.Holdtime),
		Ipv4Enabled:    types.BoolValue(in.SessionAddressingIpv4),
		Ipv6Enabled:    types.BoolValue(in.SessionAddressingIpv6),
		LocalAsn:       utils.Int64PointerValue(in.LocalAsn),
		Ipv4PeerPrefix: utils.Ipv4PrefixPointerValue(in.PrefixNeighborIpv4),
		Ipv6PeerPrefix: utils.Ipv6PrefixPointerValue(in.PrefixNeighborIpv6),
		// RoutingPolicies: types.Set{}, // set after this function is invoked
	}

	if in.Ttl > 0 {
		result.Ttl = types.Int64Value(int64(in.Ttl))
	}

	return result
}

func DynamicBgpPeeringPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []DynamicBgpPeering

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting DynamicBgpPeering primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newDynamicBgpPeering(ctx, p, diags)
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			newPrimitive.RoutingPolicies = RoutingPolicyPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()}, result, diags)
}
