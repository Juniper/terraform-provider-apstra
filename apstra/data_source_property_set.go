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

var _ datasource.DataSourceWithConfigure = &dataSourcePropertySet{}

type dataSourcePropertySet struct {
	client *apstra.Client
}

func (o *dataSourcePropertySet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_property_set"
}

func (o *dataSourcePropertySet) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourcePropertySet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific PropertySet.\n\n" +
			"At least one optional attribute is required. ",
		Attributes: design.PropertySet{}.DataSourceAttributes(),
	}
}

func (o *dataSourcePropertySet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config design.PropertySet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.PropertySet
	var ace apstra.ApstraClientErr

	switch {
	case !config.Label.IsNull():
		api, err = o.client.GetPropertySetByLabel(ctx, config.Label.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"PropertySet not found",
				fmt.Sprintf("PropertySet with label %q not found", config.Label.ValueString()))
			return
		}
	case !config.Id.IsNull():
		api, err = o.client.GetPropertySet(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"PropertySet not found",
				fmt.Sprintf("PropertySet with ID %q not found", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving PropertySet", err.Error())
		return
	}

	// create new state object
	var state design.PropertySet
	state.Id = types.StringValue(string(api.Id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
