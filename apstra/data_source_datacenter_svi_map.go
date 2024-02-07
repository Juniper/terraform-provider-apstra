package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterSvis{}
var _ datasourceWithSetBpClientFunc = &dataSourceDatacenterSvis{}

type dataSourceDatacenterSvis struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterSvis) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_svis_map"
}

func (o *dataSourceDatacenterSvis) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterSvis) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns a maps of Sets of SVI info keyed by Virtual Network ID, System ID and SVI ID.",
		Attributes:          blueprint.DatacenterSvis{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterSvis) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// fetch config
	var config blueprint.DatacenterSvis
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

	// prepare and execute a graph query which returns details of all SVIs
	svis, query := config.GetSviInfo(ctx, bp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// organize SVIs into maps keyed by SVI ID, system ID, Network ID
	modelsByInterface := make(map[string]attr.Value)
	modelsByNetwork := make(map[string][]attr.Value)
	modelsBySystem := make(map[string][]attr.Value)
	for _, svi := range svis {
		val, d := types.ObjectValueFrom(ctx, blueprint.SviMapEntry{}.AttrTypes(), &svi)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		modelsByInterface[svi.Id.ValueString()] = val
		modelsByNetwork[svi.NetworkId.ValueString()] = append(modelsByNetwork[svi.NetworkId.ValueString()], val)
		modelsBySystem[svi.SystemId.ValueString()] = append(modelsBySystem[svi.SystemId.ValueString()], val)
	}

	// convert the "by network" map of slices to a map of sets to match the schema
	setsByNetwork := make(map[string]attr.Value)
	for id, vals := range modelsByNetwork {
		setsByNetwork[id] = types.SetValueMust(types.ObjectType{AttrTypes: blueprint.SviMapEntry{}.AttrTypes()}, vals)
	}

	// convert the "by system" map of slices to a map of sets to match the schema
	setsBySystem := make(map[string]attr.Value)
	for id, vals := range modelsBySystem {
		setsBySystem[id] = types.SetValueMust(types.ObjectType{AttrTypes: blueprint.SviMapEntry{}.AttrTypes()}, vals)
	}

	// fill the required values
	config.GraphQuery = types.StringValue(query.String())
	config.InterfaceToSvi = types.MapValueMust(types.ObjectType{AttrTypes: blueprint.SviMapEntry{}.AttrTypes()}, modelsByInterface)
	config.SystemToSvi = types.MapValueMust(types.SetType{ElemType: types.ObjectType{AttrTypes: blueprint.SviMapEntry{}.AttrTypes()}}, setsBySystem)
	config.NetworkToSvi = types.MapValueMust(types.SetType{ElemType: types.ObjectType{AttrTypes: blueprint.SviMapEntry{}.AttrTypes()}}, setsByNetwork)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterSvis) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
