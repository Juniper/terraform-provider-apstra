package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ datasource.DataSourceWithConfigure = &dataSourceVniPool{}

type dataSourceVniPool struct {
	client *goapstra.Client
}

func (o *dataSourceVniPool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vni_pool"
}

func (o *dataSourceVniPool) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = dataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceVniPool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific VNI Pool.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent upon the user to ensure the lookup criteria matches exactly one VNI Pool. " +
			"Matching zero or more VNI Pools will produce an error.",
		Attributes: vniPool{}.dataSourceAttributes(),
	}
}

func (o *dataSourceVniPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config vniPool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var apiData *goapstra.VniPool
	var ace goapstra.ApstraClientErr

	switch {
	case !config.Name.IsNull():
		apiData, err = o.client.GetVniPoolByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"VNI Pool not found",
				fmt.Sprintf("VNI Pool with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		apiData, err = o.client.GetVniPool(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"VNI Pool not found",
				fmt.Sprintf("VNI Pool with ID %q not found", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving VNI Pool", err.Error())
		return
	}

	// create new state object
	var state vniPool
	state.loadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
