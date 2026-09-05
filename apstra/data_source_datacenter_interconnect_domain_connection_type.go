package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	ierrors "github.com/Juniper/terraform-provider-apstra/internal/errors"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterInterconnectDomainConnectionType{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterInterconnectDomainConnectionType{}
)

type dataSourceDatacenterInterconnectDomainConnectionType struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (d *dataSourceDatacenterInterconnectDomainConnectionType) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain_connection_type"
}

func (d *dataSourceDatacenterInterconnectDomainConnectionType) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, d, req, resp)
}

func (d *dataSourceDatacenterInterconnectDomainConnectionType) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source retrieves per-VN DCI Details within a Blueprint.",
		Attributes:          blueprint.InterconnectDomainConnectionType{}.DatasourceAttributes(),
	}
}

func (d *dataSourceDatacenterInterconnectDomainConnectionType) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from configuration.
	var state blueprint.InterconnectDomainConnectionType
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

	err = state.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(ierrors.ReadError(d), err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *dataSourceDatacenterInterconnectDomainConnectionType) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	d.getBpClientFunc = f
}
