package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
			"data": logicalDeviceDataAttributeSchema(),
		},
	}, nil
}

func (o *dataSourceLogicalDevice) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dLogicalDevice
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.Null && config.Id.Null) || (!config.Name.Null && !config.Id.Null) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'name' must be specified")
		return
	}
}

func (o *dataSourceLogicalDevice) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dLogicalDevice
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var logicalDevice *goapstra.LogicalDevice
	switch {
	case !config.Name.Null:
		logicalDevice, err = o.client.GetLogicalDeviceByName(ctx, config.Name.Value)
	case !config.Id.Null:
		logicalDevice, err = o.client.GetLogicalDevice(ctx, goapstra.ObjectId(config.Id.Value))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Logical Device", err.Error())
		return
	}

	newState := newLogicalDeviceFromApi(logicalDevice)

	// Set state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

type dLogicalDevice struct {
	Id   types.String `tfsdk:"id"`   // optional input
	Name types.String `tfsdk:"name"` // optional input
	Data types.Object `tfsdk:"data"`
}

func newLogicalDeviceFromApi(in *goapstra.LogicalDevice) *dLogicalDevice {
	return &dLogicalDevice{
		Id:   types.String{Value: string(in.Id)},
		Name: types.String{Value: in.Data.DisplayName},
		Data: parseApiLogicalDeviceToTypesObject(in.Data),
	}
}

func newLogicalDevicePanelPortGroups(in []goapstra.LogicalDevicePortGroup) []attr.Value {
	portGroups := make([]attr.Value, len(in))
	for i, pg := range in {
		roles := make([]attr.Value, len(pg.Roles.Strings()))
		for i, role := range pg.Roles.Strings() {
			roles[i] = types.String{Value: role}
		}
		portGroups[i] = types.Object{
			AttrTypes: logicalDevicePanelPortGroupsListElementAttrTypes(),
			Attrs: map[string]attr.Value{
				"port_count":     types.Int64{Value: int64(pg.Count)},
				"port_speed_bps": types.Int64{Value: pg.Speed.BitsPerSecond()},
				"port_roles": types.Set{
					ElemType: types.StringType,
					Elems:    roles,
				},
			},
		}
	}
	return portGroups
}

func newLogicalDevicePanel(in goapstra.LogicalDevicePanel) types.Object {
	return types.Object{
		AttrTypes: logicalDeviceDataPanelsListElementAttrTypes(),
		Attrs: map[string]attr.Value{
			"rows":    types.Int64{Value: int64(in.PanelLayout.RowCount)},
			"columns": types.Int64{Value: int64(in.PanelLayout.ColumnCount)},
			"port_groups": types.List{
				ElemType: logicalDeviceDataPanelPortGroupObjectSchema(),
				Elems:    newLogicalDevicePanelPortGroups(in.PortGroups),
			},
		},
	}
}

func newLogicalDevicePanels(in *goapstra.LogicalDeviceData) []attr.Value {
	out := make([]attr.Value, len(in.Panels))
	for i, panel := range in.Panels {
		out[i] = newLogicalDevicePanel(panel)
	}
	return out
}

func parseApiLogicalDeviceToTypesObject(in *goapstra.LogicalDeviceData) types.Object {
	return types.Object{
		AttrTypes: logicalDeviceDataElementAttrTypes(),
		Attrs: map[string]attr.Value{
			"name": types.String{Value: in.DisplayName},
			"panels": types.List{
				ElemType: logicalDeviceDataPanelObjectSchema(),
				Elems:    newLogicalDevicePanels(in),
			},
		},
	}
}

func logicalDeviceDataElementAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"panels": logicalDevicePanelSchema(),
	}
}

func logicalDeviceDataPanelsListElementAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rows":        types.Int64Type,
		"columns":     types.Int64Type,
		"port_groups": logicalDeviceDataPanelPortGroupsListSchema(),
	}
}

func logicalDevicePanelPortGroupsListElementAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port_count":     types.Int64Type,
		"port_speed_bps": types.Int64Type,
		"port_roles":     types.SetType{ElemType: types.StringType},
	}
}

func logicalDeviceDataPanelPortGroupObjectSchema() types.ObjectType {
	return types.ObjectType{
		AttrTypes: logicalDevicePanelPortGroupsListElementAttrTypes(),
	}
}

func logicalDeviceDataPanelPortGroupsListSchema() types.ListType {
	return types.ListType{
		ElemType: logicalDeviceDataPanelPortGroupObjectSchema(),
	}
}

func logicalDeviceDataPanelObjectSchema() types.ObjectType {
	return types.ObjectType{
		AttrTypes: logicalDeviceDataPanelsListElementAttrTypes(),
	}
}

func logicalDevicePanelSchema() attr.Type {
	return types.ListType{
		ElemType: logicalDeviceDataPanelObjectSchema(),
	}
}
