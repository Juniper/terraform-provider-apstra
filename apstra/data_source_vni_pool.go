package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/resources"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ datasource.DataSourceWithConfigure = &dataSourceVniPool{}
var _ datasourceWithSetClient = &dataSourceVniPool{}

type dataSourceVniPool struct {
	client *apstra.Client
}

func (o *dataSourceVniPool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vni_pool"
}

func (o *dataSourceVniPool) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceVniPool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryResources + "This data source provides details of a specific VNI Pool.\n\n" +
			"At least one optional attribute is required.",
		Attributes: resources.VniPool{}.DataSourceAttributes(),
	}
}

func (o *dataSourceVniPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resources.VniPool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiData *apstra.VniPool

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetVniPoolByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"VNI Pool not found",
				fmt.Sprintf("VNI Pool with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetVniPool(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"VNI Pool not found",
				fmt.Sprintf("VNI Pool with ID %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving VNI Pool", err.Error())
		return
	}

	// create new state object
	var state resources.VniPool
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceVniPool) setClient(client *apstra.Client) {
	o.client = client
}
