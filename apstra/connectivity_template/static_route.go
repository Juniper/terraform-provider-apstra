package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
)

var _ Primitive = &StaticRoute{}

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
				"Template Primitive data source",
			Computed: true,
		},
	}
}

func (o StaticRoute) Render(_ context.Context, diags *diag.Diagnostics) string {
	obj := staticRoutePrototype{
		Network:         o.Network.ValueString(),
		ShareIpEndpoint: o.ShareIpEndpoint.ValueBool(),
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling StaticRoute primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&RenderedPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachStaticRoute.String(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o StaticRoute) connectivityTemplateAttributes() (apstra.ConnectivityTemplateAttributes, error) {
	_, network, err := net.ParseCIDR(o.Network.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed parsing network CIDR string %q - %w", o.Network.ValueString(), err)
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachStaticRoute{
		ShareIpEndpoint: o.ShareIpEndpoint.ValueBool(),
		Network:         network,
	}, nil
}

type staticRoutePrototype struct {
	Network         string `json:"network"`
	ShareIpEndpoint bool   `json:"share_ip_endpoint"`
}
