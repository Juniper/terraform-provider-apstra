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

var _ datasource.DataSourceWithConfigure = &dataSourceInterfaceMap{}

type dataSourceInterfaceMap struct {
	client *apstra.Client
}

func (o *dataSourceInterfaceMap) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_map"
}

func (o *dataSourceInterfaceMap) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceInterfaceMap) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryResources + "This data source provides details of a specific Interface Map.\n\n" +
			"At least one optional attribute is required.",
		Attributes: design.InterfaceMap{}.DataSourceAttributes(),
	}
}

func (o *dataSourceInterfaceMap) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config design.InterfaceMap
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.InterfaceMap

	switch {
	case !config.Name.IsNull():
		api, err = o.client.GetInterfaceMapByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Interface Map not found",
				fmt.Sprintf("Interface Map with name %q does not exist", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		api, err = o.client.GetInterfaceMap(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Interface Map not found",
				fmt.Sprintf("Interface Map with id %q does not exist", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Interface Map", err.Error())
		return
	}

	// create new state object
	newState := design.InterfaceMap{}
	newState.Id = types.StringValue(string(api.Id))
	newState.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}
