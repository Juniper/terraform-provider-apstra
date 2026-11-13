package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSourceWithConfigure = (*dataSourceDatacenterVirtualNetworkAssignment)(nil)
	_ datasourceWithSetDcBpClientFunc    = (*dataSourceDatacenterVirtualNetworkAssignment)(nil)
)

type dataSourceDatacenterVirtualNetworkAssignment struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (d *dataSourceDatacenterVirtualNetworkAssignment) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network_assignment"
}

func (d *dataSourceDatacenterVirtualNetworkAssignment) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, d, req, resp)
}

func (d *dataSourceDatacenterVirtualNetworkAssignment) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source retrieves Virtual Network to Leaf Switch and Access Switches assignment detail within a *Datacenter* Blueprint.",
		Attributes:          blueprint.VirtualNetworkAssignment{}.DatasourceAttributes(),
	}
}

func (d *dataSourceDatacenterVirtualNetworkAssignment) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from configuration.
	var state blueprint.VirtualNetworkAssignment
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := d.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintID), err.Error())
		return
	}

	ok := state.Read(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !ok {
		resp.Diagnostics.AddError("Not found",
			fmt.Sprintf("Assignment of Virtual Network %q to leaf switch %q was not found in Blueprint %q",
				state.VNID.ValueString(), state.LeafID.ValueString(), state.BlueprintID.ValueString()),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *dataSourceDatacenterVirtualNetworkAssignment) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	d.getBpClientFunc = f
}
