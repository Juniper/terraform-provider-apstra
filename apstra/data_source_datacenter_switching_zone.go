package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithConfigure      = (*dataSourceDatacenterSwitchingZone)(nil)
	_ datasource.DataSourceWithValidateConfig = (*dataSourceDatacenterSwitchingZone)(nil)
	_ datasourceWithSetDcBpClientFunc         = (*dataSourceDatacenterSwitchingZone)(nil)
	_ datasourceWithSetClient                 = (*dataSourceDatacenterSwitchingZone)(nil)
)

type dataSourceDatacenterSwitchingZone struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterSwitchingZone) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_switching_zone"
}

func (o *dataSourceDatacenterSwitchingZone) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterSwitchingZone) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource returns details of a Switching Zone within a Datacenter Blueprint.\n\n" +
			"At least one optional attribute is required. Requires Apstra " + compatibility.SwitchingZoneOK.String() + ".",
		Attributes: blueprint.DatacenterSwitchingZone{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterSwitchingZone) ValidateConfig(_ context.Context, _ datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	// cannot proceed if the resource has not been configured
	if o.client == nil {
		return
	}

	apiVersion, err := version.NewVersion(o.client.ApiVersion())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse API version %q", o.client.ApiVersion()), err.Error())
		return
	}

	if !compatibility.SwitchingZoneOK.Check(apiVersion) {
		resp.Diagnostics.AddError(
			"Resource requires Apstra "+compatibility.SwitchingZoneOK.String(),
			"Resource requires Apstra "+compatibility.SwitchingZoneOK.String(),
		)
		return
	}
}

func (o *dataSourceDatacenterSwitchingZone) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config blueprint.DatacenterSwitchingZone
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, config.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, config.BlueprintID), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, config.BlueprintID), err.Error())
		return
	}

	var api datacenter.SwitchingZone
	switch {
	case !config.ID.IsNull():
		api, err = bp.GetSwitchingZone(ctx, config.ID.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Switching Zone not found",
				fmt.Sprintf("Switching Zone with ID %s not found", config.ID))
			return
		}
	case !config.Name.IsNull():
		api, err = bp.GetSwitchingZoneByLabel(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Switching Zone not found",
				fmt.Sprintf("Switching Zone with Name %s not found", config.Name))
			return
		}
	case !config.MACVRFName.IsNull():
		api, err = bp.GetSwitchingZoneByMACVRFName(ctx, config.MACVRFName.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("mac_vrf_name"),
				"Switching Zone not found",
				fmt.Sprintf("Switching Zone with MAC VRF Name %s not found", config.Name))
			return
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("failed reading Switching Zone", err.Error())
		return
	}
	if api.ID() == nil {
		resp.Diagnostics.AddError("failed reading Switching Zone", "Zone ID does not exist")
	}

	config.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterSwitchingZone) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *dataSourceDatacenterSwitchingZone) setClient(client *apstra.Client) {
	o.client = client
}
