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

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterPropertySet{}

type dataSourceDatacenterPropertySet struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterPropertySet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_property_set"
}

func (o *dataSourceDatacenterPropertySet) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.getBpClientFunc = DataSourceGetTwoStageL3ClosClientFunc(ctx, req, resp)
}

func (o *dataSourceDatacenterPropertySet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source provides details of a specific Property Set imported into a Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.DatacenterPropertySet{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterPropertySet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.DatacenterPropertySet
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

	var api *apstra.TwoStageL3ClosPropertySet
	switch {
	case !config.Name.IsNull():
		api, err = bp.GetPropertySetByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"DatacenterPropertySet not found",
				fmt.Sprintf("DatacenterPropertySet with label %s not found", config.Name))
			return
		}
	case !config.Id.IsNull():
		api, err = bp.GetPropertySet(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"DatacenterPropertySet not found",
				fmt.Sprintf("DatacenterPropertySet with ID %s not found", config.Id))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed reading DatacenterPropertySet", err.Error())
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
