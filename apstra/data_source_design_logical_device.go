package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	adesign "github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDesignLogicalDevice{}
	_ datasourceWithSetClient            = &dataSourceDesignLogicalDevice{}
)

type dataSourceDesignLogicalDevice struct {
	client *apstra.Client
}

func (dld *dataSourceDesignLogicalDevice) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_design_logical_device"
}

func (dld *dataSourceDesignLogicalDevice) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, dld, req, resp)
}

func (dld *dataSourceDesignLogicalDevice) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This data source provides details of a Logical Device in the Apstra _Design_ tab.\n\n" +
			"At least one optional attribute is required.",
		Attributes: design.LogicalDevice{}.DataSourceAttributes(),
	}
}

func (dld *dataSourceDesignLogicalDevice) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config design.LogicalDevice
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var sdkLD adesign.LogicalDevice

	switch {
	case !config.Name.IsNull():
		sdkLD, err = dld.client.GetLogicalDeviceByLabel2(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Logical Device not found",
				fmt.Sprintf("Logical Device with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.ID.IsNull():
		sdkLD, err = dld.client.GetLogicalDevice2(ctx, config.ID.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Logical Device not found",
				fmt.Sprintf("Logical Device with id %q not found", config.ID.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Logical Device", err.Error())
		return
	}

	// create new state object
	config.LoadAPIData(ctx, sdkLD, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (dld *dataSourceDesignLogicalDevice) setClient(client *apstra.Client) {
	dld.client = client
}
