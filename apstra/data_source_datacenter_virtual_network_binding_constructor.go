package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ datasource.DataSourceWithConfigure = &dataSourceVirtualNetworkBindingConstructor{}

type dataSourceVirtualNetworkBindingConstructor struct {
	client *apstra.Client
}

func (o *dataSourceVirtualNetworkBindingConstructor) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network_binding_constructor"
}

func (o *dataSourceVirtualNetworkBindingConstructor) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceVirtualNetworkBindingConstructor) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source can be used to calculate the " +
			"`bindings` data required by `apstra_datacenter_virtual_network`." +
			"\n\n" +
			"Given a list of switch node IDSs, it determines whether they're " +
			"leaf or access nodes, replaces individual switch IDs with ESI " +
			"or MLAG redundancy group IDs, finds required parent leaf " +
			"switches of all access switches.",
		Attributes: blueprint.VnBindingConstructor{}.DataSourceAttributes(),
	}
}

func (o *dataSourceVirtualNetworkBindingConstructor) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.VnBindingConstructor
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Compute(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
