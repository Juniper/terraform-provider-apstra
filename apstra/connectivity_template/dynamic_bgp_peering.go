package connectivitytemplate

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
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
)

var _ Primitive = &DynamicBgpPeering{}

type DynamicBgpPeering struct {
	Name            types.String `tfsdk:"name"`
	Ttl             types.Int64  `tfsdk:"ttl"`
	BfdEnabled      types.Bool   `tfsdk:"bfd_enabled"`
	Password        types.String `tfsdk:"password"`
	KeepaliveTime   types.Int64  `tfsdk:"keepalive_time"`
	HoldTime        types.Int64  `tfsdk:"hold_time"`
	Ipv4Enabled     types.Bool   `tfsdk:"ipv4_enabled"`
	Ipv6Enabled     types.Bool   `tfsdk:"ipv6_enabled"`
	LocalAsn        types.Int64  `tfsdk:"local_asn"`
	Ipv4PeerPrefix  types.String `tfsdk:"ipv4_peer_prefix"`
	Ipv6PeerPrefix  types.String `tfsdk:"ipv6_peer_prefix"`
	ChildPrimitives types.Set    `tfsdk:"child_primitives"`
	Primitive       types.String `tfsdk:"primitive"`
}

func (o DynamicBgpPeering) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Primitive name displayed in the web UI",
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
		"ipv4_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enable to allow IPv4 peers.",
			Optional:            true,
			Validators: []validator.Bool{boolvalidator.AtLeastOneOf(
				path.MatchRelative(),
				path.MatchRoot("ipv6_enabled"),
			)},
		},
		"ipv6_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enable to allow IPv6 peers.",
			Optional:            true,
		},
		"local_asn": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a " +
				"router to appear to be a member of a second autonomous system (AS) by prepending " +
				"a local-as AS number, in addition to its real AS number, announced to its eBGP " +
				"peer, resulting in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(0, math.MaxUint32+1)},
		},
		"ipv4_peer_prefix": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Omit to derive prefix from the application point.",
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ParseCidr(true, false),
				apstravalidator.WhenValueSetString(
					apstravalidator.ValueAtMustBeString(
						path.MatchRoot("ipv4_enabled"), types.BoolValue(true), false)),
			},
		},
		"ipv6_peer_prefix": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Omit to derive prefix from the application point.",
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ParseCidr(false, true),
				apstravalidator.WhenValueSetString(
					apstravalidator.ValueAtMustBeString(
						path.MatchRoot("ipv4_enabled"), types.BoolValue(true), false)),
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

func (o DynamicBgpPeering) Marshal(ctx context.Context, diags *diag.Diagnostics) string {
	obj := dynamicBgpPeeringPrototype{
		Ipv4AfiEnabled: o.Ipv4Enabled.ValueBool(),
		Ipv6AfiEnabled: o.Ipv6Enabled.ValueBool(),
		Ttl:            uint8(o.Ttl.ValueInt64()), // okay if null, then we get zero value
		BfdEnabled:     o.BfdEnabled.ValueBool(),
		Password:       o.Password.ValueStringPointer(),
		Ipv4Enabled:    o.Ipv4Enabled.ValueBool(),
		Ipv6Enabled:    o.Ipv6Enabled.ValueBool(),
		Ipv4PeerPrefix: o.Ipv4PeerPrefix.ValueStringPointer(),
		Ipv6PeerPrefix: o.Ipv6PeerPrefix.ValueStringPointer(),
	}

	if !o.KeepaliveTime.IsNull() {
		kat := uint16(o.KeepaliveTime.ValueInt64())
		obj.KeepaliveTime = &kat
	}

	if !o.HoldTime.IsNull() {
		ht := uint16(o.HoldTime.ValueInt64())
		obj.HoldTime = &ht
	}

	if !o.LocalAsn.IsNull() {
		la := uint32(o.LocalAsn.ValueInt64())
		obj.LocalAsn = &la
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
		diags.AddError("failed marshaling DynamicBgpPeering primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachBgpWithPrefixPeeringForSviOrSubinterface.String(),
		Label:         o.Name.ValueString(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *DynamicBgpPeering) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
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

	o.Ipv4Enabled = types.BoolValue(attributes.SessionAddressingIpv4)
	o.Ipv6Enabled = types.BoolValue(attributes.SessionAddressingIpv6)

	if attributes.LocalAsn != nil {
		o.LocalAsn = types.Int64Value(int64(*attributes.LocalAsn))
	}

	if attributes.PrefixNeighborIpv4 != nil {
		o.Ipv4PeerPrefix = types.StringValue(attributes.PrefixNeighborIpv4.String())
	}

	if attributes.PrefixNeighborIpv6 != nil {
		o.Ipv6PeerPrefix = types.StringValue(attributes.PrefixNeighborIpv6.String())
	}

	o.ChildPrimitives = utils.SetValueOrNull(ctx, types.StringType, SdkPrimitivesToJsonStrings(ctx, in.Subpolicies, diags), diags)
	o.Name = types.StringValue(in.Label)
}

var _ JsonPrimitive = &dynamicBgpPeeringPrototype{}

type dynamicBgpPeeringPrototype struct {
	Label           string   `json:"label,omitempty"`
	Ipv4AfiEnabled  bool     `json:"ipv4_afi_enabled"`
	Ipv6AfiEnabled  bool     `json:"ipv6_afi_enabled"`
	Ttl             uint8    `json:"ttl"`
	BfdEnabled      bool     `json:"bfd_enabled"`
	Password        *string  `json:"password"`
	KeepaliveTime   *uint16  `json:"keepalive_time"`
	HoldTime        *uint16  `json:"hold_time"`
	Ipv4Enabled     bool     `json:"ipv4_enabled"`
	Ipv6Enabled     bool     `json:"ipv6_enabled"`
	LocalAsn        *uint32  `json:"local_asn"`
	Ipv4PeerPrefix  *string  `json:"ipv4_peer_prefix"`
	Ipv6PeerPrefix  *string  `json:"ipv6_peer_prefix"`
	ChildPrimitives []string `json:"child_primitives"`
}

func (o dynamicBgpPeeringPrototype) attributes(_ context.Context, path path.Path, diags *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	var ipv4Prefix, ipv6Prefix *net.IPNet
	var err error

	if o.Ipv4PeerPrefix != nil {
		_, ipv4Prefix, err = net.ParseCIDR(*o.Ipv4PeerPrefix)
		if err != nil {
			diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv4 neighbor prefix %q", ipv4Prefix), err.Error())
			return nil
		}
	}

	if o.Ipv6PeerPrefix != nil {
		_, ipv6Prefix, err = net.ParseCIDR(*o.Ipv6PeerPrefix)
		if err != nil {
			diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv6 neighbor prefix %q", ipv6Prefix), err.Error())
			return nil
		}
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface{
		Bfd:                   o.BfdEnabled,
		Holdtime:              o.HoldTime,
		Ipv4Safi:              o.Ipv4AfiEnabled,
		Ipv6Safi:              o.Ipv6AfiEnabled,
		Keepalive:             o.KeepaliveTime,
		LocalAsn:              o.LocalAsn,
		Password:              o.Password,
		PrefixNeighborIpv4:    ipv4Prefix,
		PrefixNeighborIpv6:    ipv6Prefix,
		SessionAddressingIpv4: o.Ipv4Enabled,
		SessionAddressingIpv6: o.Ipv6Enabled,
		Ttl:                   o.Ttl,
	}
}

func (o dynamicBgpPeeringPrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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
