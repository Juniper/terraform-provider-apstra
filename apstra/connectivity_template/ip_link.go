package connectivitytemplate

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"sort"
	"strings"
)

var _ Primitive = &IpLink{}

type IpLink struct {
	Name               types.String `tfsdk:"name"`
	RoutingZoneId      types.String `tfsdk:"routing_zone_id"`
	VlanId             types.Int64  `tfsdk:"vlan_id"`
	Ipv4AddressingType types.String `tfsdk:"ipv4_addressing_type"`
	Ipv6AddressingType types.String `tfsdk:"ipv6_addressing_type"`
	Primitive          types.String `tfsdk:"primitive"`
	ChildPrimitives    types.Set    `tfsdk:"child_primitives"`
	L3Mtu              types.Int64  `tfsdk:"l3_mtu"`
}

func (o IpLink) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	ipv4AddressingTypes := []string{
		apstra.CtPrimitiveIPv4AddressingTypeNumbered.String(),
		apstra.CtPrimitiveIPv4AddressingTypeNone.String(),
	}
	ipv6AddressingTypes := []string{
		apstra.CtPrimitiveIPv6AddressingTypeLinkLocal.String(),
		apstra.CtPrimitiveIPv6AddressingTypeNumbered.String(),
		apstra.CtPrimitiveIPv6AddressingTypeNone.String(),
	}
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Primitive name displayed in the web UI",
			Optional:            true,
		},
		"routing_zone_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Routing Zone to which this IP Link belongs",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "When set, selects the 802.1Q VLAN ID to use for the link's traffic. " +
				"Omit for an untagged link.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(design.VlanMin, design.VlanMax)},
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
		"primitive": dataSourceSchema.StringAttribute{
			MarkdownDescription: "JSON output for use in the `primitives` field of an " +
				"`apstra_datacenter_connectivity_template` resource or a different Connectivity " +
				"Template JsonPrimitive data source",
			Computed: true,
		},
		"child_primitives": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of JSON strings describing Connectivity Template Primitives " +
				"which are children of this Connectivity Template Primitive. Use the `primitive` " +
				"attribute of other Connectivity Template Primitives data sources here.",
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
			Optional:    true,
		},
		"l3_mtu": dataSourceSchema.Int64Attribute{
			// Frankly, I'm not clear what this text is trying to say. It's
			// taken verbatim from the tooltip in 99.2.0-cl-4.2.0-1
			MarkdownDescription: fmt.Sprintf("L3 MTU for sub-interfaces on leaf (spine/superspine) side and "+
				"generic side. Configuration is applicable only when Fabric MTU is enabled. Value must be even "+
				"number rom %d to %d, if not specified - Default IP Links to Generic Systems MTU from Virtual "+
				"Network Policy s used", constants.L3MtuMin, constants.L3MtuMax),
			Optional: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
	}
}

func (o IpLink) Marshal(ctx context.Context, diags *diag.Diagnostics) string {
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

	obj := ipLinkPrototype{
		RoutingZoneId:      o.RoutingZoneId.ValueString(),
		Tagged:             false,
		VlanId:             nil,
		Ipv4AddressingType: o.Ipv4AddressingType.ValueString(),
		Ipv6AddressingType: o.Ipv6AddressingType.ValueString(),
		ChildPrimitives:    childPrimitives,
	}

	if !o.VlanId.IsNull() {
		obj.Tagged = true
		vlan := apstra.Vlan(o.VlanId.ValueInt64())
		obj.VlanId = &vlan
	}

	if o.Ipv4AddressingType.IsNull() {
		obj.Ipv4AddressingType = apstra.CtPrimitiveIPv4AddressingTypeNone.String()
	}

	if o.Ipv6AddressingType.IsNull() {
		obj.Ipv6AddressingType = apstra.CtPrimitiveIPv6AddressingTypeNone.String()
	}

	if !o.L3Mtu.IsNull() {
		l3Mtu := uint16(o.L3Mtu.ValueInt64())
		obj.L3Mtu = &l3Mtu
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling IpLink primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachLogicalLink.String(),
		Label:         o.Name.ValueString(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *IpLink) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	if attributes.SecurityZone != nil {
		o.RoutingZoneId = types.StringValue(attributes.SecurityZone.String())
	} else {
		o.RoutingZoneId = types.StringNull()
	}
	if attributes.Vlan != nil {
		o.VlanId = types.Int64Value(int64(*attributes.Vlan))
	} else {
		o.VlanId = types.Int64Null()
	}
	o.Ipv4AddressingType = types.StringValue(attributes.IPv4AddressingType.String())
	o.Ipv6AddressingType = types.StringValue(attributes.IPv6AddressingType.String())
	o.ChildPrimitives = utils.SetValueOrNull(ctx, types.StringType, SdkPrimitivesToJsonStrings(ctx, in.Subpolicies, diags), diags)
	o.Name = types.StringValue(in.Label)
	o.L3Mtu = utils.Int64ValueOrNull(ctx, attributes.L3Mtu, diags)
}

func (o IpLink) VersionConstraints() apiversions.Constraints {
	var response apiversions.Constraints

	if !o.L3Mtu.IsNull() {
		response.AddAttributeConstraints(
			apiversions.AttributeConstraint{
				Path:        path.Root("l3_mtu"),
				Constraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra420)),
			},
		)
	}

	return response
}

var _ JsonPrimitive = &ipLinkPrototype{}

type ipLinkPrototype struct {
	Label              string       `json:"label,omitempty"`
	RoutingZoneId      string       `json:"routing_zone_id"`
	Tagged             bool         `json:"tagged"`
	VlanId             *apstra.Vlan `json:"vlan_id,omitempty"`
	Ipv4AddressingType string       `json:"ipv4_addressing_type"`
	Ipv6AddressingType string       `json:"ipv6_addressing_type"`
	ChildPrimitives    []string     `json:"child_primitives,omitempty"`
	L3Mtu              *uint16      `json:"l3_mtu,omitempty"`
}

func (o ipLinkPrototype) attributes(_ context.Context, path path.Path, diags *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	routingZoneId := apstra.ObjectId(o.RoutingZoneId)

	var err error
	var ipv4AddressingType apstra.CtPrimitiveIPv4AddressingType
	err = ipv4AddressingType.FromString(o.Ipv4AddressingType)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv4 addressing type %q", o.Ipv4AddressingType), err.Error())
		return nil
	}

	var ipv6AddressingType apstra.CtPrimitiveIPv6AddressingType
	err = ipv6AddressingType.FromString(o.Ipv6AddressingType)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing ipv6 addressing type %q", o.Ipv6AddressingType), err.Error())
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
		SecurityZone:       &routingZoneId,
		Tagged:             o.Tagged,
		Vlan:               o.VlanId,
		IPv4AddressingType: ipv4AddressingType,
		IPv6AddressingType: ipv6AddressingType,
		L3Mtu:              o.L3Mtu,
	}
}

func (o ipLinkPrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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
