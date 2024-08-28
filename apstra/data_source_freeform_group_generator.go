package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/freeform"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceFreeformGroupGenerator{}
	_ datasourceWithSetFfBpClientFunc    = &dataSourceFreeformGroupGenerator{}
)

type dataSourceFreeformGroupGenerator struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
}

func (o *dataSourceFreeformGroupGenerator) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_group_generator"
}

func (o *dataSourceFreeformGroupGenerator) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceFreeformGroupGenerator) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This data source provides details of a specific Freeform Group Generator.\n\n" +
			"At least one optional attribute is required.",
		Attributes: freeform.GroupGenerator{}.DataSourceAttributes(),
	}
}

func (o *dataSourceFreeformGroupGenerator) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config freeform.GroupGenerator
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, config.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	var api *apstra.FreeformGroupGenerator
	switch {
	case !config.Id.IsNull():
		api, err = bp.GetGroupGenerator(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Freeform Group Generator not found",
				fmt.Sprintf("Freeform Group Generator with ID %s not found", config.Id))
			return
		}
	case !config.Name.IsNull():
		api, err = bp.GetGroupGeneratorByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Freeform Group Generator not found",
				fmt.Sprintf("Freeform Group Generator with Name %s not found", config.Name))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("failed reading Freeform Group Generator", err.Error())
		return
	}
	if api.Data == nil {
		resp.Diagnostics.AddError("failed reading Freeform Group Generator", "api response has no payload")
		return
	}

	config.Id = types.StringValue(api.Id.String())
	config.LoadApiData(ctx, api.Data)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceFreeformGroupGenerator) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}
