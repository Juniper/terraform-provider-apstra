package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSourceWithConfigure = &dataSourceTwoStageL3ClosBlueprint{}

type dataSourceTwoStageL3ClosBlueprint struct {
	client *goapstra.Client
}

func (o *dataSourceTwoStageL3ClosBlueprint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint_status"
}

func (o *dataSourceTwoStageL3ClosBlueprint) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = dataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceTwoStageL3ClosBlueprint) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source looks up summary details of a Datacenter Blueprint.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent upon the user to ensure the lookup criteria matches exactly one Datacenter Blueprint. " +
			"Matching zero or more Datacenter Blueprints will produce an error.",
		Attributes: dcBlueprintStatus{}.dataSourceAttributes(),
	}
}

func (o *dataSourceTwoStageL3ClosBlueprint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dcBlueprintStatus
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var status *goapstra.BlueprintStatus
	var ace goapstra.ApstraClientErr

	switch {
	case !config.Name.IsNull():
		status, err = o.client.GetBlueprintStatusByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Blueprint not found",
				fmt.Sprintf("Blueprint with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		status, err = o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Blueprint not found",
				fmt.Sprintf("Blueprint with ID %q not found", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Blueprint Status", err.Error())
		return
	}

	// create new state object
	var state dcBlueprintStatus
	state.loadApiData(ctx, status, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
