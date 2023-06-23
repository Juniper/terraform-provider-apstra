package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of IDs of Property Sets imported into the Blueprint.",
				Computed:            true,
				ElementType:         types.StringType,
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
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Ids         types.Set    `tfsdk:"ids"`
	}

	// get the configuration
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
			resp.Diagnostics.AddError("failed creating blueprint client", err.Error())
		}
		return
	}

	// read all PropertySets from the API because there's no ID-list-only API option
	propertySets, err := bpClient.GetAllPropertySets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed retrieving property sets", err.Error())
		return
	}

	// organize the property set IDs into a set, shove 'em into the "config" object
	psIds := make([]attr.Value, len(propertySets))
	for i, propertySet := range propertySets {
		psIds[i] = types.StringValue(propertySet.Id.String())
	}
	config.Ids = types.SetValueMust(types.StringType, psIds)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
