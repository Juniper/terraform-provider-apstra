package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/apstra/freeform"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceFreeformBlueprint{}
	_ datasourceWithSetClient            = &dataSourceFreeformBlueprint{}
	_ datasourceWithSetFfBpClientFunc    = &dataSourceFreeformBlueprint{}
)

type dataSourceFreeformBlueprint struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
}

func (o *dataSourceFreeformBlueprint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_blueprint"
}

func (o *dataSourceFreeformBlueprint) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceFreeformBlueprint) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This data source looks up summary details of a Freeform Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: freeform.Blueprint{}.DataSourceAttributes(),
	}
}

func (o *dataSourceFreeformBlueprint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config freeform.Blueprint
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiData *apstra.BlueprintStatus

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetBlueprintStatusByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Blueprint not found",
				fmt.Sprintf("Blueprint with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetBlueprintStatus(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Blueprint not found",
				fmt.Sprintf("Blueprint with ID %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Blueprint Status", err.Error())
		return
	}

	config.Id = types.StringValue(apiData.Id.String())
	config.Name = types.StringValue(apiData.Label)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceFreeformBlueprint) setClient(client *apstra.Client) {
	o.client = client
}

func (o *dataSourceFreeformBlueprint) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}
