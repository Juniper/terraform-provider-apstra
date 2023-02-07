package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func (o *dataSourceInterfaceMap) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Interface Map.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one Interface Map. " +
			"Matching zero Interface Maps or more than one Interface Map will produce an error.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Interface Map ID.  Required when the Interface Map name is omitted.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Interface Map name displayed in the Apstra web UI.  Required when Interface Map ID is omitted.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"logical_device_id": schema.StringAttribute{
				MarkdownDescription: "ID of Logical Device referenced by this interface map.",
				Computed:            true,
			},
			"device_profile_id": schema.StringAttribute{
				MarkdownDescription: "ID of Device Profile referenced by this interface map.",
				Computed:            true,
			},
			"interfaces": schema.SetNestedAttribute{
				MarkdownDescription: "Detailed mapping of each physical interface to its role in the logical device",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Physical device interface name",
							Computed:            true,
						},
						"roles": schema.SetAttribute{
							MarkdownDescription: "Logical Device role (\"connected to\") of the interface.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"position": schema.Int64Attribute{
							MarkdownDescription: "todo - need to find out what this is", // todo
							Computed:            true,
						},
						"active": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the interface is used by the Interface Map",
							Computed:            true,
						},
						"speed": schema.StringAttribute{
							MarkdownDescription: "Interface speed",
							Computed:            true,
						},
						"mapping": schema.SingleNestedAttribute{
							MarkdownDescription: "Mapping info for each physical interface",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"device_profile_port_id": schema.Int64Attribute{
									MarkdownDescription: "Port number(ID) from the Device Profile.",
									Computed:            true,
								},
								"device_profile_transformation_id": schema.Int64Attribute{
									MarkdownDescription: "Port-specific transform ID from the Device Profile.",
									Computed:            true,
								},
								"device_profile_interface_id": schema.Int64Attribute{
									MarkdownDescription: "Port-specific interface ID from the device profile (used to identify interfaces in breakout scenarios.)",
									Computed:            true,
								},
								"logical_device_panel": schema.Int64Attribute{
									MarkdownDescription: "Panel number (first panel is 1) of the Logical Device port which corresponds to this interface.",
									Computed:            true,
								},
								"logical_device_panel_port": schema.Int64Attribute{
									MarkdownDescription: "Port number (first port is 1) of the Logical Device port which corresponds to this interface.",
									Computed:            true,
								},
							},
						},
						"setting": schema.StringAttribute{
							MarkdownDescription: "Vendor specific commands needed to configure the interface, from the device profile.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (o *dataSourceInterfaceMap) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dInterfaceMap
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
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

	// create new state object
	newState := dInterfaceMap{}
	newState.loadApiResponse(ctx, interfaceMap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

type dInterfaceMap struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	LogicalDevice types.String `tfsdk:"logical_device_id"`
	DeviceProfile types.String `tfsdk:"device_profile_id"`
	Interfaces    types.Set    `tfsdk:"interfaces"`
}

func (o *dInterfaceMap) loadApiResponse(ctx context.Context, in *goapstra.InterfaceMap, diags *diag.Diagnostics) {
	var d diag.Diagnostics
	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.Label)
	o.LogicalDevice = types.StringValue(string(in.Data.LogicalDeviceId))
	o.DeviceProfile = types.StringValue(string(in.Data.DeviceProfileId))

	interfaces := make([]dInterfaceMapInterface, len(in.Data.Interfaces))
	for i := range in.Data.Interfaces {
		interfaces[i].loadApiResponse(ctx, &in.Data.Interfaces[i], diags)
	}
	o.Interfaces, d = types.SetValueFrom(ctx, dInterfaceMapInterface{}.attrType(), interfaces)
	diags.Append(d...)
}

type dInterfaceMapInterface struct {
	Name     types.String        `tfsdk:"name"`
	Roles    types.Set           `tfsdk:"roles"`
	Mapping  interfaceMapMapping `tfsdk:"mapping"`
	Active   types.Bool          `tfsdk:"active"`
	Position types.Int64         `tfsdk:"position"`
	Speed    types.String        `tfsdk:"speed"`
	Setting  types.String        `tfsdk:"setting"`
}

func (o dInterfaceMapInterface) attrType() attr.Type {
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

func (o *dInterfaceMapInterface) loadApiResponse(ctx context.Context, in *goapstra.InterfaceMapInterface, diags *diag.Diagnostics) {
	roles, d := types.SetValueFrom(ctx, types.StringType, in.Roles.Strings())
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.Name = types.StringValue(in.Name)
	o.Roles = roles
	o.Mapping.loadApiResponse(&in.Mapping)
	o.Active = types.BoolValue(bool(in.ActiveState))
	o.Position = types.Int64Value(int64(in.Position))
	o.Speed = types.StringValue(string(in.Speed))
	o.Setting = types.StringValue(in.Setting.Param)
}

type interfaceMapMapping struct {
	DPPort      types.Int64 `tfsdk:"device_profile_port_id"`
	DPTransform types.Int64 `tfsdk:"device_profile_transformation_id"`
	DPInterface types.Int64 `tfsdk:"device_profile_interface_id"`
	LDPanel     types.Int64 `tfsdk:"logical_device_panel"`
	LDPort      types.Int64 `tfsdk:"logical_device_panel_port"`
}

func (o interfaceMapMapping) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"device_profile_port_id":           types.Int64Type,
		"device_profile_transformation_id": types.Int64Type,
		"device_profile_interface_id":      types.Int64Type,
		"logical_device_panel":             types.Int64Type,
		"logical_device_panel_port":        types.Int64Type,
	}
}

func (o interfaceMapMapping) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes()}
}

func (o *interfaceMapMapping) loadApiResponse(in *goapstra.InterfaceMapMapping) {
	o.DPPort = types.Int64Value(int64(in.DPPortId))
	o.DPTransform = types.Int64Value(int64(in.DPTransformId))
	o.DPInterface = types.Int64Value(int64(in.DPInterfaceId))
	o.LDPanel = types.Int64Value(int64(in.LDPanel))
	o.LDPort = types.Int64Value(int64(in.LDPort))

	if o.LDPanel.ValueInt64() == -1 {
		o.LDPanel = types.Int64Null()
	}

	if o.LDPort.ValueInt64() == -1 {
		o.LDPort = types.Int64Null()
	}
}
