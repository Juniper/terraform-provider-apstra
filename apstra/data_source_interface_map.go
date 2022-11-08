package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceInterfaceMap{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceInterfaceMap{}

type dataSourceInterfaceMap struct {
	client *goapstra.Client
}

func (o *dataSourceInterfaceMap) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_map"
}

func (o *dataSourceInterfaceMap) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceInterfaceMap) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details of a specific Interface Map.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one Interface Map. " +
			"Matching zero Interface Maps or more than one Interface Map will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Interface Map ID.  Required when the Interface Map name is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Interface Map name displayed in the Apstra web UI.  Required when Interface Map ID is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"logical_device_id": {
				MarkdownDescription: "ID of Logical Device referenced by this interface map.",
				Computed:            true,
				Type:                types.StringType,
			},
			"device_profile_id": {
				MarkdownDescription: "ID of Device Profile referenced by this interface map.",
				Computed:            true,
				Type:                types.StringType,
			},
			"interfaces": {
				MarkdownDescription: "Detailed mapping of each physical interface to its role in the logical device",
				Computed:            true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Physical device interface name",
						Type:                types.StringType,
						Computed:            true,
					},
					"roles": {
						MarkdownDescription: "Logical Device role (\"connected to\") of the interface.",
						Type:                types.SetType{ElemType: types.StringType},
						Computed:            true,
					},
					"position": {
						MarkdownDescription: "todo - need to find out what this is", // todo
						Type:                types.Int64Type,
						Computed:            true,
					},
					"active": {
						MarkdownDescription: "Indicates whether the interface is used by the Interface Map",
						Type:                types.BoolType,
						Computed:            true,
					},
					"speed": {
						MarkdownDescription: "Interface speed",
						Type:                types.StringType,
						Computed:            true,
					},
					"mapping": {
						MarkdownDescription: "Mapping info for each physical interface",
						Computed:            true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"device_profile_port_id": {
								MarkdownDescription: "Port number(ID) from the Device Profile.",
								Type:                types.Int64Type,
								Computed:            true,
							},
							"device_profile_transformation_id": {
								MarkdownDescription: "Port-specific transform ID from the Device Profile.",
								Type:                types.Int64Type,
								Computed:            true,
							},
							"device_profile_interface_id": {
								MarkdownDescription: "Port-specific interface ID from the device profile (used to identify interfaces in breakout scenarios.)",
								Type:                types.Int64Type,
								Computed:            true,
							},
							"logical_device_panel": {
								MarkdownDescription: "Panel number (first panel is 1) of the Logical Device port which corresponds to this interface.",
								Type:                types.Int64Type,
								Computed:            true,
							},
							"logical_device_port": {
								MarkdownDescription: "Port number (first port is 1) of the Logical Device port which corresponds to this interface.",
								Type:                types.Int64Type,
								Computed:            true,
							},
						}),
					},
					"setting": {
						MarkdownDescription: "Vendor specific commands needed to configure the interface, from the device profile.",
						Type:                types.StringType,
						Computed:            true,
					},
				}),
			},
		},
	}, nil
}

func (o *dataSourceInterfaceMap) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dInterfaceMap
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

func (o *dataSourceInterfaceMap) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dInterfaceMap
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var interfaceMap *goapstra.InterfaceMap
	var ace goapstra.ApstraClientErr

	// maybe the config gave us the interface map name?
	if !config.Name.IsNull() { // fetch rack type by name
		interfaceMap, err = o.client.GetInterfaceMapByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Interface Map not found",
				fmt.Sprintf("Interface Map with name '%s' does not exist", config.Name.ValueString()))
			return
		}
	}

	// maybe the config gave us the interface map id?
	if !config.Id.IsNull() { // fetch rack type by ID
		interfaceMap, err = o.client.GetInterfaceMap(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Interface Map not found",
				fmt.Sprintf("Interface Map with id '%s' does not exist", config.Id.ValueString()))
			return
		}
	}

	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Interface Map", err.Error())
		return
	}

	newState := &dInterfaceMap{}
	newState.parseApi(ctx, interfaceMap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//Set state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

type dInterfaceMap struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	LogicalDevice types.String `tfsdk:"logical_device_id"`
	DeviceProfile types.String `tfsdk:"device_profile_id"`
	Interfaces    types.Set    `tfsdk:"interfaces"`
}

func (o *dInterfaceMap) parseApi(ctx context.Context, in *goapstra.InterfaceMap, diags *diag.Diagnostics) {
	var d diag.Diagnostics
	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.Label)
	o.LogicalDevice = types.StringValue(string(in.Data.LogicalDeviceId))
	o.DeviceProfile = types.StringValue(string(in.Data.DeviceProfileId))

	interfaces := make([]interfaceMapInterface, len(in.Data.Interfaces))
	for i := range in.Data.Interfaces {
		interfaces[i].parseApi(&in.Data.Interfaces[i])
	}
	o.Interfaces, d = types.SetValueFrom(ctx, interfaceMapInterface{}.attrType(), interfaces)
	diags.Append(d...)
}

type interfaceMapInterface struct {
	Name     string              `tfsdk:"name"`
	Roles    []string            `tfsdk:"roles"`
	Mapping  interfaceMapMapping `tfsdk:"mapping"`
	Active   bool                `tfsdk:"active"`
	Position int                 `tfsdk:"position"`
	Speed    string              `tfsdk:"speed"`
	Setting  string              `tfsdk:"setting"`
}

func (o interfaceMapInterface) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":     types.StringType,
			"roles":    types.SetType{ElemType: types.StringType},
			"mapping":  interfaceMapMapping{}.attrType(),
			"active":   types.BoolType,
			"position": types.Int64Type,
			"speed":    types.StringType,
			"setting":  types.StringType,
		}}
}

func (o *interfaceMapInterface) parseApi(in *goapstra.InterfaceMapInterface) {
	o.Name = in.Name
	o.Roles = in.Roles.Strings()
	o.Mapping.parseApi(&in.Mapping)
	o.Active = bool(in.ActiveState)
	o.Position = in.Position
	o.Speed = string(in.Speed)
	o.Setting = in.Setting.Param
}

type interfaceMapMapping struct {
	DPPort      int64 `tfsdk:"device_profile_port_id"`
	DPTransform int64 `tfsdk:"device_profile_transformation_id"`
	DPInterface int64 `tfsdk:"device_profile_interface_id"`
	LDPanel     int64 `tfsdk:"logical_device_panel"`
	LDPort      int64 `tfsdk:"logical_device_port"`
}

func (o interfaceMapMapping) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"device_profile_port_id":           types.Int64Type,
		"device_profile_transformation_id": types.Int64Type,
		"device_profile_interface_id":      types.Int64Type,
		"logical_device_panel":             types.Int64Type,
		"logical_device_port":              types.Int64Type,
	}
}

func (o interfaceMapMapping) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes()}
}

func (o *interfaceMapMapping) parseApi(in *goapstra.InterfaceMapMapping) {
	o.DPPort = int64(in.DPPortId)
	o.DPTransform = int64(in.DPTransformId)
	o.DPInterface = int64(in.DPInterfaceId)
	o.LDPanel = int64(in.LDPanel)
	o.LDPort = int64(in.LDPort)
}
