package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	connectivitytemplate "github.com/Juniper/terraform-provider-apstra/apstra/connectivity_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithValidateConfig = &dataSourceDatacenterCtBgpPeeringGenericSystem{}

type dataSourceDatacenterCtBgpPeeringGenericSystem struct{}

func (o *dataSourceDatacenterCtBgpPeeringGenericSystem) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_bgp_peering_generic_system"
}

func (o *dataSourceDatacenterCtBgpPeeringGenericSystem) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source composes a Connectivity Template Primitive as a JSON string, " +
			"suitable for use in the `primitives` attribute of an `apstra_datacenter_connectivity_template` " +
			"resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.",
		Attributes: connectivitytemplate.BgpPeeringGenericSystem{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtBgpPeeringGenericSystem) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config connectivitytemplate.BgpPeeringGenericSystem
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	v4NoneString := apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone.String() // "none"
	v6NoneString := apstra.CtPrimitiveIPv6ProtocolSessionAddressingNone.String() // "none"

	v4Unconfigured := config.Ipv4AddressingType.IsNull() || config.Ipv4AddressingType.ValueString() == v4NoneString
	v6Unconfigured := config.Ipv6AddressingType.IsNull() || config.Ipv6AddressingType.ValueString() == v6NoneString

	if v4Unconfigured && v6Unconfigured {
		resp.Diagnostics.AddError("Invalid Attribute Combination", "At least one attribute "+
			"out of 'ipv4_addressing_type' and 'ipv6_addressing_type' must be enabled")
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
