package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/resources"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAsnPool{}

type dataSourceAsnPool struct {
	client *apstra.Client
}

func (o *dataSourceAsnPool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pool"
}

func (o *dataSourceAsnPool) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceAsnPool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific ASN Pool.\n\n" +
			"At least one optional attribute is required.",
		Attributes: resources.AsnPool{}.DataSourceAttributes(),
	}
}

func (o *dataSourceAsnPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resources.AsnPool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiData *apstra.AsnPool

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetAsnPoolByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"ASN Pool not found",
				fmt.Sprintf("ASN Pool with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetAsnPool(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"ASN Pool not found",
				fmt.Sprintf("ASN Pool with ID %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving ASN Pool", err.Error())
		return
	}

	// create new state object
	var state resources.AsnPool
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
