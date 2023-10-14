package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceIbaPredefinedProbes{}

type dataSourceIbaPredefinedProbes struct {
	client *apstra.Client
}

func (o *dataSourceIbaPredefinedProbes) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iba_predefined_probes"
}

func (o *dataSourceIbaPredefinedProbes) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceIbaPredefinedProbes) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the IDs of the IBA Predefined Probes in a Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint ID. " +
					"Used to identify the Blueprint that the IBA Predefined Probes belongs to.",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"names": schema.SetAttribute{
				MarkdownDescription: "A set of names representing the IBA Predefined Probes in the blueprint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceIbaPredefinedProbes) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Names       types.Set    `tfsdk:"names"`
	}

	// get the configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(config.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	ws, err := bpClient.GetAllIbaPredefinedProbes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving IBA Predefined Probes", err.Error())
		return
	}

	ids := make([]attr.Value, len(ws))
	for i, j := range ws {
		ids[i] = types.StringValue(j.Name)
	}
	nameSet := types.SetValueMust(types.StringType, ids)

	// create new state object
	state := struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Names       types.Set    `tfsdk:"names"`
	}{
		BlueprintId: config.BlueprintId,
		Names:       nameSet,
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
