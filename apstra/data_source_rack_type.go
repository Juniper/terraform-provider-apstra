package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/design"
)

var _ datasource.DataSourceWithConfigure = &dataSourceRackType{}

type dataSourceRackType struct {
	client *apstra.Client
}

func (o *dataSourceRackType) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_type"
}

func (o *dataSourceRackType) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceRackType) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Rack Type.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one Rack Type. " +
			"Matching zero Rack Types or more than one Rack Type will produce an error.",
		Attributes: design.RackType{}.DataSourceAttributes(),
	}
}

func (o *dataSourceRackType) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config design.RackType
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.RackType
	var ace apstra.ApstraClientErr

	switch {
	case !config.Name.IsNull():
		api, err = o.client.GetRackTypeByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with name %q does not exist", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		api, err = o.client.GetRackType(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with ID %q does not exist", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Rack Type", err.Error())
		return
	}

	// catch problems which would crash the provider
	design.ValidateRackType(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create new state object
	var newState design.RackType
	newState.Id = types.StringValue(string(api.Id))
	newState.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}
