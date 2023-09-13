package design

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type LogicalDevicePanel struct {
	Rows       types.Int64 `tfsdk:"rows"`
	Columns    types.Int64 `tfsdk:"columns"`
	PortGroups types.List  `tfsdk:"port_groups"`
}

func (o LogicalDevicePanel) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: LogicalDevicePanelPortGroup{}.DataSourceAttributes(),
			},
		},
	}
}

func (o LogicalDevicePanel) ResourceAttributes() map[string]resourceSchema.Attribute {
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
			Required:            true,
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: LogicalDevicePanelPortGroup{}.ResourceAttributes(),
			},
		},
	}
}

func (o LogicalDevicePanel) ResourceAttributesReadOnly() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"rows": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical vertical dimension of the panel.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"columns": resourceSchema.Int64Attribute{
			MarkdownDescription: "Physical horizontal dimension of the panel.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"port_groups": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
			Computed:            true,
			PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: LogicalDevicePanelPortGroup{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o LogicalDevicePanel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":        types.Int64Type,
		"columns":     types.Int64Type,
		"port_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: LogicalDevicePanelPortGroup{}.AttrTypes()}},
	}
}

func (o *LogicalDevicePanel) LoadApiData(ctx context.Context, in *apstra.LogicalDevicePanel, diags *diag.Diagnostics) {
	portGroups := make([]LogicalDevicePanelPortGroup, len(in.PortGroups))
	for i := range in.PortGroups {
		portGroups[i].LoadApiData(ctx, &in.PortGroups[i], diags)
		if diags.HasError() {
			return
		}
	}

	portGroupList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: LogicalDevicePanelPortGroup{}.AttrTypes()}, portGroups)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.Rows = types.Int64Value(int64(in.PanelLayout.RowCount))
	o.Columns = types.Int64Value(int64(in.PanelLayout.ColumnCount))
	o.PortGroups = portGroupList
}

func (o *LogicalDevicePanel) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.LogicalDevicePanel {
	tfPortGroups := make([]LogicalDevicePanelPortGroup, len(o.PortGroups.Elements()))
	diags.Append(o.PortGroups.ElementsAs(ctx, &tfPortGroups, false)...)
	if diags.HasError() {
		return nil
	}

	reqPortGroups := make([]apstra.LogicalDevicePortGroup, len(tfPortGroups))
	for i, pg := range tfPortGroups {
		roleStrings := make([]string, len(pg.PortRoles.Elements()))
		diags.Append(pg.PortRoles.ElementsAs(ctx, &roleStrings, false)...)
		if diags.HasError() {
			return nil
		}
		var reqRoles apstra.LogicalDevicePortRoleFlags
		err := reqRoles.FromStrings(roleStrings)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing port roles: '%s'", strings.Join(roleStrings, "','")), err.Error())
		}
		reqPortGroups[i] = apstra.LogicalDevicePortGroup{
			Count: int(pg.PortCount.ValueInt64()),
			Speed: apstra.LogicalDevicePortSpeed(pg.PortSpeed.ValueString()),
			Roles: reqRoles,
		}
	}

	return &apstra.LogicalDevicePanel{
		PanelLayout: apstra.LogicalDevicePanelLayout{
			RowCount:    int(o.Rows.ValueInt64()),
			ColumnCount: int(o.Columns.ValueInt64()),
		},
		PortIndexing: apstra.LogicalDevicePortIndexing{
			Order:      apstra.PortIndexingHorizontalFirst,
			StartIndex: 1,
			Schema:     apstra.PortIndexingSchemaAbsolute,
		},
		PortGroups: reqPortGroups,
	}
}

func (o *LogicalDevicePanel) Validate(ctx context.Context, i int, diags *diag.Diagnostics) {
	if o.Rows.IsUnknown() || o.Columns.IsUnknown() || o.PortGroups.IsUnknown() {
		return
	}

	portGroups := o.GetPortGroups(ctx, diags)
	if diags.HasError() {
		return
	}

	// count up the ports in each port group
	var panelPortsByPortGroup int64
	for _, portGroup := range portGroups {
		panelPortsByPortGroup = panelPortsByPortGroup + portGroup.PortCount.ValueInt64()
	}

	// use panel geometry to determine total panel ports
	panelPortsByDimensions := o.Rows.ValueInt64() * o.Columns.ValueInt64()
	if panelPortsByDimensions != panelPortsByPortGroup {
		diags.AddAttributeError(path.Root("panels").AtListIndex(i),
			errInvalidConfig,
			fmt.Sprintf("panel[%d] (%d rows of %d ports) has %d ports by dimensions, but the total by port group is %d",
				i, o.Rows.ValueInt64(), o.Columns.ValueInt64(), panelPortsByDimensions, panelPortsByPortGroup))
		return
	}
}

func (o *LogicalDevicePanel) GetPortGroups(ctx context.Context, diags *diag.Diagnostics) []LogicalDevicePanelPortGroup {
	portGroups := make([]LogicalDevicePanelPortGroup, len(o.PortGroups.Elements()))
	diags.Append(o.PortGroups.ElementsAs(ctx, &portGroups, false)...)
	if diags.HasError() {
		return nil
	}

	return portGroups
}

func NewLogicalDevicePanelList(ctx context.Context, in []apstra.LogicalDevicePanel, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: LogicalDevicePanel{}.AttrTypes()})
	}

	panels := make([]LogicalDevicePanel, len(in))
	for i, panel := range in {
		panels[i].LoadApiData(ctx, &panel, diags)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: LogicalDevicePanel{}.AttrTypes()})
		}
	}

	return utils.ListValueOrNull(ctx, types.ObjectType{AttrTypes: LogicalDevicePanel{}.AttrTypes()}, panels, diags)
}
