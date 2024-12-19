package freeform

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LinkEndpoint struct {
	InterfaceName    types.String         `tfsdk:"interface_name"`
	InterfaceId      types.String         `tfsdk:"interface_id"`
	TransformationId types.Int64          `tfsdk:"transformation_id"`
	Ipv4Address      cidrtypes.IPv4Prefix `tfsdk:"ipv4_address"`
	Ipv6Address      cidrtypes.IPv6Prefix `tfsdk:"ipv6_address"`
	Tags             types.Set            `tfsdk:"tags"`
}

func (o LinkEndpoint) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interface_name":    types.StringType,
		"interface_id":      types.StringType,
		"transformation_id": types.Int64Type,
		"ipv4_address":      cidrtypes.IPv4PrefixType{},
		"ipv6_address":      cidrtypes.IPv6PrefixType{},
		"tags":              types.SetType{ElemType: types.StringType},
	}
}

func (o LinkEndpoint) DatasourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"interface_name": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The interface name, as found in the associated Device Profile, e.g. `xe-0/0/0`",
		},
		"interface_id": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Graph node ID of the associated interface",
		},
		"transformation_id": dataSourceSchema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "ID # of the transformation in the Device Profile",
		},
		"ipv4_address": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Ipv4 address of the interface in CIDR notation",
			CustomType:          cidrtypes.IPv4PrefixType{},
		},
		"ipv6_address": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Ipv6 address of the interface in CIDR notation",
			CustomType:          cidrtypes.IPv6PrefixType{},
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags applied to the interface",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o LinkEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"interface_name": resourceSchema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The interface name, as found in the associated Device Profile, e.g. `xe-0/0/0`",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("transformation_id").Resolve()),
			},
		},
		"interface_id": resourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Graph node ID of the associated interface",
		},
		"transformation_id": resourceSchema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "ID # of the transformation in the Device Profile",
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("interface_name").Resolve()),
			},
		},
		"ipv4_address": resourceSchema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Ipv4 address of the interface in CIDR notation",
			CustomType:          cidrtypes.IPv4PrefixType{},
		},
		"ipv6_address": resourceSchema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Ipv6 address of the interface in CIDR notation",
			CustomType:          cidrtypes.IPv6PrefixType{},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags applied to the interface",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *LinkEndpoint) request(ctx context.Context, systemId string, diags *diag.Diagnostics) *apstra.FreeformEthernetEndpoint {
	var ipNet4, ipNet6 *net.IPNet
	if utils.HasValue(o.Ipv4Address) {
		var ip4 net.IP
		ip4, ipNet4, _ = net.ParseCIDR(o.Ipv4Address.ValueString())
		ipNet4.IP = ip4
	}
	if utils.HasValue(o.Ipv6Address) {
		var ip6 net.IP
		ip6, ipNet6, _ = net.ParseCIDR(o.Ipv6Address.ValueString())
		ipNet6.IP = ip6
	}

	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)

	var transformationId *int
	if !o.TransformationId.IsNull() {
		transformationId = utils.ToPtr(int(o.TransformationId.ValueInt64()))
	}

	return &apstra.FreeformEthernetEndpoint{
		SystemId: apstra.ObjectId(systemId),
		Interface: apstra.FreeformInterface{
			Data: &apstra.FreeformInterfaceData{
				IfName:           o.InterfaceName.ValueStringPointer(),
				TransformationId: transformationId,
				Ipv4Address:      ipNet4,
				Ipv6Address:      ipNet6,
				Tags:             tags,
			},
		},
	}
}

func (o *LinkEndpoint) loadApiData(ctx context.Context, in apstra.FreeformEthernetEndpoint, diags *diag.Diagnostics) {
	if in.Interface.Id == nil {
		diags.AddError(
			fmt.Sprintf("api returned nil interface Id for system %s", in.SystemId),
			"interface IDs should always be populated",
		)
		return
	}

	var transformationId *int64
	if in.Interface.Data.TransformationId != nil {
		transformationId = utils.ToPtr(int64(*in.Interface.Data.TransformationId))
	}

	o.InterfaceName = types.StringPointerValue(in.Interface.Data.IfName)
	o.InterfaceId = types.StringValue(in.Interface.Id.String())
	o.TransformationId = types.Int64PointerValue(transformationId)
	o.Ipv4Address = cidrtypes.NewIPv4PrefixValue(in.Interface.Data.Ipv4Address.String())
	if strings.Contains(o.Ipv4Address.ValueString(), "nil") {
		o.Ipv4Address = cidrtypes.NewIPv4PrefixNull()
	}
	o.Ipv6Address = cidrtypes.NewIPv6PrefixValue(in.Interface.Data.Ipv6Address.String())
	if strings.Contains(o.Ipv6Address.ValueString(), "nil") {
		o.Ipv6Address = cidrtypes.NewIPv6PrefixNull()
	}
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Interface.Data.Tags, diags)
}

func newFreeformEndpointMap(ctx context.Context, in [2]apstra.FreeformEthernetEndpoint, diags *diag.Diagnostics) types.Map {
	endpoints := make(map[string]LinkEndpoint, len(in))
	for i := range in {
		var endpoint LinkEndpoint
		endpoint.loadApiData(ctx, in[i], diags)
		endpoints[in[i].SystemId.String()] = endpoint
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: LinkEndpoint{}.attrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: LinkEndpoint{}.attrTypes()}, endpoints, diags)
}
