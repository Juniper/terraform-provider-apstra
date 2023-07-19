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

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterPropertySet{}

type dataSourceDatacenterPropertySet struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterPropertySet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_property_set"
}

func (o *dataSourceDatacenterPropertySet) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterPropertySet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Property Set imported into a Blueprint.\n\n" +
			"At least one optional attribute is required. ",
		Attributes: blueprint.DatacenterPropertySet{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterPropertySet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.DatacenterPropertySet
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

	var api *apstra.TwoStageL3ClosPropertySet
	switch {
	case !config.Name.IsNull():
		api, err = bpClient.GetPropertySetByName(ctx, config.Name.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"DatacenterPropertySet not found",
					fmt.Sprintf("DatacenterPropertySet with label %s not found", config.Name))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading DatacenterPropertySet", err.Error(),
			)
			return
		}
	case !config.Id.IsNull():
		api, err = bpClient.GetPropertySet(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"DatacenterPropertySet not found",
					fmt.Sprintf("DatacenterPropertySet with ID %s not found", config.Id))
				return
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"), "Failed reading DatacenterPropertySet", err.Error(),
			)
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}

	// create new state object
	var state blueprint.DatacenterPropertySet
	state.BlueprintId = config.BlueprintId
	state.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
