package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceLogicalDevice{}

type dataSourceLogicalDevice struct {
	client *goapstra.Client
}

func (o *dataSourceLogicalDevice) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logical_device"
}

func (o *dataSourceLogicalDevice) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = dataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceLogicalDevice) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Logical Device.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent upon the user to ensure the lookup criteria matches exactly one Logical Device. " +
			"Matching zero or more Logical Devices will produce an error.",
		Attributes: logicalDevice{}.dataSourceAttributes(),
	}
}

func (o *dataSourceLogicalDevice) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config logicalDevice
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var ld *goapstra.LogicalDevice
	var ace goapstra.ApstraClientErr

	switch {
	case !config.Name.IsNull():
		ld, err = o.client.GetLogicalDeviceByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Logical Device not found",
				fmt.Sprintf("Logical Device with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		ld, err = o.client.GetLogicalDevice(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Logical Device not found",
				fmt.Sprintf("Logical Device with id %q not found", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Logical Device", err.Error())
		return
	}

	// create new state object
	var state logicalDevice
	state.Id = types.StringValue(string(ld.Id))
	state.loadApiData(ctx, ld.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//// read-only schema for logical device panels is a stand-alone function because
//// it gets re-used by rack-type and template data sources
//func dPanelAttributeSchema() schema.ListNestedAttribute {
//	return schema.ListNestedAttribute{
//		MarkdownDescription: "Details physical layout of interfaces on the device.",
//		Computed:            true,
//		NestedObject: schema.NestedAttributeObject{
//			Attributes: map[string]schema.Attribute{
//				"rows": schema.Int64Attribute{
//					MarkdownDescription: "Physical vertical dimension of the panel.",
//					Computed:            true,
//				},
//				"columns": schema.Int64Attribute{
//					MarkdownDescription: "Physical horizontal dimension of the panel.",
//					Computed:            true,
//				},
//				"port_groups": schema.ListNestedAttribute{
//					MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
//					Computed:            true,
//					NestedObject: schema.NestedAttributeObject{
//						Attributes: map[string]schema.Attribute{
//							"port_count": schema.Int64Attribute{
//								MarkdownDescription: "Number of ports in the group.",
//								Computed:            true,
//							},
//							"port_speed": schema.StringAttribute{
//								MarkdownDescription: "Port speed.",
//								Computed:            true,
//							},
//							"port_roles": schema.SetAttribute{
//								MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
//								Computed:            true,
//								ElementType:         types.StringType,
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//}
