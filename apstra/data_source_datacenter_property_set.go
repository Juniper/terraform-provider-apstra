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
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}
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
		} else {
			resp.Diagnostics.AddError("error creating blueprint client", err.Error())
		}
		return
	}

	var api *apstra.TwoStageL3ClosPropertySet
	switch {
	case !config.Label.IsNull():
		api, err = bpClient.GetPropertySetByName(ctx, config.Label.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"DatacenterPropertySet not found",
					fmt.Sprintf("DatacenterPropertySet with label %q not found", config.Label.ValueString()))
			} else {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"Error Getting DatacenterPropertySet",
					fmt.Sprintf("DatacenterPropertySet with label %q failed with error %q", config.Label.ValueString(), err.Error()))
			}
			return
		}
	case !config.Id.IsNull():
		api, err = bpClient.GetPropertySet(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"DatacenterPropertySet not found",
					fmt.Sprintf("DatacenterPropertySet with ID %q not found", config.Id.ValueString()))
			} else {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"DatacenterPropertySet not found",
					fmt.Sprintf("DatacenterPropertySet with ID %q failed with error %q", config.Id.ValueString(), err.Error()))
			}
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving DatacenterPropertySet", err.Error())
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
