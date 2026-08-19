package design

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/apstra-go-sdk/speed"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type logicalDevicePanelPortGroup struct {
	PortCount types.Int64  `tfsdk:"port_count"`
	PortSpeed types.String `tfsdk:"port_speed"`
	PortRoles types.Set    `tfsdk:"port_roles"`
}

func (ldppg logicalDevicePanelPortGroup) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port_count": types.Int64Type,
		"port_speed": types.StringType,
		"port_roles": types.SetType{ElemType: types.StringType},
	}
}

func (ldppg logicalDevicePanelPortGroup) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"port_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of ports in the group.",
			Computed:            true,
		},
		"port_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Port speed.",
			Computed:            true,
		},
		"port_roles": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Describes the device types to which this port can connect.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (ldppg logicalDevicePanelPortGroup) resourceAttributes() map[string]resourceSchema.Attribute {
	// collect all port roles for use in inline documentation and defaulter
	var allPortRoles apstra.LogicalDevicePortRoles
	allPortRoles.IncludeAllUses()

	// prepare []attr.Value for defaulter
	defaultRoles := make([]attr.Value, len(allPortRoles.Strings()))
	for i, role := range allPortRoles.Strings() {
		defaultRoles[i] = types.StringValue(role)
	}

	return map[string]resourceSchema.Attribute{
		"port_count": resourceSchema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "Number of ports in the group.",
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"port_speed": resourceSchema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Port speed.",
			Validators: []validator.String{
				apstravalidator.ParseSpeed(),
			},
		},
		"port_roles": resourceSchema.SetAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Optional:    true,
			MarkdownDescription: fmt.Sprintf(
				"One or more of: '%s', by default all values are selected.",
				strings.Join(allPortRoles.Strings(), "', '")),
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.OneOf(allPortRoles.Strings()...)),
			},
			Default: setdefault.StaticValue(types.SetValueMust(types.StringType, defaultRoles)),
		},
	}
}

func (ldppg logicalDevicePanelPortGroup) request(ctx context.Context, diags *diag.Diagnostics) design.LogicalDevicePanelPortGroup {
	result := design.LogicalDevicePanelPortGroup{
		Count: int(ldppg.PortCount.ValueInt64()),
		Speed: speed.Speed(ldppg.PortSpeed.ValueString()),
		Roles: nil, // handled below
	}

	var roles []string
	diags.Append(ldppg.PortRoles.ElementsAs(ctx, &roles, false)...)
	result.Roles.FromStrings(roles)

	return result
}

func (ldppg *logicalDevicePanelPortGroup) loadApiData(ctx context.Context, in design.LogicalDevicePanelPortGroup, diags *diag.Diagnostics) {
	portRoles, d := types.SetValueFrom(ctx, types.StringType, in.Roles.Strings())
	diags.Append(d...)

	ldppg.PortCount = types.Int64Value(int64(in.Count))
	ldppg.PortSpeed = types.StringValue(string(in.Speed))
	ldppg.PortRoles = portRoles
}
