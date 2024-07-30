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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceFreeformResourceGroup{}
	_ datasourceWithSetFfBpClientFunc    = &dataSourceFreeformResourceGroup{}
)

type dataSourceFreeformResourceGroup struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
}

func (o *dataSourceFreeformResourceGroup) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_resource_group"
}

func (o *dataSourceFreeformResourceGroup) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceFreeformResourceGroup) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This data source provides details of a specific Freeform Resource Group.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.FreeformResourceGroup{}.DataSourceAttributes(),
	}
}

func (o *dataSourceFreeformResourceGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.FreeformResourceGroup
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

	var api *apstra.FreeformRaGroup
	switch {
	case !config.Id.IsNull():
		api, err = bp.GetRaGroup(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Freeform Resource Group not found",
				fmt.Sprintf("Freeform Resource Group with ID %s not found", config.Id))
			return
		}
	case !config.Name.IsNull():
		api, err = bp.GetRaGroupByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Freeform Resource Group not found",
				fmt.Sprintf("Freeform resource allocation group with name %s not found", config.Name))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("failed reading Freeform Resource Group", err.Error())
		return
	}
	if api.Data == nil {
		resp.Diagnostics.AddError("failed reading Freeform Resource Group", "api response has no payload")
		return
	}

	config.Id = types.StringValue(api.Id.String())
	config.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceFreeformResourceGroup) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}
