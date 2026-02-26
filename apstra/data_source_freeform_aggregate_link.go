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
	_ datasource.DataSourceWithConfigure = (*dataSourceFreeformAggregateLink)(nil)
	_ datasourceWithSetFfBpClientFunc    = (*dataSourceFreeformAggregateLink)(nil)
)

type dataSourceFreeformAggregateLink struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
}

func (o dataSourceFreeformAggregateLink) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_aggregate_link"
}

func (o *dataSourceFreeformAggregateLink) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o dataSourceFreeformAggregateLink) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This data source provides details of a specific Aggregate Link.\n\n" +
			"At least one optional attribute is required.",
		Attributes: freeform.AggregateLink{}.DataSourceAttributes(),
	}
}

func (o dataSourceFreeformAggregateLink) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config freeform.AggregateLink
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, config.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Blueprint %s not found", config.BlueprintID), err.Error())
			return
		}
		resp.Diagnostics.AddError("Failed to create Blueprint client", err.Error())
		return
	}

	var api apstra.FreeformAggregateLink
	switch {
	case !config.ID.IsNull():
		api, err = bp.GetAggregateLink(ctx, config.ID.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Freeform Aggregate Link not found",
				fmt.Sprintf("Freeform Aggregate Link with ID %s not found", config.ID))
			return
		}
	case !config.Name.IsNull():
		api, err = bp.GetAggregateLinkByLabel(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Freeform Aggregate Link not found",
				fmt.Sprintf("Freeform Aggregate Link with Name %s not found", config.Name))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed reading Freeform Aggregate Link", err.Error())
		return
	}

	if config.ID.IsUnknown() && api.ID() != nil {
		config.ID = types.StringValue(*api.ID())
	}
	config.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceFreeformAggregateLink) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}
