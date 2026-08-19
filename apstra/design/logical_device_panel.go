package design

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type logicalDevicePanel struct {
	Rows       types.Int64 `tfsdk:"rows"`
	Columns    types.Int64 `tfsdk:"columns"`
	PortGroups types.List  `tfsdk:"port_groups"`
}

func (o logicalDevicePanel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":        types.Int64Type,
		"columns":     types.Int64Type,
		"port_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: logicalDevicePanelPortGroup{}.attrTypes()}},
	}
}

func (o logicalDevicePanel) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"rows": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Physical vertical dimension of the panel.",
			Computed:            true,
		},
		"columns": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Physical horizontal dimension of the panel.",
			Computed:            true,
		},
		"port_groups": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel.",
			Computed:            true,
			NestedObject:        dataSourceSchema.NestedAttributeObject{Attributes: logicalDevicePanelPortGroup{}.dataSourceAttributes()},
		},
	}
}

func (o logicalDevicePanel) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"rows": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical vertical dimension of the panel.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"columns": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical horizontal dimension of the panel.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"port_groups": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Required:            true,
			Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			NestedObject:        resourceSchema.NestedAttributeObject{Attributes: logicalDevicePanelPortGroup{}.resourceAttributes()},
		},
	}
}

func (o logicalDevicePanel) resourceAttributesDefinition() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"rows": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical vertical dimension of the panel.",
			Computed:            true,
		},
		"columns": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical horizontal dimension of the panel.",
			Computed:            true,
		},
		"port_groups": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Computed:            true,
			NestedObject:        resourceSchema.NestedAttributeObject{Attributes: logicalDevicePanelPortGroup{}.resourceAttributesDefinition()},
		},
	}
}

func (o *logicalDevicePanel) loadApiData(ctx context.Context, in design.LogicalDevicePanel, diags *diag.Diagnostics) {
	portGroups := make([]logicalDevicePanelPortGroup, len(in.PortGroups))
	for i := range in.PortGroups {
		portGroups[i].loadApiData(ctx, in.PortGroups[i], diags)
		if diags.HasError() {
			return
		}
	}

	portGroupList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: logicalDevicePanelPortGroup{}.attrTypes()}, portGroups)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.Rows = types.Int64Value(int64(in.PanelLayout.RowCount))
	o.Columns = types.Int64Value(int64(in.PanelLayout.ColumnCount))
	o.PortGroups = portGroupList
}

func (o *logicalDevicePanel) request(ctx context.Context, diags *diag.Diagnostics) design.LogicalDevicePanel {
	result := design.LogicalDevicePanel{
		PanelLayout:  design.LogicalDevicePanelLayout{RowCount: int(o.Rows.ValueInt64()), ColumnCount: int(o.Columns.ValueInt64())},
		PortGroups:   make([]design.LogicalDevicePanelPortGroup, o.PortGroups.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true, UnhandledUnknownAsZero: true})), // handled below
		PortIndexing: enum.DesignLogicalDevicePanelPortIndexingLRTB,                                                                                                               // always this configuration
	}

	for i, v := range o.PortGroups.Elements() {
		var portGroup logicalDevicePanelPortGroup
		diags.Append(v.(types.Object).As(ctx, &portGroup, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return result
		}
		result.PortGroups[i] = portGroup.request(ctx, diags)
	}

	return result
}

func (o *logicalDevicePanel) validate(ctx context.Context, path path.Path, diags *diag.Diagnostics) {
	if o.Rows.IsUnknown() || o.Columns.IsUnknown() || o.PortGroups.IsUnknown() {
		return
	}

	var portGroups []logicalDevicePanelPortGroup
	diags.Append(o.PortGroups.ElementsAs(ctx, &portGroups, false)...)
	if diags.HasError() {
		return
	}

	// count up the ports in each port group
	panelPortsByPortGroup := int64(0)
	for _, portGroup := range portGroups {
		if portGroup.PortCount.IsUnknown() {
			return // cannot validate with any unknown port count
		}

		panelPortsByPortGroup = panelPortsByPortGroup + portGroup.PortCount.ValueInt64()
	}

	// use panel geometry to determine total panel ports
	panelPortsByDimensions := o.Rows.ValueInt64() * o.Columns.ValueInt64()
	if panelPortsByDimensions != panelPortsByPortGroup {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("panel (%d rows of %d ports) has %d ports by dimensions, but the total by port group is %d",
				o.Rows.ValueInt64(), o.Columns.ValueInt64(), panelPortsByDimensions, panelPortsByPortGroup))
		return
	}
}

func NewLogicalDevicePanelList(ctx context.Context, in []design.LogicalDevicePanel, diags *diag.Diagnostics) types.List {
	panels := make([]logicalDevicePanel, len(in))
	for i, panel := range in {
		panels[i].loadApiData(ctx, panel, diags)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()})
		}
	}

	return value.ListOrNull(ctx, types.ObjectType{AttrTypes: logicalDevicePanel{}.attrTypes()}, panels, diags)
}
