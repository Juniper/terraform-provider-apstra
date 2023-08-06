package tfapstra

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	connectivitytemplate "terraform-provider-apstra/apstra/connectivity_template"
)

var _ datasource.DataSourceWithValidateConfig = &dataSourceDatacenterCtBgpPeeringIpEndpoint{}

type dataSourceDatacenterCtBgpPeeringIpEndpoint struct{}

func (o *dataSourceDatacenterCtBgpPeeringIpEndpoint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_bgp_peering_ip_endpoint"
}

func (o *dataSourceDatacenterCtBgpPeeringIpEndpoint) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source composes a Connectivity Template Primitive as a JSON string, " +
			"suitable for use in the `primitives` attribute of an `apstra_datacenter_connectivity_template` " +
			"resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.",
		Attributes: connectivitytemplate.BgpPeeringIpEndpoint{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtBgpPeeringIpEndpoint) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config connectivitytemplate.BgpPeeringIpEndpoint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Ipv4AfiEnabled.ValueBool() && !config.Ipv6AfiEnabled.ValueBool() {
		resp.Diagnostics.AddError("invalid attribute combination",
			fmt.Sprintf("at least one of %q and %q must have value 'true'",
				path.Root("ipv4_afi_enabled"), path.Root("ipv6_afi_enabled")))
	}
}

func (o *dataSourceDatacenterCtBgpPeeringIpEndpoint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectivitytemplate.BgpPeeringIpEndpoint
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
