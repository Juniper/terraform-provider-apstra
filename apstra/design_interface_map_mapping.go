package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type interfaceMapMapping struct {
	DPPort      types.Int64 `tfsdk:"device_profile_port_id"`
	DPTransform types.Int64 `tfsdk:"device_profile_transformation_id"`
	DPInterface types.Int64 `tfsdk:"device_profile_interface_id"`
	LDPanel     types.Int64 `tfsdk:"logical_device_panel"`
	LDPort      types.Int64 `tfsdk:"logical_device_panel_port"`
}

func (o interfaceMapMapping) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"device_profile_port_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Port number(ID) from the Device Profile.",
			Computed:            true,
		},
		"device_profile_transformation_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Port-specific transform ID from the Device Profile.",
			Computed:            true,
		},
		"device_profile_interface_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Port-specific interface ID from the device profile (used to identify interfaces in breakout scenarios.)",
			Computed:            true,
		},
		"logical_device_panel": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Panel number (first panel is 1) of the Logical Device port which corresponds to this interface.",
			Computed:            true,
		},
		"logical_device_panel_port": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Port number (first port is 1) of the Logical Device port which corresponds to this interface.",
			Computed:            true,
		},
	}
}

func (o interfaceMapMapping) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"device_profile_port_id":           types.Int64Type,
		"device_profile_transformation_id": types.Int64Type,
		"device_profile_interface_id":      types.Int64Type,
		"logical_device_panel":             types.Int64Type,
		"logical_device_panel_port":        types.Int64Type,
	}
}

func (o *interfaceMapMapping) loadApiData(_ context.Context, in *goapstra.InterfaceMapMapping, _ *diag.Diagnostics) {
	o.DPPort = types.Int64Value(int64(in.DPPortId))
	o.DPTransform = types.Int64Value(int64(in.DPTransformId))
	o.DPInterface = types.Int64Value(int64(in.DPInterfaceId))
	o.LDPanel = types.Int64Value(int64(in.LDPanel))
	o.LDPort = types.Int64Value(int64(in.LDPort))

	if o.LDPanel.ValueInt64() == -1 {
		o.LDPanel = types.Int64Null()
	}

	if o.LDPort.ValueInt64() == -1 {
		o.LDPort = types.Int64Null()
	}
}

func newInterfaceMapMappingObject(ctx context.Context, in *goapstra.InterfaceMapMapping, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(interfaceMapMapping{}.attrTypes())
	}

	var im interfaceMapMapping
	im.loadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(interfaceMapMapping{}.attrTypes())
	}

	result, d := types.ObjectValueFrom(ctx, interfaceMapMapping{}.attrTypes(), &im)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(interfaceMapMapping{}.attrTypes())
	}

	return result
}
