package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceLogicalDeviceType struct{}

func (r dataSourceLogicalDevice) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "apstra_logical_device"
}

func (r dataSourceLogicalDevice) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"panels": {
				MarkdownDescription: "Detail connectivity features of the logical device.",
				Computed:            true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"rows": {
						MarkdownDescription: "Physical vertical dimension of the panel.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"columns": {
						MarkdownDescription: "Physical horizontal dimension of the panel.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"port_groups": {
						MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
						Computed:            true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"port_count": {
								MarkdownDescription: "Number of ports in the group.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"port_speed_gbps": {
								MarkdownDescription: "Port speed in Gbps.",
								Computed:            true,
								Type:                types.Int64Type,
							},
							"port_roles": {
								MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
								Computed:            true,
								Type:                types.SetType{ElemType: types.StringType},
							},
						}),
					},
				}),
			},
		},
	}, nil
}

func (r dataSourceLogicalDeviceType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceLogicalDevice{
		p: *(p.(*Provider)),
	}, nil
}

type dataSourceLogicalDevice struct {
	p Provider
}

func (r dataSourceLogicalDevice) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config DataLogicalDevice
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

func (r dataSourceLogicalDevice) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataLogicalDevice
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var logicalDevice *goapstra.LogicalDevice
	if config.Name.Null == false {
		logicalDevice, err = r.p.client.GetLogicalDeviceByName(ctx, config.Name.Value)
	}
	if config.Id.Null == false {
		logicalDevice, err = r.p.client.GetLogicalDevice(ctx, goapstra.ObjectId(config.Id.Value))
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Logical Device", err.Error())
		return
	}

	var panels []LogicalDevicePanel
	for _, p := range logicalDevice.Panels {
		var portGroups []LogicalDevicePortGroup
		for _, pg := range p.PortGroups {
			var roles []types.String
			for _, role := range pg.Roles.Strings() {
				roles = append(roles, types.String{Value: role})
			}
			portGroups = append(portGroups, LogicalDevicePortGroup{
				Count: types.Int64{Value: int64(pg.Count)},
				Speed: types.Int64{Value: pg.Speed.BitsPerSecond() / 1000 / 1000 / 1000},
				Roles: roles,
			})
		}
		panels = append(panels, LogicalDevicePanel{
			// todo: restore
			Rows: types.Int64{Value: int64(p.PanelLayout.RowCount)},
			//Columns: types.Int64{Value: int64(p.PanelLayout.ColumnCount)},
			//PortGroups: portGroups,
		})
	}

	// Set state
	diags = resp.State.Set(ctx, &DataLogicalDevice{
		Id:     types.String{Value: string(logicalDevice.Id)},
		Name:   types.String{Value: logicalDevice.DisplayName},
		Panels: panels,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
