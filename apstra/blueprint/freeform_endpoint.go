package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

type freeformEndpoint struct {
	SystemId         types.String         `tfsdk:"system_id"`
	InterfaceName    types.String         `tfsdk:"interface_name"`
	TransformationId types.Int64          `tfsdk:"transformation_id"`
	Ipv4Address      cidrtypes.IPv4Prefix `tfsdk:"ipv4_address"`
	Ipv6Address      cidrtypes.IPv6Prefix `tfsdk:"ipv6_address"`
}

func (o freeformEndpoint) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"system_id":         types.StringType,
		"interface_name":    types.StringType,
		"transformation_id": types.Int64Type,
		"ipv4_address":      cidrtypes.IPv4PrefixType{},
		"ipv6_address":      cidrtypes.IPv6PrefixType{},
	}
}

func (o freeformEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"system_id": resourceSchema.StringAttribute{
			Required:            true,
			MarkdownDescription: "System ID ",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"interface_name": resourceSchema.StringAttribute{
			Required:            true,
			MarkdownDescription: "fill this out",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"transformation_id": resourceSchema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "fill this out",
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

func (o *freeformEndpoint) request() *apstra.FreeformEndpoint {
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
		SystemId: apstra.ObjectId(o.SystemId.ValueString()),
		Interface: apstra.FreeformInterfaceData{
			IfName:           o.InterfaceName.ValueString(),
			TransformationId: int(o.TransformationId.ValueInt64()),
			Ipv4Address:      ipNet4,
			Ipv6Address:      ipNet6,
		},
	}
}

func (o *freeformEndpoint) loadApiData(_ context.Context, in apstra.FreeformEndpoint, _ *diag.Diagnostics) {
	o.InterfaceName = types.StringValue(in.Interface.IfName)
	o.SystemId = types.StringValue(in.SystemId.String())
}

func newFreeformEndpointSet(ctx context.Context, in [2]apstra.FreeformEndpoint, diags *diag.Diagnostics) types.Set {
	endpoints := make([]freeformEndpoint, len(in))
	for i, endpoint := range in {
		endpoints[i].loadApiData(ctx, endpoint, diags)
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: freeformEndpoint{}.attrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: freeformEndpoint{}.attrTypes()}, endpoints, diags)
}
