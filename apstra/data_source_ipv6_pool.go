package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"terraform-provider-apstra/apstra/resources"
)

var _ datasource.DataSourceWithConfigure = &dataSourceIpv6Pool{}

type dataSourceIpv6Pool struct {
	client *apstra.Client
}

func (o *dataSourceIpv6Pool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_pool"
}

func (o *dataSourceIpv6Pool) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceIpv6Pool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific IPv6 Pool.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent upon the user to ensure the lookup criteria matches exactly one IPv6 Pool. " +
			"Matching zero or more IPv6 Pools will produce an error.",
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
	var ace apstra.ApstraClientErr

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetIp6PoolByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"IPv6 Pool not found",
				fmt.Sprintf("IPv6 Pool with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetIp6Pool(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"IPv6 Pool not found",
				fmt.Sprintf("IPv6 Pool with ID %q not found", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
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
