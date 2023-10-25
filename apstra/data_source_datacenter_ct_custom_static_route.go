package tfapstra

import (
	"context"
	"fmt"
	connectivitytemplate "github.com/Juniper/terraform-provider-apstra/apstra/connectivity_template"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

var _ datasource.DataSourceWithValidateConfig = &dataSourceDatacenterCtCustomStaticRoute{}

type dataSourceDatacenterCtCustomStaticRoute struct{}

func (o *dataSourceDatacenterCtCustomStaticRoute) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ct_custom_static_route"
}

func (o *dataSourceDatacenterCtCustomStaticRoute) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source composes a Connectivity Template Primitive as a JSON string, " +
			"suitable for use in the `primitives` attribute of an `apstra_datacenter_connectivity_template` " +
			"resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.",
		Attributes: connectivitytemplate.CustomStaticRoute{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterCtCustomStaticRoute) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config connectivitytemplate.CustomStaticRoute
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We're checking if these attributes are incompatible. If either are null/unknown, there's nothing to do.
	if !utils.Known(config.Network) || !utils.Known(config.NextHop) {
		return
	}

	// extract network as *net.IPNet
	_, network, err := net.ParseCIDR(config.Network.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("network"),
			fmt.Sprintf("failed parsing network %s", config.Network),
			err.Error())
		return
	}

	// extract nextHop as net.IP
	nextHop := net.ParseIP(config.NextHop.ValueString())
	if nextHop == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("network"),
			fmt.Sprintf("failed parsing next hop %s", config.Network),
			err.Error())
		return
	}

	// both should have the same length (either 4 or 16 bytes)
	if len(network.IP.To4()) != len(nextHop.To4()) {
		resp.Diagnostics.AddError("invalid attribute combination",
			fmt.Sprintf("'network' and 'next_hop' must be same type (IPv4 or IPv6), got %q and %q",
				config.Network, config.NextHop))
	}
}

func (o *dataSourceDatacenterCtCustomStaticRoute) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectivitytemplate.CustomStaticRoute
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
