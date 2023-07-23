package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math"
	"strings"
)

type BgpPeeringGenericSystem struct {
	Ipv4AfiEnabled     types.Bool   `tfsdk:"ipv4_afi_enabled"`
	Ipv6AfiEnabled     types.Bool   `tfsdk:"ipv6_afi_enabled"`
	Ttl                types.Int64  `tfsdk:"ttl"`
	BfdEnabled         types.Bool   `tfsdk:"bfd_enabled"`
	Password           types.String `tfsdk:"password"`
	KeepaliveTime      types.Int64  `tfsdk:"keepalive_time"`
	HoldTime           types.Int64  `tfsdk:"hold_time"`
	Ipv4AddressingType types.String `tfsdk:"ipv4_addressing_type"`
	Ipv6AddressingType types.String `tfsdk:"ipv6_addressing_type"`
	LocalAsn           types.Int64  `tfsdk:"local_asn"`
	NeighborAsnDynamic types.Bool   `tfsdk:"neighbor_asn_dynamic"`
	PeerFromLoopback   types.Bool   `tfsdk:"peer_from_loopback"`
	PeerTo             types.String `tfsdk:"peer_to"`
	Children           types.List   `tfsdk:"children"`
	Primitive          types.String `tfsdk:"primitive"`
}

func (o BgpPeeringGenericSystem) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	ipv4AddressingTypes := []string{
		apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone.String(),
		apstra.CtPrimitiveIPv4ProtocolSessionAddressingAddressed.String(),
	}
	ipv6AddressingTypes := []string{
		apstra.CtPrimitiveIPv6ProtocolSessionAddressingNone.String(),
		apstra.CtPrimitiveIPv6ProtocolSessionAddressingAddressed.String(),
		apstra.CtPrimitiveIPv6ProtocolSessionAddressingLinkLocal.String(),
	}
	peerTo := []string{
		apstra.CtPrimitiveBgpPeerToLoopback.String(),
		apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint.String(),
		apstra.CtPrimitiveBgpPeerToInterfaceOrSharedIpEndpoint.String(),
	}
	return map[string]dataSourceSchema.Attribute{
		"ipv4_afi_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "IPv4 Address Family Identifier",
			Optional:            true,
			Validators: []validator.Bool{boolvalidator.AtLeastOneOf(path.Expressions{
				path.MatchRelative(),
				path.MatchRoot("ipv6_afi_enabled"),
			}...)},
		},
		"ipv6_afi_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "IPv6 Address Family Identifier",
			Optional:            true,
		},
		"ttl": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP Time To Live. Omit to use device defaults.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(0, math.MaxUint8+1)},
		},
		"bfd_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enable BFD.",
			Optional:            true,
		},
		"password": dataSourceSchema.StringAttribute{
			MarkdownDescription: "",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"keepalive_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP keepalive time (seconds).",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(0, math.MaxUint16+1)},
		},
		"hold_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(0, math.MaxUint16+1)},
		},
		"ipv4_addressing_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of `%s` (or omit)",
				strings.Join(ipv4AddressingTypes, "`, `"),
			),
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(ipv4AddressingTypes...),
				stringvalidator.AtLeastOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("ipv6_addressing_type"),
				}...),
			},
		},
		"ipv6_addressing_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of `%s` (or omit)",
				strings.Join(ipv6AddressingTypes, "`, `"),
			),
			Optional:   true,
			Validators: []validator.String{stringvalidator.OneOf(ipv6AddressingTypes...)},
		},
		"local_asn": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a " +
				"router to appear to be a member of a second autonomous system (AS) by prepending " +
				"a local-as AS number, in addition to its real AS number, announced to its eBGP " +
				"peer, resulting in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(0, math.MaxUint32+1)},
		},
		"neighbor_asn_dynamic": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Default behavior is `static`",
			Optional:            true,
		},
		"peer_from_loopback": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enable to peer from loopback interface. Default behavior peers from physical interface.",
			Optional:            true,
		},
		"peer_to": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of `%s` (or omit)",
				strings.Join(peerTo, "`, `"),
			),
			Optional:   true,
			Validators: []validator.String{stringvalidator.OneOf(peerTo...)},
		},
		"children": dataSourceSchema.ListAttribute{
			MarkdownDescription: "A list of JSON strings describing Connectivity Template Primitives " +
				"which are children of this Connectivity Template JsonPrimitive. Use the `primitive` " +
				"attribute of other Connectivity Template Primitives data sources here.",
			ElementType: types.StringType,
			Optional:    true,
		},
		"primitive": dataSourceSchema.StringAttribute{
			MarkdownDescription: "JSON output for use in the `primitives` field of an " +
				"`apstra_datacenter_connectivity_template` resource or a different Connectivity " +
				"Template JsonPrimitive data source",
			Computed: true,
		},
	}
}

func (o BgpPeeringGenericSystem) Marshal(ctx context.Context, diags *diag.Diagnostics) string {
	var children []string
	diags.Append(o.Children.ElementsAs(ctx, &children, false)...)
	if diags.HasError() {
		return ""
	}

	ttl := uint8(o.Ttl.ValueInt64())

	var keepaliveTime *uint16
	if !o.KeepaliveTime.IsNull() {
		t := uint16(o.KeepaliveTime.ValueInt64())
		keepaliveTime = &t
	}

	var holdTime *uint16
	if !o.HoldTime.IsNull() {
		t := uint16(o.HoldTime.ValueInt64())
		holdTime = &t
	}

	ipv4addressingType := apstra.CtPrimitiveIPv4AddressingTypeNone.String()
	if !o.Ipv4AddressingType.IsNull() {
		ipv4addressingType = o.Ipv4AddressingType.ValueString()
	}

	ipv6addressingType := apstra.CtPrimitiveIPv6AddressingTypeNone.String()
	if !o.Ipv6AddressingType.IsNull() {
		ipv6addressingType = o.Ipv6AddressingType.ValueString()
	}

	var localAsn *uint32
	if !o.LocalAsn.IsNull() {
		la := uint32(o.LocalAsn.ValueInt64())
		localAsn = &la
	}

	peerTo := apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint.String()
	if !o.PeerTo.IsNull() {
		peerTo = o.PeerTo.ValueString()
	}

	obj := bgpPeeringGenericSystemPrototype{
		Ipv4AfiEnabled:     o.Ipv4AfiEnabled.ValueBool(),
		Ipv6AfiEnabled:     o.Ipv6AfiEnabled.ValueBool(),
		Ttl:                &ttl,
		BfdEnabled:         o.BfdEnabled.ValueBool(),
		Password:           o.Password.ValueStringPointer(),
		KeepaliveTime:      keepaliveTime,
		HoldTime:           holdTime,
		Ipv4AddressingType: ipv4addressingType,
		Ipv6AddressingType: ipv6addressingType,
		LocalAsn:           localAsn,
		NeighborAsnDynamic: o.NeighborAsnDynamic.ValueBool(),
		PeerFromLoopback:   o.PeerFromLoopback.ValueBool(),
		PeerTo:             peerTo,
		Children:           children,
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling BgpPeeringGenericSystem primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&TfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachLogicalLink.String(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

var _ JsonPrimitive = &bgpPeeringGenericSystemPrototype{}

type bgpPeeringGenericSystemPrototype struct {
	Ipv4AfiEnabled     bool     `json:"ipv4_afi_enabled"`
	Ipv6AfiEnabled     bool     `json:"ipv6_afi_enabled"`
	Ttl                *uint8   `json:"ttl"`
	BfdEnabled         bool     `json:"bfd_enabled"`
	Password           *string  `json:"password"`
	KeepaliveTime      *uint16  `json:"keepalive_time"`
	HoldTime           *uint16  `json:"hold_time"`
	Ipv4AddressingType string   `json:"ipv4_addressing_type"`
	Ipv6AddressingType string   `json:"ipv6_addressing_type"`
	LocalAsn           *uint32  `json:"local_asn"`
	NeighborAsnDynamic bool     `json:"neighbor_asn_dynamic"`
	PeerFromLoopback   bool     `json:"peer_from_loopback"`
	PeerTo             string   `json:"peer_to"`
	Children           []string `json:"children"`
}

func (o bgpPeeringGenericSystemPrototype) attributes(_ context.Context, path path.Path, diags *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	var err error
	var ipv4AddressingType apstra.CtPrimitiveIPv4AddressingType
	err = ipv4AddressingType.FromString(o.Ipv4AddressingType)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv4 addressing type %s", o.Ipv4AddressingType), err.Error())
		return nil
	}

	var ipv6AddressingType apstra.CtPrimitiveIPv6AddressingType
	err = ipv6AddressingType.FromString(o.Ipv6AddressingType)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv6 addressing type %s", o.Ipv6AddressingType), err.Error())
		return nil
	}

	var peerTo apstra.CtPrimitiveBgpPeerTo
	err = peerTo.FromString(o.PeerTo)
	if err != nil {
		diags.AddAttributeError(path, "failed parsing peer_to", err.Error())
		return nil
	}

	var sessionAddressingIpv4 apstra.CtPrimitiveIPv4ProtocolSessionAddressing
	err = sessionAddressingIpv4.FromString(o.Ipv4AddressingType)
	if err != nil {
		diags.AddAttributeError(path, "failed parsing ipv4_addressing_type", err.Error())
		return nil
	}

	var sessionAddressingIpv6 apstra.CtPrimitiveIPv6ProtocolSessionAddressing
	err = sessionAddressingIpv6.FromString(o.Ipv6AddressingType)
	if err != nil {
		diags.AddAttributeError(path, "failed parsing ipv6_addressing_type", err.Error())
		return nil
	}

	var ttl uint8
	if o.Ttl != nil {
		ttl = *o.Ttl
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpOverSubinterfacesOrSvi{
		Bfd:                   o.BfdEnabled,
		Holdtime:              o.HoldTime,
		Ipv4Safi:              o.Ipv4AfiEnabled,
		Ipv6Safi:              o.Ipv6AfiEnabled,
		Keepalive:             o.KeepaliveTime,
		LocalAsn:              o.LocalAsn,
		NeighborAsnDynamic:    o.NeighborAsnDynamic,
		Password:              o.Password,
		PeerFromLoopback:      o.PeerFromLoopback,
		PeerTo:                peerTo,
		SessionAddressingIpv4: sessionAddressingIpv4,
		SessionAddressingIpv6: sessionAddressingIpv6,
		Ttl:                   ttl,
	}
}

func (o bgpPeeringGenericSystemPrototype) SdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	attributes := o.attributes(ctx, path, diags)
	if diags.HasError() {
		return nil
	}

	children := ChildPrimitivesFromListOfJsonStrings(ctx, o.Children, path, diags)
	if diags.HasError() {
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitive{
		Id:          nil, // calculated later
		Attributes:  attributes,
		Subpolicies: children,
		BatchId:     nil, // calculated later
		PipelineId:  nil, // calculated later
	}
}
