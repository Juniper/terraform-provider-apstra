package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDeprecatedLogicalDevice{}
	_ datasourceWithSetClient            = &dataSourceDeprecatedLogicalDevice{}
)

type dataSourceDeprecatedLogicalDevice struct {
	client *apstra.Client
}

func (o *dataSourceDeprecatedLogicalDevice) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logical_device"
}

func (o *dataSourceDeprecatedLogicalDevice) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDeprecatedLogicalDevice) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: "This resource will be deprecated in a future release, no earlier than v2.0.0. " +
			"Users are encouraged to migrate their configurations to use `apstra_design_logical_device`, which is a drop-in replacement.",
		MarkdownDescription: docCategoryDesign + "This data source provides details of a specific Logical Device.\n\n" +
			"At least one optional attribute is required.",
		Attributes: design.DeprecatedLogicalDevice{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDeprecatedLogicalDevice) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config design.DeprecatedLogicalDevice
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.LogicalDevice

	switch {
	case !config.Name.IsNull():
		api, err = o.client.GetLogicalDeviceByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Logical Device not found",
				fmt.Sprintf("Logical Device with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		api, err = o.client.GetLogicalDevice(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Logical Device not found",
				fmt.Sprintf("Logical Device with id %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Logical Device", err.Error())
		return
	}

	// create new state object
	var state design.DeprecatedLogicalDevice
	state.Id = types.StringValue(string(api.Id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceDeprecatedLogicalDevice) setClient(client *apstra.Client) {
	o.client = client
}
