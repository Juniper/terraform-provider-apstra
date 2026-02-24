package freeform

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
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

type AggregateLinkEndpoint struct {
	SystemID      types.String         `tfsdk:"system_id"`
	IfName        types.String         `tfsdk:"if_name"`
	IPv4Address   cidrtypes.IPv4Prefix `tfsdk:"ipv4_address"`
	IPv6Address   cidrtypes.IPv6Prefix `tfsdk:"ipv6_address"`
	PortChannelID types.Int64          `tfsdk:"port_channel_id"`
	LAGMode       types.String         `tfsdk:"lag_mode"`
	Tags          types.Set            `tfsdk:"tags"`

	ID types.String `tfsdk:"id"` // read only
}

func (o AggregateLinkEndpoint) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"system_id":       types.StringType,
		"if_name":         types.StringType,
		"ipv4_address":    cidrtypes.IPv4PrefixType{},
		"ipv6_address":    cidrtypes.IPv6PrefixType{},
		"port_channel_id": types.Int64Type,
		"lag_mode":        types.StringType,
		"tags":            types.SetType{ElemType: types.StringType},
		"id":              types.StringType,
	}
}

func (o AggregateLinkEndpoint) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of a `system` node.",
			Computed:            true,
		},
		"if_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of the logical aggregate interface e.g. `ae1`, `bond0`.",
			Computed:            true,
		},
		"ipv4_address": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address of the logical aggregate interface, if any.",
			Computed:            true,
		},
		"ipv6_address": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address of the logical aggregate interface, if any.",
			Computed:            true,
		},
		"port_channel_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Numerical Port Channel index of the logical aggregate interface.",
			Computed:            true,
		},
		"lag_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "LAG mode of the logical aggregate interface.",
			Computed:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Set of tags associated with this Aggregate Link Endpoint.",
			Computed:            true,
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the logical aggregate interface associated belonging to both this Aggregate " +
				"Link Endpoint and to `system_id`.",
			Computed: true,
		},
	}
}

func (o AggregateLinkEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"system_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of a `system` node.",
			Required:            true,
		},
		"if_name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the logical aggregate interface e.g. `ae1`, `bond0`.",
			Optional:            true,
			Computed:            true,
		},
		"ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address of the logical aggregate interface, if any.",
			Optional:            true,
		},
		"ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address of the logical aggregate interface, if any.",
			Optional:            true,
		},
		"port_channel_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Numerical Port Channel index of the logical aggregate interface.",
			Required:            true,
		},
		"lag_mode": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("LAG mode of the logical aggregate interface. Must be one of: `%s`.",
				strings.Join(enum.LAGModes.Values(), "`, `")),
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(enum.LAGModes.Values()...)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of tags associated with this Aggregate Link Endpoint.",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the logical aggregate interface associated belonging to both this Aggregate " +
				"Link Endpoint and to `system_id`.",
			Computed: true,
		},
	}
}

func (o AggregateLinkEndpoint) Request(ctx context.Context, path path.Path, diags *diag.Diagnostics) apstra.FreeformAggregateLinkEndpoint {
	result := apstra.FreeformAggregateLinkEndpoint{
		SystemID: o.SystemID.ValueString(),
		IfName:   o.IfName.ValueString(),
		// IPv4Addr:   nil, // see below
		// IPv6Addr:   nil, // see below
		PortChannelID: int(o.PortChannelID.ValueInt64()),
		// Tags:       nil, // see below
		// LAGMode:    enum.LAGMode{}, // see below
	}

	var err error

	if !o.IPv4Address.IsNull() {
		*result.IPv4Addr, err = netip.ParsePrefix(o.IPv4Address.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.AtName("ipv4_address"),
				fmt.Sprintf("Failed to parse IPv4 address %q", o.IPv4Address.ValueString()),
				err.Error(),
			)
			return result
		}
	}

	if !o.IPv6Address.IsNull() {
		*result.IPv6Addr, err = netip.ParsePrefix(o.IPv6Address.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.AtName("ipv6_address"),
				fmt.Sprintf("Failed to parse IPv6 address %q", o.IPv6Address.ValueString()),
				err.Error(),
			)
			return result
		}
	}

	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	err = result.LAGMode.FromString(o.LAGMode.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.AtName("lag_mode"),
			fmt.Sprintf("Failed to parse lag_mode value %q", o.LAGMode.ValueString()),
			err.Error(),
		)
		return result
	}

	return result
}

func (o *AggregateLinkEndpoint) LoadAPIData(ctx context.Context, in apstra.FreeformAggregateLinkEndpoint, diags *diag.Diagnostics) {
	var ipv4Addr, ipv6Addr *string
	if in.IPv4Addr != nil && in.IPv4Addr.IsValid() {
		*ipv4Addr = in.IPv4Addr.String()
	}
	if in.IPv6Addr != nil && in.IPv6Addr.IsValid() {
		*ipv6Addr = in.IPv6Addr.String()
	}

	o.SystemID = types.StringValue(in.SystemID)
	o.IfName = types.StringValue(in.IfName)
	o.IPv4Address = cidrtypes.NewIPv4PrefixPointerValue(ipv4Addr)
	o.IPv6Address = cidrtypes.NewIPv6PrefixPointerValue(ipv6Addr)
	o.PortChannelID = types.Int64Value(int64(in.PortChannelID))
	o.Tags = value.SetOrNull(ctx, types.StringType, in.Tags, diags)
	o.LAGMode = types.StringValue(in.LAGMode.String())
	o.ID = types.StringPointerValue(in.ID())
}
