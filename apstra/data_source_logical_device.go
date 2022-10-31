package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceLogicalDevice{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceLogicalDevice{}

type dataSourceLogicalDevice struct {
	client *goapstra.Client
}

func (o *dataSourceLogicalDevice) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataDetail,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *dataSourceLogicalDevice) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logical_device"
}

func (o *dataSourceLogicalDevice) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details about a specific logical device " +
			"(a logical device is a template used by apstra when creating rack types (rack types are also templates)).\n\n" +
			"The logical device can be specified by id xor by name. " +
			"Returns an error if 0 matches or more than 1 match. " +
			"Note on looking up logical devices by name:\n" +
			"\n1. Apstra allows multiple logical devices to have the same name, although this is not recommended." +
			"\n1. To lookup a logical device that shares a name with any other device(s) you must lookup by id.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the logical device. Required when name is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Name of the logical device. Required when id is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"panels": dPanelsAttributeSchema(),
		},
	}, nil
}

func (o *dataSourceLogicalDevice) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config logicalDevice
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'name' must be specified")
		return
	}
}

func (o *dataSourceLogicalDevice) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config logicalDevice
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiResponse *goapstra.LogicalDevice
	switch {
	case !config.Name.IsNull():
		apiResponse, err = o.client.GetLogicalDeviceByName(ctx, config.Name.ValueString())
	case !config.Id.IsNull():
		apiResponse, err = o.client.GetLogicalDevice(ctx, goapstra.ObjectId(config.Id.ValueString()))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Logical Device", err.Error())
		return
	}

	var newState logicalDevice
	newState.parseApi(ctx, apiResponse, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func dPanelsAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details physical layout of interfaces on the device.",
		Computed:            true,
		PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"rows": {
				MarkdownDescription: "Physical vertical dimension of the panel.",
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Type:                types.Int64Type,
			},
			"columns": {
				MarkdownDescription: "Physical horizontal dimension of the panel.",
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Type:                types.Int64Type,
			},
			"port_groups": {
				MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"port_count": {
						MarkdownDescription: "Number of ports in the group.",
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Type:                types.Int64Type,
					},
					"port_speed_bps": {
						MarkdownDescription: "Port speed in Gbps.",
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Type:                types.Int64Type,
					},
					"port_roles": {
						MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Type:                types.SetType{ElemType: types.StringType},
					},
				}),
			},
		}),
	}
}

type logicalDevice struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Panels types.List   `tfsdk:"panels"`
}

func (o logicalDevice) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":     types.StringType,
			"name":   types.StringType,
			"panels": types.ListType{ElemType: logicalDevicePanel{}.attrType()}}}
}

func (o *logicalDevice) parseApi(ctx context.Context, in *goapstra.LogicalDevice, diags *diag.Diagnostics) {
	panels := make([]logicalDevicePanel, len(in.Data.Panels))
	for i := range in.Data.Panels {
		panels[i].parseApi(&in.Data.Panels[i])
	}

	var d diag.Diagnostics

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Panels, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: panelAttrTypes()}, panels)

	diags.Append(d...)
}

type logicalDevicePanel struct {
	Rows       int64                         `tfsdk:"rows"`
	Columns    int64                         `tfsdk:"columns"`
	PortGroups []logicalDevicePanelPortGroup `tfsdk:"port_groups"`
}

func (o logicalDevicePanel) attrType() attr.Type {
	var portGroups logicalDevicePanelPortGroup
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"rows":        types.Int64Type,
			"columns":     types.Int64Type,
			"port_groups": types.ListType{ElemType: portGroups.attrType()}}}
}

func (o *logicalDevicePanel) parseApi(in *goapstra.LogicalDevicePanel) {
	portGroups := make([]logicalDevicePanelPortGroup, len(in.PortGroups))
	for i := range in.PortGroups {
		portGroups[i] = *parseApiPanelPortGroup(in.PortGroups[i])
	}
	o.Rows = int64(in.PanelLayout.RowCount)
	o.Columns = int64(in.PanelLayout.ColumnCount)
	o.PortGroups = portGroups
}

type logicalDevicePanelPortGroup struct {
	PortCount    int64    `tfsdk:"port_count"`
	PortSpeedBps int64    `tfsdk:"port_speed_bps"`
	PortRoles    []string `tfsdk:"port_roles"`
}

func (o *logicalDevicePanelPortGroup) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"port_count":     types.Int64Type,
			"port_speed_bps": types.Int64Type,
			"port_roles":     types.SetType{ElemType: types.StringType}}}
}

func (o *logicalDevicePanelPortGroup) parseApi(in *goapstra.LogicalDevicePortGroup) {
	o.PortCount = int64(in.Count)
	o.PortSpeedBps = in.Speed.BitsPerSecond()
	o.PortRoles = in.Roles.Strings()
}

func parseApiPanelPortGroup(in goapstra.LogicalDevicePortGroup) *logicalDevicePanelPortGroup {
	return &logicalDevicePanelPortGroup{
		PortCount:    int64(in.Count),
		PortSpeedBps: in.Speed.BitsPerSecond(),
		PortRoles:    in.Roles.Strings(),
	}
}

func panelPortGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port_count":     types.Int64Type,
		"port_speed_bps": types.Int64Type,
		"port_roles":     types.SetType{ElemType: types.StringType}}
}

func panelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":        types.Int64Type,
		"columns":     types.Int64Type,
		"port_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: panelPortGroupAttrTypes()}},
	}
}
