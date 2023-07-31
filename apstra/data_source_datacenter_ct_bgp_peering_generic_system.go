package tfapstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	connectivitytemplate "terraform-provider-apstra/apstra/connectivity_template"
)

var _ datasource.DataSource = &dataSourceDatacenterCtBgpPeeringGenericSystem{}

type dataSourceDatacenterCtBgpPeeringGenericSystem struct{}

func (o *dataSourceDatacenterCtBgpPeeringGenericSystem) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_bgp_peering_generic_system"
}

func (o *dataSourceDatacenterCtBgpPeeringGenericSystem) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source composes a Connectivity Template Primitive as a JSON string, " +
			"suitable for use in the `primitives` attribute of either an `apstra_datacenter_connectivity_template` " +
			"resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.",
		Attributes: connectivitytemplate.BgpPeeringGenericSystem{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtBgpPeeringGenericSystem) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectivitytemplate.BgpPeeringGenericSystem
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
