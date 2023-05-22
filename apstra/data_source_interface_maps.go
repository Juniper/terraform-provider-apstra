package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceInterfaceMaps{}
var _ versionValidator = &dataSourceBlueprints{}

type dataSourceInterfaceMaps struct {
	client *apstra.Client
}

func (o *dataSourceInterfaceMaps) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_maps"
}

func (o *dataSourceInterfaceMaps) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceInterfaceMaps) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the ID numbers of Interface Maps.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"device_profile_id": schema.StringAttribute{
				MarkdownDescription: "Optional filter to select only Interface Maps associated with the specified Device Profile.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"logical_device_id": schema.StringAttribute{
				MarkdownDescription: "Optional filter to select only Interface Maps associated with the specified Logical Device.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
		},
	}
}

func (o *dataSourceInterfaceMaps) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config interfaceMaps
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ids []apstra.ObjectId
	var err error
	if config.LogicalDeviceId.IsNull() && config.DeviceProfileId.IsNull() {
		ids, err = o.client.ListAllInterfaceMapIds(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error listing Interface Map IDs", err.Error())
			return
		}
	} else {
		interfaceMaps, err := o.client.GetAllInterfaceMaps(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving Interface Maps", err.Error())
			return
		}
		for _, im := range interfaceMaps {
			if !config.LogicalDeviceId.IsNull() && config.LogicalDeviceId.ValueString() != im.Data.LogicalDeviceId.String() {
				continue
			}
			if !config.DeviceProfileId.IsNull() && config.DeviceProfileId.ValueString() != im.Data.DeviceProfileId.String() {
				continue
			}
			ids = append(ids, im.Id)
		}
	}

	idSet, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// save the list of IDs in the config object
	config.Ids = idSet

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type interfaceMaps struct {
	Ids             types.Set    `tfsdk:"ids"`
	LogicalDeviceId types.String `tfsdk:"logical_device_id"`
	DeviceProfileId types.String `tfsdk:"device_profile_id"`
}
