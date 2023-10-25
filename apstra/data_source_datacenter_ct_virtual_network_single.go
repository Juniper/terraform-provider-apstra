package tfapstra

import (
	"context"
	connectivitytemplate "github.com/Juniper/terraform-provider-apstra/apstra/connectivity_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &dataSourceDatacenterCtVnSingle{}

type dataSourceDatacenterCtVnSingle struct{}

func (o *dataSourceDatacenterCtVnSingle) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_virtual_network_single"
}

func (o *dataSourceDatacenterCtVnSingle) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source composes a Connectivity Template Primitive as a JSON string, " +
			"suitable for use in the `primitives` attribute of an `apstra_datacenter_connectivity_template` " +
			"resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.",
		Attributes: connectivitytemplate.VnSingle{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtVnSingle) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectivitytemplate.VnSingle
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rendered := config.Marshal(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Primitive = types.StringValue(rendered)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
