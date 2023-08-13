package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"terraform-provider-apstra/apstra/resources"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceIpv4Pool{}

type dataSourceIpv4Pool struct {
	client *apstra.Client
}

func (o *dataSourceIpv4Pool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv4_pool"
}

func (o *dataSourceIpv4Pool) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceIpv4Pool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific IPv4 Pool.\n\n" +
			"At least one optional attribute is required.",
		Attributes: resources.Ipv4Pool{}.DataSourceAttributes(),
	}
}

func (o *dataSourceIpv4Pool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resources.Ipv4Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiData *apstra.IpPool

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetIp4PoolByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"IPv4 Pool not found",
				fmt.Sprintf("IPv4 Pool with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetIp4Pool(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"IPv4 Pool not found",
				fmt.Sprintf("IPv4 Pool with ID %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving IPv4 Pool", err.Error())
		return
	}

	// create new state object
	var state resources.Ipv4Pool
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
