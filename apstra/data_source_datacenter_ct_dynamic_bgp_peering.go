package tfapstra

import (
	"context"

	connectivitytemplate "github.com/Juniper/terraform-provider-apstra/apstra/connectivity_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithValidateConfig = &dataSourceDatacenterCtDynamicBgpPeering{}

type dataSourceDatacenterCtDynamicBgpPeering struct{}

func (o *dataSourceDatacenterCtDynamicBgpPeering) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_dynamic_bgp_peering"
}

func (o *dataSourceDatacenterCtDynamicBgpPeering) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source composes a Connectivity Template Primitive as a JSON string, " +
			"suitable for use in the `primitives` attribute of an `apstra_datacenter_connectivity_template` " +
			"resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.",
		DeprecationMessage: "This data source will be removed in a future version. Please migrate your use of the " +
			"`apstra_datacenter_connectivity_template` resource (the likely reason this data source is being invoked) " +
			"to one of the new resources which do not depend on this data source: " +
			"`apstra_datacenter_connectivity_template_interface`, `apstra_datacenter_connectivity_template_loopback`, " +
			"`apstra_datacenter_connectivity_template_svi`, or `apstra_datacenter_connectivity_template_system`.",
		Attributes: connectivitytemplate.DynamicBgpPeering{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtDynamicBgpPeering) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config connectivitytemplate.DynamicBgpPeering
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Ipv4Enabled.IsUnknown() || config.Ipv6Enabled.IsUnknown() {
		return // cannot validate while either value is unknown
	}

	if !config.Ipv4Enabled.ValueBool() && !config.Ipv6Enabled.ValueBool() {
		resp.Diagnostics.AddError("Invalid Attribute Combination", "At least one attribute "+
			"out of 'ipv4_enabled' and 'ipv6_enabled' must be `true`")
	}
}

func (o *dataSourceDatacenterCtDynamicBgpPeering) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectivitytemplate.DynamicBgpPeering
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
