package tfapstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	connectivitytemplate "terraform-provider-apstra/apstra/connectivity_template"
)

var _ datasource.DataSource = &dataSourceDatacenterCtStaticRoute{}

type dataSourceDatacenterCtStaticRoute struct{}

func (o *dataSourceDatacenterCtStaticRoute) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_static_route"
}

func (o *dataSourceDatacenterCtStaticRoute) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source composes a Connectivity Template Primitive suitable " +
			"for use in the `primitives` attribute of either an `apstra_datacenter_connectivity_template` " +
			"resource or the `children` attribute of a Different Connectivity Template Primitive.",
		Attributes: connectivitytemplate.StaticRoute{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtStaticRoute) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectivitytemplate.StaticRoute
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rendered := config.Render(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Primitive = types.StringValue(string(rendered))

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
