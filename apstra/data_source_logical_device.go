package apstra

import (
	"context"
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceLogicalDeviceType struct{}

func (r dataSourceLogicalDeviceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"name": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"panels": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"rows": {
						Computed: true,
						Type:     types.Int64Type,
					},
					"columns": {
						Computed: true,
						Type:     types.Int64Type,
					},
					"port_groups": {
						Computed: true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"port_count": {
								Computed: true,
								Type:     types.Int64Type,
							},
							"port_speed_gbps": {
								Computed: true,
								Type:     types.Int64Type,
							},
							"port_roles": {
								Computed: true,
								Type:     types.SetType{ElemType: types.StringType},
							},
						}),
					},
				}),
			},
		},
	}, nil
}

func (r dataSourceLogicalDeviceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceLogicalDevice{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceLogicalDevice struct {
	p provider
}

func (r dataSourceLogicalDevice) ValidateConfig(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
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

func (r dataSourceLogicalDevice) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
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
			Rows:       types.Int64{Value: int64(p.PanelLayout.RowCount)},
			Columns:    types.Int64{Value: int64(p.PanelLayout.ColumnCount)},
			PortGroups: portGroups,
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
