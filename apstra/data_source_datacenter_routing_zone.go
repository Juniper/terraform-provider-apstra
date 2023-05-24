package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ datasource.DataSourceWithConfigure = &dataSourceRoutingZone{}

type dataSourceRoutingZone struct {
	client *apstra.Client
}

func (o *dataSourceRoutingZone) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zone"
}

func (o *dataSourceRoutingZone) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceRoutingZone) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource returns details of a Routing Zone within a Datacenter Blueprint.",
		Attributes:          blueprint.DatacenterRoutingZone{}.DataSourceAttributes(),
	}
}

func (o *dataSourceRoutingZone) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	// Retrieve values from config.
	var config blueprint.DatacenterRoutingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(config.BlueprintId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found",
				config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("error creating blueprint client", err.Error())
		return
	}

	sz, err := bp.GetSecurityZone(ctx, apstra.ObjectId(config.Id.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.Diagnostics.AddError(fmt.Sprintf("security zone %s in blueprint %s not found",
				config.Id, config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("error retrieving security zone", err.Error())
		return
	}

	config.LoadApiData(ctx, sz.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dhcpServers, err := bp.GetSecurityZoneDhcpServers(ctx, sz.Id)
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
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
