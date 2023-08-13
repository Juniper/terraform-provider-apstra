package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterRoutingZone{}

type dataSourceDatacenterRoutingZone struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterRoutingZone) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zone"
}

func (o *dataSourceDatacenterRoutingZone) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterRoutingZone) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource returns details of a Routing Zone within a Datacenter Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.DatacenterRoutingZone{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterRoutingZone) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(config.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found",
				config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(blueprint.ErrDCBlueprintCreate, config.BlueprintId), err.Error())
		return
	}

	var api *apstra.SecurityZone
	switch {
	case !config.Id.IsNull():
		api, err = bpClient.GetSecurityZone(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"Routing Zone  not found",
					fmt.Sprintf("Routing Zone with ID %s not found", config.Id))
				return
			}
			resp.Diagnostics.AddError(
				"Failed reading Routing Zone", err.Error(),
			)
			return
		}
	case !config.Name.IsNull():
		api, err = bpClient.GetSecurityZoneByVrfName(ctx, config.Name.ValueString())
		if err != nil {
			if utils.IsApstra404(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"Routing Zone not found",
					fmt.Sprintf("Routing Zone with Name %s not found", config.Name))
				return
			}
			resp.Diagnostics.AddError(
				"Failed reading Routing Zone", err.Error(),
			)
			return
		}
	}

	config.Id = types.StringValue(api.Id.String())
	config.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dhcpServers, err := bpClient.GetSecurityZoneDhcpServers(ctx, api.Id)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error retrieving security zone", err.Error())
		return
	}

	config.LoadApiDhcpServers(ctx, dhcpServers, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
