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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterRoutingZoneConstraint{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterRoutingZoneConstraint{}
)

type dataSourceDatacenterRoutingZoneConstraint struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterRoutingZoneConstraint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zone_constraint"
}

func (o *dataSourceDatacenterRoutingZoneConstraint) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterRoutingZoneConstraint) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource returns details of a Routing Zone Constraint within a Datacenter Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.DatacenterRoutingZoneConstraint{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterRoutingZoneConstraint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterRoutingZoneConstraint
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

	var api *apstra.RoutingZoneConstraint
	switch {
	case !config.Id.IsNull():
		api, err = bp.GetRoutingZoneConstraint(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Routing Zone not found",
				fmt.Sprintf("Routing Zone Constraint with ID %s not found", config.Id))
			return
		}
	case !config.Name.IsNull():
		api, err = bp.GetRoutingZoneConstraintByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Routing Zone not found",
				fmt.Sprintf("Routing Zone Constraint with Name %s not found", config.Name))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("failed reading Routing Zone Constraint", err.Error())
		return
	}
	if api == nil || api.Data == nil {
		resp.Diagnostics.AddError("failed reading Routing Zone Constraint", "api response has no payload")
		return
	}

	config.Id = types.StringValue(api.Id.String())
	config.LoadApiData(ctx, *api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterRoutingZoneConstraint) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
