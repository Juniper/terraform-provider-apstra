package connectivitytemplate

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math"
	"net"
	"sort"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/utils"
)

var _ Primitive = &BgpPeeringIpEndpoint{}

type BgpPeeringIpEndpoint struct {
	Label           types.String `tfsdk:"label"`
	NeighborAsn     types.Int64  `tfsdk:"neighbor_asn"`
	Ttl             types.Int64  `tfsdk:"ttl"`
	BfdEnabled      types.Bool   `tfsdk:"bfd_enabled"`
	Password        types.String `tfsdk:"password"`
	KeepaliveTime   types.Int64  `tfsdk:"keepalive_time"`
	HoldTime        types.Int64  `tfsdk:"hold_time"`
	LocalAsn        types.Int64  `tfsdk:"local_asn"`
	Ipv4Address     types.String `tfsdk:"ipv4_address"`
	Ipv6Address     types.String `tfsdk:"ipv6_address"`
	ChildPrimitives types.Set    `tfsdk:"child_primitives"`
	Primitive       types.String `tfsdk:"primitive"`
}

func (o BgpPeeringIpEndpoint) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"label": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Primitive label displayed in the web UI",
			Optional:            true,
		},
		"neighbor_asn": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Neighbor ASN. Omit for *Neighbor ASN Type Dynamic*.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(0, math.MaxUint32+1)},
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
			Validators: []validator.Int64{
				int64validator.Between(0, math.MaxUint16+1),
				int64validator.AlsoRequires(path.MatchRoot("hold_time")),
			},
		},
		"hold_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(0, math.MaxUint16+1),
				int64validator.AlsoRequires(path.MatchRoot("keepalive_time")),
				apstravalidator.AtLeastProductOf(3, path.MatchRoot("keepalive_time")),
			},
		},
		"local_asn": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a router " +
				"to appear to be a member of a second autonomous system (AS) by prepending a local-as " +
				"AS number, in addition to its real AS number, announced to its eBGP peer, resulting " +
				"in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(0, math.MaxUint32+1)},
		},
		"ipv4_address": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address of peer",
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ParseIp(true, false),
				stringvalidator.AtLeastOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("ipv6_address"),
				}...),
			},
		},
		"ipv6_address": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address of peer",
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ParseIp(false, true),
			},
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

func (o BgpPeeringIpEndpoint) Marshal(ctx context.Context, diags *diag.Diagnostics) string {
	prototype := bgpPeeringIpEndpointPrototype{
		Ipv4AfiEnabled:  !o.Ipv4Address.IsNull(),
		Ipv6AfiEnabled:  !o.Ipv6Address.IsNull(),
		Ttl:             uint8(o.Ttl.ValueInt64()), // okay if null, then we get zero value
		BfdEnabled:      o.BfdEnabled.ValueBool(),
		Password:        o.Password.ValueStringPointer(),
		Ipv4Address:     o.Ipv4Address.ValueStringPointer(),
		Ipv6Address:     o.Ipv6Address.ValueStringPointer(),
		ChildPrimitives: nil,
	}

	if o.NeighborAsn.IsNull() {
		prototype.NeighborAsnDynaimc = true
	} else {
		nasn := uint32(o.NeighborAsn.ValueInt64())
		prototype.NeighborAsn = &nasn
	}

	if !o.KeepaliveTime.IsNull() {
		kat := uint16(o.KeepaliveTime.ValueInt64())
		prototype.KeepaliveTime = &kat
	}

	if !o.HoldTime.IsNull() {
		ht := uint16(o.HoldTime.ValueInt64())
		prototype.HoldTime = &ht
	}

	if !o.LocalAsn.IsNull() {
		lasn := uint32(o.LocalAsn.ValueInt64())
		prototype.LocalAsn = &lasn
	}

	diags.Append(o.ChildPrimitives.ElementsAs(ctx, &prototype.ChildPrimitives, false)...)
	if diags.HasError() {
		return ""
	}

	// sort the childPrimitives by their SHA1 sums for easier comparison of nested strings
	sort.Slice(prototype.ChildPrimitives, func(i, j int) bool {
		sum1 := sha1.Sum([]byte(prototype.ChildPrimitives[i]))
		sum2 := sha1.Sum([]byte(prototype.ChildPrimitives[j]))
		return bytes.Compare(sum1[:], sum2[:]) >= 0
	})

	data, err := json.Marshal(&prototype)
	if err != nil {
		diags.AddError("failed marshaling BgpPeeringIpEndpoint primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachIpEndpointWithBgpNsxt.String(),
		Label:         o.Label.ValueString(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *BgpPeeringIpEndpoint) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	if attributes.Asn != nil {
		o.NeighborAsn = types.Int64Value(int64(*attributes.Asn))
	}

	o.Ttl = types.Int64Value(int64(attributes.Ttl))
	o.BfdEnabled = types.BoolValue(attributes.Bfd)

	if attributes.Password != nil {
		o.Password = types.StringValue(*attributes.Password)
	}

	if attributes.Keepalive != nil {
		o.KeepaliveTime = types.Int64Value(int64(*attributes.Keepalive))
	}

	if attributes.Holdtime != nil {
		o.HoldTime = types.Int64Value(int64(*attributes.Holdtime))
	}

	if attributes.LocalAsn != nil {
		o.LocalAsn = types.Int64Value(int64(*attributes.LocalAsn))
	}

	if attributes.Ipv4Addr != nil {
		o.Ipv4Address = types.StringValue(attributes.Ipv4Addr.String())
	}

	if attributes.Ipv6Addr != nil {
		o.Ipv6Address = types.StringValue(attributes.Ipv6Addr.String())
	}

	o.ChildPrimitives = utils.SetValueOrNull(ctx, types.StringType, SdkPrimitivesToJsonStrings(ctx, in.Subpolicies, diags), diags)
	o.Label = types.StringValue(in.Label)
}

var _ JsonPrimitive = &bgpPeeringIpEndpointPrototype{}

type bgpPeeringIpEndpointPrototype struct {
	Label              string   `json:"label,omitempty"`
	NeighborAsn        *uint32  `json:"neighbor_asn"`
	NeighborAsnDynaimc bool     `json:"neighbor_asn_dynaimc"`
	Ipv4AfiEnabled     bool     `json:"ipv4_afi_enabled"`
	Ipv6AfiEnabled     bool     `json:"ipv6_afi_enabled"`
	Ttl                uint8    `json:"ttl"`
	BfdEnabled         bool     `json:"bfd_enabled"`
	Password           *string  `json:"password"`
	KeepaliveTime      *uint16  `json:"keepalive_time"`
	HoldTime           *uint16  `json:"hold_time"`
	LocalAsn           *uint32  `json:"local_asn"`
	Ipv4Address        *string  `json:"ipv4_address"`
	Ipv6Address        *string  `json:"ipv6_address"`
	ChildPrimitives    []string `json:"child_primitives"`
}

func (o bgpPeeringIpEndpointPrototype) attributes(_ context.Context, _ path.Path, _ *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	var ipv4Addr, ipv6Addr net.IP
	if o.Ipv4Address != nil {
		ipv4Addr = net.ParseIP(*o.Ipv4Address)
	}
	if o.Ipv6Address != nil {
		ipv6Addr = net.ParseIP(*o.Ipv6Address)
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt{
		Asn:                o.NeighborAsn,
		Bfd:                o.BfdEnabled,
		Holdtime:           o.HoldTime,
		Ipv4Addr:           ipv4Addr,
		Ipv6Addr:           ipv6Addr,
		Ipv4Safi:           o.Ipv4AfiEnabled,
		Ipv6Safi:           o.Ipv6AfiEnabled,
		Keepalive:          o.KeepaliveTime,
		LocalAsn:           o.LocalAsn,
		NeighborAsnDynamic: o.NeighborAsnDynaimc,
		Password:           o.Password,
		Ttl:                o.Ttl,
	}
}

func (o bgpPeeringIpEndpointPrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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
