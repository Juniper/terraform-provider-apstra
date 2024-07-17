package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	"strings"
)

type freeformEndpoint struct {
	// SystemId         types.String         `tfsdk:"system_id"`
	InterfaceName    types.String         `tfsdk:"interface_name"`
	InterfaceId      types.String         `tfsdk:"interface_id"`
	TransformationId types.Int64          `tfsdk:"transformation_id"`
	Ipv4Address      cidrtypes.IPv4Prefix `tfsdk:"ipv4_address"`
	Ipv6Address      cidrtypes.IPv6Prefix `tfsdk:"ipv6_address"`
}

func (o freeformEndpoint) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		// "system_id":         types.StringType,
		"interface_name":    types.StringType,
		"interface_id":      types.StringType,
		"transformation_id": types.Int64Type,
		"ipv4_address":      cidrtypes.IPv4PrefixType{},
		"ipv6_address":      cidrtypes.IPv6PrefixType{},
	}
}

func (o freeformEndpoint) DatasourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"interface_name": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "the interface name",
		},
		"interface_id": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Graph node ID of the attached interface for this side of the link endpoint ",
		},
		"transformation_id": dataSourceSchema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "ID of the transformation in the device profile",
		},
		"ipv4_address": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Ipv4 address of the interface",
			CustomType:          cidrtypes.IPv4PrefixType{},
		},
		"ipv6_address": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Ipv6 address of the interface",
			CustomType:          cidrtypes.IPv6PrefixType{},
		},
	}
}

func (o freeformEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"interface_name": resourceSchema.StringAttribute{
			Required:            true,
			MarkdownDescription: "the interface name",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"interface_id": resourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Graph node ID of the attached interface for this side of the link endpoint.",
		},
		"transformation_id": resourceSchema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "ID of the transformation in the device profile",
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"ipv4_address": resourceSchema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Ipv4 address of the interface",
			CustomType:          cidrtypes.IPv4PrefixType{},
		},
		"ipv6_address": resourceSchema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Ipv6 address of the interface",
			CustomType:          cidrtypes.IPv6PrefixType{},
		},
	}
}

func (o *freeformEndpoint) request(systemId string) *apstra.FreeformEndpoint {
	var ipNet4, ipNet6 *net.IPNet
	if !o.Ipv4Address.IsNull() {
		var ip4 net.IP
		ip4, ipNet4, _ = net.ParseCIDR(o.Ipv4Address.ValueString())
		ipNet4.IP = ip4
	}
	if !o.Ipv6Address.IsNull() {
		var ip6 net.IP
		ip6, ipNet6, _ = net.ParseCIDR(o.Ipv6Address.ValueString())
		ipNet6.IP = ip6
	}

	return &apstra.FreeformEndpoint{
		SystemId: apstra.ObjectId(systemId),
		Interface: apstra.FreeformInterface{
			Data: &apstra.FreeformInterfaceData{
				IfName:           o.InterfaceName.ValueString(),
				TransformationId: int(o.TransformationId.ValueInt64()),
				Ipv4Address:      ipNet4,
				Ipv6Address:      ipNet6,
			},
		},
	}
}

func (o *freeformEndpoint) loadApiData(_ context.Context, in apstra.FreeformEndpoint, diags *diag.Diagnostics) {
	if in.Interface.Id == nil {
		diags.AddError(
			fmt.Sprintf("api returned nil interface Id for system %s", in.SystemId),
			"interface IDs should always be populated",
		)
		return
	}

	o.InterfaceName = types.StringValue(in.Interface.Data.IfName)
	o.InterfaceId = types.StringValue(in.Interface.Id.String())
	//o.SystemId = types.StringValue(in.SystemId.String())
	o.TransformationId = types.Int64Value(int64(in.Interface.Data.TransformationId))
	o.Ipv4Address = cidrtypes.NewIPv4PrefixValue(in.Interface.Data.Ipv4Address.String())
	if strings.Contains(o.Ipv4Address.ValueString(), "nil") {
		o.Ipv4Address = cidrtypes.NewIPv4PrefixNull()
	}
	o.Ipv6Address = cidrtypes.NewIPv6PrefixValue(in.Interface.Data.Ipv6Address.String())
	if strings.Contains(o.Ipv6Address.ValueString(), "nil") {
		o.Ipv6Address = cidrtypes.NewIPv6PrefixNull()
	}
}

func newFreeformEndpointMap(ctx context.Context, in [2]apstra.FreeformEndpoint, diags *diag.Diagnostics) types.Map {
	endpoints := make(map[string]freeformEndpoint, len(in))
	for i := range in {
		var endpoint freeformEndpoint
		endpoint.loadApiData(ctx, in[i], diags)
		endpoints[in[i].SystemId.String()] = endpoint
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: freeformEndpoint{}.attrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: freeformEndpoint{}.attrTypes()}, endpoints, diags)
}
