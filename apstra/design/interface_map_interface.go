package design

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InterfaceMapInterface struct {
	Name     types.String `tfsdk:"name"`
	Roles    types.Set    `tfsdk:"roles"`
	Mapping  types.Object `tfsdk:"mapping"`
	Active   types.Bool   `tfsdk:"active"`
	Position types.Int64  `tfsdk:"position"`
	Speed    types.String `tfsdk:"speed"`
	Setting  types.String `tfsdk:"setting"`
}

func (o InterfaceMapInterface) DataSourceSchema() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Physical device interface name",
			Computed:            true,
		},
		"roles": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Logical Device role (\"connected to\") of the interface.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"position": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "todo - need to find out what this is", // todo
			Computed:            true,
		},
		"active": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the interface is used by the Interface Map",
			Computed:            true,
		},
		"speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Interface speed",
			Computed:            true,
		},
		"mapping": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Mapping info for each physical interface",
			Computed:            true,
			Attributes:          InterfaceMapMapping{}.DataSourceAttributes(),
		},
		"setting": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Vendor specific commands needed to configure the interface, from the device profile.",
			Computed:            true,
		},
	}
}

func (o InterfaceMapInterface) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.StringType,
		"roles":    types.SetType{ElemType: types.StringType},
		"mapping":  types.ObjectType{AttrTypes: InterfaceMapMapping{}.AttrTypes()},
		"active":   types.BoolType,
		"position": types.Int64Type,
		"speed":    types.StringType,
		"setting":  types.StringType,
	}
}

func (o *InterfaceMapInterface) LoadApiData(ctx context.Context, in *apstra.InterfaceMapInterface, diags *diag.Diagnostics) {
	var mapping InterfaceMapMapping
	mapping.LoadApiData(ctx, &in.Mapping, diags)

	o.Name = types.StringValue(in.Name)
	o.Roles = value.SetOrNull(ctx, types.StringType, in.Roles.Strings(), diags)
	o.Mapping = NewInterfaceMapMappingObject(ctx, &in.Mapping, diags)
	o.Active = types.BoolValue(bool(in.ActiveState))
	o.Position = types.Int64Value(int64(in.Position))
	o.Speed = types.StringValue(string(in.Speed))
	o.Setting = types.StringValue(in.Setting.Param)
}

func NewInterfaceMapInterfaceSet(ctx context.Context, in []apstra.InterfaceMapInterface, diags *diag.Diagnostics) types.Set {
	interfaces := make([]InterfaceMapInterface, len(in))
	for i := range in {
		interfaces[i].LoadApiData(ctx, &in[i], diags)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: InterfaceMapInterface{}.AttrTypes()})
		}
	}
	return value.SetOrNull(ctx, types.ObjectType{AttrTypes: InterfaceMapInterface{}.AttrTypes()}, interfaces, diags)
}
