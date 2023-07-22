package tfapstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	connectivitytemplate "terraform-provider-apstra/apstra/connectivity_template"
)

var _ datasource.DataSource = &dataSourceDatacenterCtIpLink{}

type dataSourceDatacenterCtIpLink struct{}

func (o *dataSourceDatacenterCtIpLink) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_ip_link"
}

func (o *dataSourceDatacenterCtIpLink) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source composes a Connectivity Template JsonPrimitive suitable " +
			"for use in the `primitives` attribute of either an `apstra_datacenter_connectivity_template` " +
			"resource or the `children` attribute of a Different Connectivity Template JsonPrimitive.",
		Attributes: connectivitytemplate.IpLink{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtIpLink) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectivitytemplate.IpLink
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rendered := config.Marshal(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Primitive = types.StringValue(string(rendered))

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
