package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
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
					"port_speed": {
						MarkdownDescription: "Port speed.",
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
						Type:                types.StringType,
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
