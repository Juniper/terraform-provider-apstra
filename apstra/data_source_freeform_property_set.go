package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSourceWithConfigure = &dataSourceFreeformPropertySet{}
var _ datasourceWithSetClient = &dataSourceFreeformPropertySet{}

type dataSourceFreeformPropertySet struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
	client          *apstra.Client
}

func (o *dataSourceFreeformPropertySet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_property_set"
}

func (o *dataSourceFreeformPropertySet) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceFreeformPropertySet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This data source provides details of a specific Freeform PropertySet.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.FreeformPropertySet{}.DataSourceAttributes(),
	}
}
func (o *dataSourceFreeformPropertySet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state blueprint.FreeformPropertySet

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", state.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	api, err := bp.GetPropertySet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving PropertySet", err.Error())
		return
	}

	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceFreeformPropertySet) setClient(client *apstra.Client) {
	o.client = client
}
