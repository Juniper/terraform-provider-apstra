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

//var _ datasource.DataSourceWithValidateConfig = &dataSourceLogicalDevice{}

type dataSourceLogicalDevice struct {
	client *goapstra.Client
}

func (o *dataSourceLogicalDevice) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceLogicalDevice) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "apstra_logical_device"
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
			"data": {
				MarkdownDescription: "Logical Device data which can be cloned into rack objects.",
				Computed:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Name of the logical device.",
						Computed:            true,
						Type:                types.StringType,
					},
					//"panels": {
					//	MarkdownDescription: "Detail connectivity features of the logical device.",
					//	Computed:            true,
					//	Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					//		"rows": {
					//			MarkdownDescription: "Physical vertical dimension of the panel.",
					//			Computed:            true,
					//			Type:                types.Int64Type,
					//		},
					//		"columns": {
					//			MarkdownDescription: "Physical horizontal dimension of the panel.",
					//			Computed:            true,
					//			Type:                types.Int64Type,
					//		},
					//	}),
					//},
					"panels": {
						MarkdownDescription: "Detail connectivity features of the logical device.",
						Computed:            true,
						Type:                logicalDevicePanelSchema(),
						//Type: types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
						//	"rows":    types.Int64Type,
						//	"columns": types.Int64Type,
						//}}},
					},
					//Type: types.ListType{
					//	ElemType: types.ObjectType{
					//		AttrTypes: map[string]attr.Type{
					//			"rows":    types.Int64Type,
					//			"columns": types.Int64Type,
					//		},
					//	},
					//},
					//"port_groups": {
					//	MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
					//	Computed:            true,
					//	Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					//		"port_count": {
					//			MarkdownDescription: "Number of ports in the group.",
					//			Computed:            true,
					//			Type:                types.Int64Type,
					//		},
					//		"port_speed_gbps": {
					//			MarkdownDescription: "Port speed in Gbps.",
					//			Computed:            true,
					//			Type:                types.Int64Type,
					//		},
					//		"port_roles": {
					//			MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
					//			Computed:            true,
					//			Type:                types.SetType{ElemType: types.StringType},
					//		},
					//	}),
					//},
				}),
			},
		},
	}, nil
}

func (o *dataSourceLogicalDevice) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config DLogicalDevice
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
	var config DLogicalDevice
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
	//config.Id = types.String{Value: string(logicalDevice.Id)}
	//config.Name = types.String{Value: logicalDevice.Data.DisplayName}
	//config.Data = newLogicalDeviceData(logicalDevice)
	//config.Data = types.Object{
	//	Attrs: map[string]attr.Value{
	//		"name": types.String{},
	//	},
	//	DisplayName: types.String{},
	//	Panels:      make([]LogicalDevicePanel, len(logicalDevice.Panels)),
	//}
	//for i, panel := range logicalDevice.Panels {
	//	config.Data.Attrs["panels"].(types.List).Elems[i] = types.Object{
	//		AttrTypes: map[string]attr.Type{
	//			"rows":    types.Int64Type,
	//			"columns": types.Int64Type,
	//		},
	//		Attrs: map[string]attr.Value{
	//			"rows":    types.Int64{Value: int64(panel.PanelLayout.RowCount)},
	//			"columns": types.Int64{Value: int64(panel.PanelLayout.ColumnCount)},
	//		},
	//	}
	//}
	//for i, panel := range logicalDevice.Panels {
	//	config.Data.Panels[i].AttrTypes = map[string]attr.Type{
	//		"rows":    types.Int64Type,
	//		"columns": types.Int64Type,
	//		//"port_groups": types.ListType{},
	//	}
	//	//var portGroups []logicalDevicePortGroup
	//	//for _, pg := range p.PortGroups {
	//	//	var roles []types.String
	//	//	for _, role := range pg.Roles.Strings() {
	//	//		roles = append(roles, types.String{Value: role})
	//	//	}
	//	//	portGroups = append(portGroups, logicalDevicePortGroup{
	//	//		Count: types.Int64{Value: int64(pg.Count)},
	//	//		Speed: types.Int64{Value: pg.Speed.BitsPerSecond() / 1000 / 1000 / 1000},
	//	//		Roles: roles,
	//	//	})
	//	//}
	//	//config.Data.Panels[i].Attrs = map[string]attr.Value{
	//	//	"rows":    types.Int64{Value: int64(panel.PanelLayout.RowCount)},
	//	//	"columns": types.Int64{Value: int64(panel.PanelLayout.ColumnCount)},
	//	//}
	//}

	// Set state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

type DLogicalDevice struct {
	Id   types.String `tfsdk:"id"`   // optional input
	Name types.String `tfsdk:"name"` // optional input
	Data types.Object `tfsdk:"data"`
}

type LogicalDeviceData struct {
	DisplayName types.String `tfsdk:"name"`
	//Panels      []LogicalDevicePanel `tfsdk:"panels"`
}

type LogicalDevicePanel struct {
	Rows    types.Int64 `tfsdk:"rows"`
	Columns types.Int64 `tfsdk:"columns"`
	//PortGroups []LogicalDevicePortGroup `tfsdk:"port_groups"`
}

type LogicalDevicePortGroup struct {
	Count types.Int64    `tfsdk:"port_count"`
	Speed types.Int64    `tfsdk:"port_speed_gbps"`
	Roles []types.String `tfsdk:"port_roles"`
}

func newLogicalDeviceFromApi(in *goapstra.LogicalDevice) *DLogicalDevice {
	return &DLogicalDevice{
		Id:   types.String{Value: string(in.Id)},
		Name: types.String{Value: in.Data.DisplayName},
		Data: newLogicalDeviceData(in),
	}
}

func logicalDevicePanelSchema() attr.Type {
	return types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"rows":    types.Int64Type,
				"columns": types.Int64Type,
			},
		},
	}
}

func newLogicalDevicePanel(in goapstra.LogicalDevicePanel) types.Object {
	return types.Object{
		AttrTypes: map[string]attr.Type{
			"rows":    types.Int64Type,
			"columns": types.Int64Type,
		},
		Attrs: map[string]attr.Value{
			"rows":    types.Int64{Value: int64(in.PanelLayout.RowCount)},
			"columns": types.Int64{Value: int64(in.PanelLayout.ColumnCount)},
		},
	}
}

func newLogicalDevicePanels(in *goapstra.LogicalDevice) []attr.Value {
	out := make([]attr.Value, len(in.Data.Panels))
	for i, panel := range in.Data.Panels {
		out[i] = newLogicalDevicePanel(panel)
	}
	return out
}

func newLogicalDeviceData(in *goapstra.LogicalDevice) types.Object {
	panels := make([]types.Object, len(in.Data.Panels))
	for i, panel := range in.Data.Panels {
		panels[i] = newLogicalDevicePanel(panel)
	}

	return types.Object{
		AttrTypes: map[string]attr.Type{
			"name":   types.StringType,
			"panels": logicalDevicePanelSchema(),
			//"panels": types.ListType{
			//	ElemType: types.ObjectType{
			//		AttrTypes: map[string]attr.Type{
			//			"rows":    types.Int64Type,
			//			"columns": types.Int64Type,
			//		},
			//	},
			//},
		},
		Attrs: map[string]attr.Value{
			"name": types.String{Value: in.Data.DisplayName},
			"panels": types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"rows":    types.Int64Type,
						"columns": types.Int64Type,
					},
				},
				Elems: newLogicalDevicePanels(in),
			},
		},
	}
}

// 	state.Ratings2 = types.List{
//		Elems: make([]attr.Value, len(apiResponse.Ratings)),
//		ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
//			"source": types.StringType,
//			"value":  types.StringType,
//		}},
//	}
//	for i, rating := range apiResponse.Ratings {
//		state.Ratings2.Elems[i] = types.Object{
//			AttrTypes: map[string]attr.Type{
//				"source": types.StringType,
//				"value":  types.StringType,
//			},
//			Attrs: map[string]attr.Value{
//				"source": types.String{Value: rating.Source},
//				"value":  types.String{Value: rating.Value},
//			},
//		}
//	}
