package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	adevice "github.com/Juniper/apstra-go-sdk/device"
	"github.com/Juniper/terraform-provider-apstra/apstra/device"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDeviceProfile{}
var _ datasourceWithSetClient = &dataSourceDeviceProfile{}

type dataSourceDeviceProfile struct {
	client *apstra.Client
}

func (o dataSourceDeviceProfile) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_profile"
}

func (o *dataSourceDeviceProfile) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o dataSourceDeviceProfile) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDevices + "This data source provides details of a specific Device Profile from the " +
			"Global Catalog. Because of the limited attributes returned by this resource, it is primarily useful for discovering " +
			"the ID of a Device Profile when the name is known to the practitioner.\n\n" +
			"At least one optional attribute is required.",
		Attributes: device.Profile{}.DataSourceAttributes(),
	}
}

func (o dataSourceDeviceProfile) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config device.Profile
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api adevice.Profile

	switch {
	case !config.Name.IsNull():
		api, err = o.client.GetDeviceProfileByLabel(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Device Profile not found",
				fmt.Sprintf("Device Profile with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.ID.IsNull():
		api, err = o.client.GetDeviceProfile(ctx, config.ID.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Device Profile not found",
				fmt.Sprintf("Device Profile with id %q not found", config.ID.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Device Profile", err.Error())
		return
	}

	// load API response and set state
	config.LoadAPIData(ctx, api, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDeviceProfile) setClient(client *apstra.Client) {
	o.client = client
}
