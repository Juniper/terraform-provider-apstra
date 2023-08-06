package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
)

var _ Primitive = &CustomStaticRoute{}

type CustomStaticRoute struct {
	RoutingZoneId types.String `tfsdk:"routing_zone_id"`
	Network       types.String `tfsdk:"network"`
	NextHop       types.String `tfsdk:"next_hop"`
	Primitive     types.String `tfsdk:"primitive"`
}

func (o CustomStaticRoute) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"routing_zone_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of Routing Zone",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"network": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 or IPv6 prefix in CIDR notation",
			Required:            true,
			Validators:          []validator.String{apstravalidator.ParseCidr(false, false)},
		},
		"next_hop": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 or IPv6 address of next hop router",
			Required:            true,
			Validators:          []validator.String{apstravalidator.ParseIp(false, false)},
		},
		"primitive": dataSourceSchema.StringAttribute{
			MarkdownDescription: "JSON output for use in the `primitives` field of an " +
				"`apstra_datacenter_connectivity_template` resource or a different Connectivity " +
				"Template JsonPrimitive data source",
			Computed: true,
		},
	}
}

func (o CustomStaticRoute) Marshal(_ context.Context, diags *diag.Diagnostics) string {
	obj := customStaticRoutePrototype{}

	if !o.RoutingZoneId.IsNull() {
		rzId := o.RoutingZoneId.ValueString()
		obj.RoutingZoneId = &rzId
	}

	if !o.Network.IsNull() {
		network := o.Network.ValueString()
		obj.Network = &network
	}

	if !o.NextHop.IsNull() {
		nextHop := o.NextHop.ValueString()
		obj.NextHop = &nextHop
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling CustomStaticRoute primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachCustomStaticRoute.String(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *CustomStaticRoute) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	o.loadSdkPrimitiveAttributes(ctx, attributes, diags)
	if diags.HasError() {
		return
	}
}

func (o *CustomStaticRoute) loadSdkPrimitiveAttributes(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute, _ *diag.Diagnostics) {
	o.RoutingZoneId = types.StringNull()
	if in.SecurityZone != nil {
		o.RoutingZoneId = types.StringValue(in.SecurityZone.String())
	}

	o.Network = types.StringNull()
	if in.Network != nil {
		o.Network = types.StringValue(in.Network.String())
	}

	o.NextHop = types.StringNull()
	if in.NextHop != nil {
		o.NextHop = types.StringValue(in.NextHop.String())
	}
}

var _ JsonPrimitive = &customStaticRoutePrototype{}

type customStaticRoutePrototype struct {
	RoutingZoneId *string `json:"routing_zone_id"`
	Network       *string `json:"network"`
	NextHop       *string `json:"next_hop_ip_address"`
}

func (o customStaticRoutePrototype) attributes(_ context.Context, path path.Path, diags *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	var result apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute
	var err error

	if o.Network != nil {
		_, result.Network, err = net.ParseCIDR(*o.Network)
		if err != nil {
			diags.AddAttributeError(path, fmt.Sprintf("failed parsing network CIDR string %q", o.Network), err.Error())
			return nil
		}
	}

	if o.NextHop != nil {
		result.NextHop = net.ParseIP(*o.NextHop)
		if result.NextHop == nil {
			diags.AddAttributeError(path, fmt.Sprintf("failed parsing next hop IP address string %q", o.Network), err.Error())
		}
	}

	if o.RoutingZoneId != nil {
		id := apstra.ObjectId(*o.RoutingZoneId)
		result.SecurityZone = &id
	}

	return &result
}

func (o customStaticRoutePrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	attributes := o.attributes(ctx, path, diags)
	if diags.HasError() {
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitive{
		Id:          nil, // calculated later
		Attributes:  attributes,
		Subpolicies: nil, // this primitive has no children
		BatchId:     nil, // this primitive has no children
		PipelineId:  nil, // calculated later
	}
}
