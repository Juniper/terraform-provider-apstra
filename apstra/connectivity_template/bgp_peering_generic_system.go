package connectivitytemplate

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ Primitive = &BgpPeeringGenericSystem{}

type BgpPeeringGenericSystem struct {
	Name               types.String `tfsdk:"name"`
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
	ChildPrimitives    types.Set    `tfsdk:"child_primitives"`
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
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Primitive name displayed in the web UI",
			Optional:            true,
		},
		"ttl": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP Time To Live. Omit to use device defaults.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(1, math.MaxUint8)},
		},
		"bfd_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enable BFD.",
			Optional:            true,
		},
		"password": dataSourceSchema.StringAttribute{
			MarkdownDescription: "BGP TCP authentication password.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"keepalive_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP keepalive time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, math.MaxUint16),
				int64validator.AlsoRequires(path.MatchRoot("hold_time")),
			},
		},
		"hold_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, math.MaxUint16),
				int64validator.AlsoRequires(path.MatchRoot("keepalive_time")),
				apstravalidator.AtLeastProductOf(3, path.MatchRoot("keepalive_time")),
			},
		},
		"ipv4_addressing_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of `%s` (or omit)",
				strings.Join(ipv4AddressingTypes, "`, `"),
			),
			Optional:   true,
			Validators: []validator.String{stringvalidator.OneOf(ipv4AddressingTypes...)},
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
			Validators: []validator.Int64{int64validator.Between(1, math.MaxUint32)},
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
		"child_primitives": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of JSON strings describing Connectivity Template Primitives " +
				"which are children of this Connectivity Template Primitive. Use the `primitive` " +
				"attribute of other Connectivity Template Primitives data sources here.",
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
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
	obj := bgpPeeringGenericSystemPrototype{
		Ipv4AfiEnabled:     !o.Ipv4AddressingType.IsNull() && o.Ipv4AddressingType.ValueString() != apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone.String(),
		Ipv6AfiEnabled:     !o.Ipv6AddressingType.IsNull() && o.Ipv6AddressingType.ValueString() != apstra.CtPrimitiveIPv6ProtocolSessionAddressingNone.String(),
		Ttl:                uint8(o.Ttl.ValueInt64()),
		BfdEnabled:         o.BfdEnabled.ValueBool(),
		Password:           o.Password.ValueStringPointer(),
		Ipv4AddressingType: o.Ipv4AddressingType.ValueString(),
		Ipv6AddressingType: o.Ipv6AddressingType.ValueString(),
		NeighborAsnDynamic: o.NeighborAsnDynamic.ValueBool(),
		PeerFromLoopback:   o.PeerFromLoopback.ValueBool(),
		PeerTo:             o.PeerTo.ValueString(), // see below
	}

	if !o.KeepaliveTime.IsNull() {
		t := uint16(o.KeepaliveTime.ValueInt64())
		obj.KeepaliveTime = &t
	}

	if !o.HoldTime.IsNull() {
		t := uint16(o.HoldTime.ValueInt64())
		obj.HoldTime = &t
	}

	if obj.Ipv4AddressingType == "" { // set default for omitted attribute
		obj.Ipv4AddressingType = apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone.String()
	}

	if obj.Ipv6AddressingType == "" { // set default for omitted attribute
		obj.Ipv6AddressingType = apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone.String()
	}

	if !o.LocalAsn.IsNull() {
		la := uint32(o.LocalAsn.ValueInt64())
		obj.LocalAsn = &la
	}

	if obj.PeerTo == "" { // set default for omitted attribute
		obj.PeerTo = apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint.String()
	}

	var childPrimitives []string
	diags.Append(o.ChildPrimitives.ElementsAs(ctx, &childPrimitives, false)...)
	if diags.HasError() {
		return ""
	}

	// sort the childPrimitives by their SHA1 sums for easier comparison of nested strings
	sort.Slice(childPrimitives, func(i, j int) bool {
		sum1 := sha1.Sum([]byte(childPrimitives[i]))
		sum2 := sha1.Sum([]byte(childPrimitives[j]))
		return bytes.Compare(sum1[:], sum2[:]) >= 0
	})

	obj.ChildPrimitives = childPrimitives

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling BgpPeeringGenericSystem primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachBgpOverSubinterfacesOrSvi.String(),
		Label:         o.Name.ValueString(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *BgpPeeringGenericSystem) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpOverSubinterfacesOrSvi)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	o.Ttl = types.Int64Value(int64(attributes.Ttl))
	o.BfdEnabled = types.BoolValue(attributes.Bfd)
	if attributes.Password != nil {
		o.Password = types.StringValue(*attributes.Password)
	} else {
		o.Password = types.StringNull()
	}
	if attributes.Keepalive != nil {
		o.KeepaliveTime = types.Int64Value(int64(*attributes.Keepalive))
	} else {
		o.KeepaliveTime = types.Int64Null()
	}
	if attributes.Holdtime != nil {
		o.HoldTime = types.Int64Value(int64(*attributes.Holdtime))
	} else {
		o.HoldTime = types.Int64Null()
	}
	o.Ipv4AddressingType = types.StringValue(attributes.SessionAddressingIpv4.String())
	o.Ipv6AddressingType = types.StringValue(attributes.SessionAddressingIpv6.String())
	if attributes.LocalAsn != nil {
		o.LocalAsn = types.Int64Value(int64(*attributes.LocalAsn))
	} else {
		o.LocalAsn = types.Int64Null()
	}
	o.NeighborAsnDynamic = types.BoolValue(attributes.NeighborAsnDynamic)
	o.PeerFromLoopback = types.BoolValue(attributes.PeerFromLoopback)
	o.PeerTo = types.StringValue(attributes.PeerTo.String())
	o.ChildPrimitives = utils.SetValueOrNull(ctx, types.StringType, SdkPrimitivesToJsonStrings(ctx, in.Subpolicies, diags), diags)
	o.Name = types.StringValue(in.Label)
}

var _ JsonPrimitive = &bgpPeeringGenericSystemPrototype{}

type bgpPeeringGenericSystemPrototype struct {
	Label              string   `json:"label,omitempty"`
	Ipv4AfiEnabled     bool     `json:"ipv4_afi_enabled"`
	Ipv6AfiEnabled     bool     `json:"ipv6_afi_enabled"`
	Ttl                uint8    `json:"ttl"`
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
	ChildPrimitives    []string `json:"child_primitives"`
}

func (o bgpPeeringGenericSystemPrototype) attributes(_ context.Context, path path.Path, diags *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	var err error
	var ipv4AddressingType apstra.CtPrimitiveIPv4ProtocolSessionAddressing
	err = ipv4AddressingType.FromString(o.Ipv4AddressingType)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv4 addressing type %q", o.Ipv4AddressingType), err.Error())
		return nil
	}

	var ipv6AddressingType apstra.CtPrimitiveIPv6ProtocolSessionAddressing
	err = ipv6AddressingType.FromString(o.Ipv6AddressingType)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv6 addressing type %q", o.Ipv6AddressingType), err.Error())
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
		Ttl:                   o.Ttl,
	}
}

func (o bgpPeeringGenericSystemPrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	attributes := o.attributes(ctx, path, diags)
	if diags.HasError() {
		return nil
	}

	childPrimitives := ChildPrimitivesFromListOfJsonStrings(ctx, o.ChildPrimitives, path, diags)
	if diags.HasError() {
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitive{
		Id:          nil, // calculated later
		Label:       o.Label,
		Attributes:  attributes,
		Subpolicies: childPrimitives,
		BatchId:     nil, // calculated later
		PipelineId:  nil, // calculated later
	}
}
