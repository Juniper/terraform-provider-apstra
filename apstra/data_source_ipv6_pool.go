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
)

var _ datasource.DataSourceWithConfigure = &dataSourceIpv6Pool{}
var _ datasourceWithSetClient = &dataSourceIpv6Pool{}

type dataSourceIpv6Pool struct {
	client *apstra.Client
}

func (o *dataSourceIpv6Pool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_pool"
}

func (o *dataSourceIpv6Pool) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceIpv6Pool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryResources + "This data source provides details of a specific IPv6 Pool.\n\n" +
			"At least one optional attribute is required.",
		Attributes: resources.Ipv6Pool{}.DataSourceAttributes(),
	}
}

func (o *dataSourceIpv6Pool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resources.Ipv6Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiData *apstra.IpPool

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetIp6PoolByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"IPv6 Pool not found",
				fmt.Sprintf("IPv6 Pool with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetIp6Pool(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"IPv6 Pool not found",
				fmt.Sprintf("IPv6 Pool with ID %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving IPv6 Pool", err.Error())
		return
	}

	// create new state object
	var state resources.Ipv6Pool
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceIpv6Pool) setClient(client *apstra.Client) {
	o.client = client
}
