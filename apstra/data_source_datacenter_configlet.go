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

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterConfiglet{}

type dataSourceDatacenterConfiglet struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterConfiglet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_configlet"
}

func (o *dataSourceDatacenterConfiglet) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterConfiglet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Configlet imported into a Blueprint." +
			"\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.DatacenterConfiglet{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterConfiglet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.DatacenterConfiglet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(config.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found",
				config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	var api *apstra.TwoStageL3ClosConfiglet
	switch {
	case !config.Name.IsNull():
		api, err = bpClient.GetConfigletByName(ctx, config.Name.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"Datacenter Configlet not found",
					fmt.Sprintf("DatacenterConfiglet with label %s not found", config.Name))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading Datacenter Configlet", err.Error(),
			)
			return
		}
	case !config.Id.IsNull():
		api, err = bpClient.GetConfiglet(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"DatacenterConfiglet not found",
					fmt.Sprintf("DatacenterConfiglet with ID %s not found", config.Id))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading DatacenterConfiglet", err.Error(),
			)
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}

	// create new state object
	var state blueprint.DatacenterConfiglet
	state.BlueprintId = config.BlueprintId
	state.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
