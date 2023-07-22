package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
)

type StaticRoute struct {
	Network         types.String `tfsdk:"network"`
	ShareIpEndpoint types.Bool   `tfsdk:"share_ip_endpoint"`
	Primitive       types.String `tfsdk:"primitive"`
}

func (o StaticRoute) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"network": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 or IPv6 prefix in CIDR notation",
			Required:            true,
			Validators:          []validator.String{apstravalidator.ParseCidr(false, false)},
		},
		"share_ip_endpoint": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the next-hop IP address is shared across " +
				"multiple remote systems. Default:  Default: `false`",
			Optional: true,
		},
		"primitive": dataSourceSchema.StringAttribute{
			MarkdownDescription: "JSON output for use in the `primitives` field of an " +
				"`apstra_datacenter_connectivity_template` resource or a different Connectivity " +
				"Template JsonPrimitive data source",
			Computed: true,
		},
	}
}

func (o StaticRoute) Marshal(_ context.Context, diags *diag.Diagnostics) string {
	obj := staticRoutePrototype{
		Network:         o.Network.ValueString(),
		ShareIpEndpoint: o.ShareIpEndpoint.ValueBool(),
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling StaticRoute primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&TfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachStaticRoute.String(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

var _ JsonPrimitive = &staticRoutePrototype{}

type staticRoutePrototype struct {
	Network         string `json:"network"`
	ShareIpEndpoint bool   `json:"share_ip_endpoint"`
}

func (o staticRoutePrototype) attributes(_ context.Context, path path.Path, diags *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	_, network, err := net.ParseCIDR(o.Network)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing network CIDR string %s", o.Network), err.Error())
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachStaticRoute{
		ShareIpEndpoint: o.ShareIpEndpoint,
		Network:         network,
	}
}

func (o staticRoutePrototype) SdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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
