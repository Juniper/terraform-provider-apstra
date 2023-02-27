package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type interfaceMapInterface struct {
	Name     types.String `tfsdk:"name"`
	Roles    types.Set    `tfsdk:"roles"`
	Mapping  types.Object `tfsdk:"mapping"`
	Active   types.Bool   `tfsdk:"active"`
	Position types.Int64  `tfsdk:"position"`
	Speed    types.String `tfsdk:"speed"`
	Setting  types.String `tfsdk:"setting"`
}

func (o interfaceMapInterface) dataSourceSchema() map[string]dataSourceSchema.Attribute {
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
			Attributes:          interfaceMapMapping{}.dataSourceAttributes(),
		},
		"setting": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Vendor specific commands needed to configure the interface, from the device profile.",
			Computed:            true,
		},
	}
}

func (o interfaceMapInterface) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.StringType,
		"roles":    types.SetType{ElemType: types.StringType},
		"mapping":  types.ObjectType{AttrTypes: interfaceMapMapping{}.attrTypes()},
		"active":   types.BoolType,
		"position": types.Int64Type,
		"speed":    types.StringType,
		"setting":  types.StringType,
	}
}

func (o *interfaceMapInterface) loadApiData(ctx context.Context, in *goapstra.InterfaceMapInterface, diags *diag.Diagnostics) {
	var mapping interfaceMapMapping
	mapping.loadApiData(ctx, &in.Mapping, diags)

	o.Name = types.StringValue(in.Name)
	o.Roles = setValueOrNull(ctx, types.StringType, in.Roles.Strings(), diags)
	o.Mapping = newInterfaceMapMappingObject(ctx, &in.Mapping, diags)
	o.Active = types.BoolValue(bool(in.ActiveState))
	o.Position = types.Int64Value(int64(in.Position))
	o.Speed = types.StringValue(string(in.Speed))
	o.Setting = types.StringValue(in.Setting.Param)
}

func newInterfaceMapInterfaceSet(ctx context.Context, in []goapstra.InterfaceMapInterface, diags *diag.Diagnostics) types.Set {
	interfaces := make([]interfaceMapInterface, len(in))
	for i := range in {
		interfaces[i].loadApiData(ctx, &in[i], diags)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: interfaceMapInterface{}.attrTypes()})
		}
	}
	return setValueOrNull(ctx, types.ObjectType{AttrTypes: interfaceMapInterface{}.attrTypes()}, interfaces, diags)
}
