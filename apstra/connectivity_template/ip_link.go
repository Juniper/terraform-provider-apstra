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
	"sort"
	"strings"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/utils"
)

var _ Primitive = &IpLink{}

type IpLink struct {
	Label              types.String `tfsdk:"label"`
	RoutingZoneId      types.String `tfsdk:"routing_zone_id"`
	VlanId             types.Int64  `tfsdk:"vlan_id"`
	Ipv4AddressingType types.String `tfsdk:"ipv4_addressing_type"`
	Ipv6AddressingType types.String `tfsdk:"ipv6_addressing_type"`
	Primitive          types.String `tfsdk:"primitive"`
	ChildPrimitives    types.Set    `tfsdk:"child_primitives"`
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
		"label": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Primitive label displayed in the web UI",
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
			Validators: []validator.Int64{int64validator.Between(design.VlanMin-1, design.VlanMax+1)},
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

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling IpLink primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachLogicalLink.String(),
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
}

var _ JsonPrimitive = &ipLinkPrototype{}

type ipLinkPrototype struct {
	RoutingZoneId      string       `json:"routing_zone_id"`
	Tagged             bool         `json:"tagged"`
	VlanId             *apstra.Vlan `json:"vlan_id,omitempty"`
	Ipv4AddressingType string       `json:"ipv4_addressing_type"`
	Ipv6AddressingType string       `json:"ipv6_addressing_type"`
	ChildPrimitives    []string     `json:"child_primitives,omitempty"`
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
		Attributes:  attributes,
		Subpolicies: childPrimitives,
		BatchId:     nil, // calculated later
		PipelineId:  nil, // calculated later
	}
}
