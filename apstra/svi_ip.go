package tfapstra

import (
	"context"
	"fmt"
	"net"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SviIp struct {
	SystemId   types.String `tfsdk:"system_id"`
	IPv4Address types.String `tfsdk:"ipv4_address"`
	IPv4Mode   types.String `tfsdk:"ipv4_mode"`
	IPv6Address types.String `tfsdk:"ipv6_address"`
	IPv6Mode   types.String `tfsdk:"ipv6_mode"`
}

func (o SviIp) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"system_id":    types.StringType,
		"ipv4_address": types.StringType,
		"ipv4_mode":    types.StringType,
		"ipv6_address": types.StringType,
		"ipv6_mode":    types.StringType,
	}
}

func (o SviIp) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"system_id": resourceSchema.StringAttribute{
			MarkdownDescription: "System ID of the switch for this SVI IP assignment",
			Required:            true,
		},
		"ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address with CIDR notation (e.g., '192.0.2.2/24')",
			Optional:            true,
		},
		"ipv4_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "SVI IPv4 mode: 'disabled', 'enabled', or 'forced'",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("enabled"),
			Validators: []validator.String{
				stringvalidator.OneOf("disabled", "enabled", "forced"),
			},
		},
		"ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address with CIDR notation (e.g., '2001:db8::2/64')",
			Optional:            true,
		},
		"ipv6_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "SVI IPv6 mode: 'disabled', 'enabled', 'forced', or 'link_local'",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("disabled"),
			Validators: []validator.String{
				stringvalidator.OneOf("disabled", "enabled", "forced", "link_local"),
			},
		},
	}
}

func (o SviIp) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.SviIp {
	var ipv4Mode enum.SviIpv4Mode
	var ipv6Mode enum.SviIpv6Mode

	err := ipv4Mode.FromString(o.IPv4Mode.ValueString())
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error parsing ipv4_mode %q", o.IPv4Mode.ValueString()),
			err.Error())
		return nil
	}

	err = ipv6Mode.FromString(o.IPv6Mode.ValueString())
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error parsing ipv6_mode %q", o.IPv6Mode.ValueString()),
			err.Error())
		return nil
	}

	result := &apstra.SviIp{
		SystemId: apstra.ObjectId(o.SystemId.ValueString()),
		Ipv4Mode: ipv4Mode,
		Ipv6Mode: ipv6Mode,
	}

	// Parse IPv4 address if provided
	if !o.IPv4Address.IsNull() && o.IPv4Address.ValueString() != "" {
		var ip net.IP
		ip, result.Ipv4Addr, err = net.ParseCIDR(o.IPv4Address.ValueString())
		if err != nil {
			diags.AddError(
				fmt.Sprintf("error parsing ipv4_address %q", o.IPv4Address.ValueString()),
				err.Error())
			return nil
		}
		result.Ipv4Addr.IP = ip
	}

	// Parse IPv6 address if provided
	if !o.IPv6Address.IsNull() && o.IPv6Address.ValueString() != "" {
		var ip net.IP
		ip, result.Ipv6Addr, err = net.ParseCIDR(o.IPv6Address.ValueString())
		if err != nil {
			diags.AddError(
				fmt.Sprintf("error parsing ipv6_address %q", o.IPv6Address.ValueString()),
				err.Error())
			return nil
		}
		result.Ipv6Addr.IP = ip
	}

	return result
}

func (o *SviIp) LoadApiData(ctx context.Context, in apstra.SviIp, diags *diag.Diagnostics) {
	o.SystemId = types.StringValue(string(in.SystemId))
	o.IPv4Mode = types.StringValue(in.Ipv4Mode.String())
	o.IPv6Mode = types.StringValue(in.Ipv6Mode.String())

	if in.Ipv4Addr != nil {
		o.IPv4Address = types.StringValue(in.Ipv4Addr.String())
	} else {
		o.IPv4Address = types.StringNull()
	}

	if in.Ipv6Addr != nil {
		o.IPv6Address = types.StringValue(in.Ipv6Addr.String())
	} else {
		o.IPv6Address = types.StringNull()
	}
}