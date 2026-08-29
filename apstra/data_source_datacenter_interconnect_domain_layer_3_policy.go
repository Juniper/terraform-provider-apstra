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
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterInterconnectDomainL3Policy{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterInterconnectDomainL3Policy{}
)

type dataSourceDatacenterInterconnectDomainL3Policy struct {
	lockFunc        func(context.Context, string) error
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (r *dataSourceDatacenterInterconnectDomainL3Policy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain_layer_3_policy"
}

func (r *dataSourceDatacenterInterconnectDomainL3Policy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, r, req, resp)
}

func (r *dataSourceDatacenterInterconnectDomainL3Policy) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source retrieve details of an Interconnect Domain Layer 3 Policy within a Blueprint.",
		Attributes:          blueprint.InterconnectDomainL3Policy{}.DatasourceAttributes(),
	}
}

func (r *dataSourceDatacenterInterconnectDomainL3Policy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from configuration.
	var state blueprint.InterconnectDomainL3Policy
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := r.getBpClientFunc(ctx, state.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, state.BlueprintID), err.Error())
		return
	}

	state.Read(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataSourceDatacenterInterconnectDomainL3Policy) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	r.getBpClientFunc = f
}
