package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterInterconnectDomainGateway{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterInterconnectDomainGateway{}
)

type dataSourceDatacenterInterconnectDomainGateway struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterInterconnectDomainGateway) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_interconnect_domain_gateway"
}

func (o *dataSourceDatacenterInterconnectDomainGateway) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterInterconnectDomainGateway) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource returns details of an Interconnect Domain Gateway within a Datacenter Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.InterconnectDomainGateway{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterInterconnectDomainGateway) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config blueprint.InterconnectDomainGateway
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, config.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, config.BlueprintId), err.Error())
		return
	}

	err = config.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		if utils.IsApstra404(err) {
			switch config.Id.IsNull() {
			case true:
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"External Gateway not found",
					fmt.Sprintf("Blueprint %q External Gateway with Name %s not found", bp.Id(), config.Name))
			case false:
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"External Gateway not found",
					fmt.Sprintf("Blueprint %q External Gateway with ID %s not found", bp.Id(), config.Id))
			}
			return
		}
		resp.Diagnostics.AddError("Failed to fetch External Gateway", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterInterconnectDomainGateway) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
