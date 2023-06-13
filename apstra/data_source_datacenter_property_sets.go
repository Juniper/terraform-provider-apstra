package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterPropertySets{}

type dataSourceDatacenterPropertySets struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterPropertySets) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_property_sets"
}

func (o *dataSourceDatacenterPropertySets) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterPropertySets) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the ID numbers of all Property Sets in the Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"property_sets": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.DatacenterPropertySet{}.DataSourceAttributes(),
				},
				Computed:            true,
				MarkdownDescription: "A set of Apstra Imported Property Sets Imported into the Blueprint",
			},
		},
	}
}

func (o *dataSourceDatacenterPropertySets) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config struct {
		BlueprintId  types.String `tfsdk:"blueprint_id"`
		PropertySets types.Set    `tfsdk:"property_sets"`
	}

	// get the configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(config.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Blueprint Client", err.Error())
		return
	}
	dps, err := bpClient.GetAllPropertySets(ctx)
	ps := make([]blueprint.DatacenterPropertySet, len(dps))
	for i, j := range dps {
		ps[i].LoadApiData(ctx, &j, &resp.Diagnostics)
		ps[i].BlueprintId = config.BlueprintId
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving PropertySet", err.Error())
		return
	}
	psSet := utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: blueprint.DatacenterPropertySet{}.AttrTypes()}, ps, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// create new state object
	var state struct {
		BlueprintId  types.String `tfsdk:"blueprint_id"`
		PropertySets types.Set    `tfsdk:"property_sets"`
	}
	state.BlueprintId = config.BlueprintId
	state.PropertySets = psSet
	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
