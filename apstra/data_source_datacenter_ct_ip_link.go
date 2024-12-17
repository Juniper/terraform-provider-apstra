package tfapstra

import (
	"context"

	connectivitytemplate "github.com/Juniper/terraform-provider-apstra/apstra/connectivity_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceDatacenterCtIpLink struct{}

func (o *dataSourceDatacenterCtIpLink) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_ip_link"
}

func (o *dataSourceDatacenterCtIpLink) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterCtIpLink) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source composes a Connectivity Template Primitive as a JSON string, " +
			"suitable for use in the `primitives` attribute of an `apstra_datacenter_connectivity_template` " +
			"resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.",
		DeprecationMessage: "This data source will be removed in a future version. Please migrate your use of the " +
			"`apstra_datacenter_connectivity_template` resource (the likely reason this data source is being invoked) " +
			"to one of the new resources which do not depend on this data source: " +
			"`apstra_datacenter_connectivity_template_interface`, `apstra_datacenter_connectivity_template_loopback`, " +
			"`apstra_datacenter_connectivity_template_svi`, or `apstra_datacenter_connectivity_template_system`.",
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

	config.Primitive = types.StringValue(rendered)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
