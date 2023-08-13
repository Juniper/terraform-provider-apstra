package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterBlueprint{}

type dataSourceDatacenterBlueprint struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterBlueprint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint"
}

func (o *dataSourceDatacenterBlueprint) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterBlueprint) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source looks up summary details of a Datacenter Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.Blueprint{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterBlueprint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.Blueprint
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

	// create new state object
	var state blueprint.Blueprint
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
