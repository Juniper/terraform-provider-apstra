package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-apstra/apstra/design"
)

var _ Primitive = &IpLink{}

type IpLink struct {
	RoutingZoneId      types.String `tfsdk:"routing_zone_id"`
	VlanId             types.Int64  `tfsdk:"vlan_id"`
	Ipv4AddressingType types.String `tfsdk:"ipv4_addressing_type"`
	Ipv6AddressingType types.String `tfsdk:"ipv6_addressing_type"`
	Primitive          types.String `tfsdk:"primitive"`
	Children           types.List   `tfsdk:"children"`
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
		"children": dataSourceSchema.ListAttribute{
			MarkdownDescription: "A list of JSON strings describing Connectivity Template Primitives " +
				"which are children of this Connectivity Template JsonPrimitive. Use the `primitive` " +
				"attribute of other Connectivity Template Primitives data sources here.",
			ElementType: types.StringType,
			Optional:    true,
		},
	}
}

func (o IpLink) Marshal(ctx context.Context, diags *diag.Diagnostics) string {
	var children []string
	diags.Append(o.Children.ElementsAs(ctx, &children, false)...)
	if diags.HasError() {
		return ""
	}

	obj := ipLinkPrototype{
		RoutingZoneId:      o.RoutingZoneId.ValueString(),
		Tagged:             false,
		VlanId:             nil,
		Ipv4AddressingType: o.Ipv4AddressingType.ValueString(),
		Ipv6AddressingType: o.Ipv6AddressingType.ValueString(),
		Children:           children,
	}

	if !o.VlanId.IsNull() {
		obj.Tagged = true
		vlan := o.VlanId.ValueInt64()
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
	switch attributes := in.Attributes.(type) {
	case *apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink:
		o.loadSdkPrimitiveAttributes(ctx, attributes, diags)
		if diags.HasError() {
			return
		}
	default:
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %t", in))
		return
	}

	o.loadSdkPrimitiveChildren(ctx, in.Subpolicies, diags)
	if diags.HasError() {
		return
	}
}

func (o *IpLink) loadSdkPrimitiveAttributes(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink, _ *diag.Diagnostics) {
	routingZone := types.StringNull()
	if in.SecurityZone != nil {
		routingZone = types.StringValue(in.SecurityZone.String())
	}

	o.RoutingZoneId = routingZone
	o.VlanId = types.Int64Value(int64(*in.Vlan))
	o.Ipv4AddressingType = types.StringValue(in.IPv4AddressingType.String())
	o.Ipv6AddressingType = types.StringValue(in.IPv6AddressingType.String())
}

// loadSdkPrimitiveChildren imports the child policies into o.Children as JSON strings
func (o *IpLink) loadSdkPrimitiveChildren(ctx context.Context, in []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	children := SdkPrimitivesToJsonStrings(ctx, in, diags)
	o.Children = types.ListValueMust(types.StringType, children)
}

var _ JsonPrimitive = &ipLinkPrototype{}

type ipLinkPrototype struct {
	RoutingZoneId      string   `json:"routing_zone_id"`
	Tagged             bool     `json:"tagged"`
	VlanId             *int64   `json:"vlan_id,omitempty"`
	Ipv4AddressingType string   `json:"ipv4_addressing_type"`
	Ipv6AddressingType string   `json:"ipv6_addressing_type"`
	Children           []string `json:"children,omitempty"`
}

func (o ipLinkPrototype) attributes(_ context.Context, path path.Path, diags *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	routingZoneId := apstra.ObjectId(o.RoutingZoneId)

	vlanId := apstra.Vlan(*o.VlanId)

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

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
		SecurityZone:       &routingZoneId,
		Tagged:             o.Tagged,
		Vlan:               &vlanId,
		IPv4AddressingType: ipv4AddressingType,
		IPv6AddressingType: ipv6AddressingType,
	}
}

func (o ipLinkPrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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
