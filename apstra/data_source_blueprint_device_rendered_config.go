package tfapstra

import (
	"context"
	"errors"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceBlueprintNodeConfig{}
	_ datasourceWithSetClient            = &dataSourceBlueprintNodeConfig{}
)

type dataSourceBlueprintNodeConfig struct {
	client *apstra.Client
}

func (o *dataSourceBlueprintNodeConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_device_rendered_config"
}

func (o *dataSourceBlueprintNodeConfig) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintNodeConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny +
			"This data source retrieves rendered device configuration for a system in a Blueprint.",
		Attributes: blueprint.RenderedConfig{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintNodeConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config blueprint.RenderedConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bpId := apstra.ObjectId(config.BlueprintId.ValueString())

	var err error
	var deployed, staged, incremental string
	var ace apstra.ClientErr

	switch {
	case !config.NodeId.IsNull():
		node := apstra.ObjectId(config.NodeId.ValueString())
		deployed, err = o.client.GetNodeRenderedConfig(ctx, bpId, node, enum.RenderedConfigTypeDeployed)
		if err != nil {
			if !errors.As(err, &ace) || ace.Type() != apstra.ErrNotfound {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch deployed configuration for node %s", config.NodeId), err.Error())
				return
			}
		}
		staged, err = o.client.GetNodeRenderedConfig(ctx, bpId, node, enum.RenderedConfigTypeStaging)
		if err != nil {
			if !errors.As(err, &ace) || ace.Type() != apstra.ErrNotfound {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch staged configuration for node %s", config.NodeId), err.Error())
				return
			}
		}
		diff, err := o.client.GetNodeRenderedConfigDiff(ctx, bpId, node)
		if err != nil {
			if !errors.As(err, &ace) || ace.Type() != apstra.ErrNotfound {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch incremental configuration for node %s", config.NodeId), err.Error())
				return
			}
		}
		incremental = diff.Config
	case !config.SystemId.IsNull():
		system := apstra.ObjectId(config.SystemId.ValueString())
		deployed, err = o.client.GetSystemRenderedConfig(ctx, bpId, system, enum.RenderedConfigTypeDeployed)
		if err != nil {
			if !errors.As(err, &ace) || ace.Type() != apstra.ErrNotfound {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch deployed configuration for system %s", config.SystemId), err.Error())
				return
			}
		}
		staged, err = o.client.GetSystemRenderedConfig(ctx, bpId, system, enum.RenderedConfigTypeStaging)
		if err != nil {
			if !errors.As(err, &ace) || ace.Type() != apstra.ErrNotfound {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch staged configuration for system %s", config.SystemId), err.Error())
				return
			}
		}
		diff, err := o.client.GetSystemRenderedConfigDiff(ctx, bpId, system)
		if err != nil {
			if !errors.As(err, &ace) || ace.Type() != apstra.ErrNotfound {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch incremental configuration for system %s", config.SystemId), err.Error())
				return
			}
		}
		incremental = diff.Config
	}

	config.DeployedCfg = utils.StringValueOrNull(ctx, deployed, &resp.Diagnostics)
	config.StagedCfg = utils.StringValueOrNull(ctx, staged, &resp.Diagnostics)
	config.Incremental = utils.StringValueOrNull(ctx, incremental, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceBlueprintNodeConfig) setClient(client *apstra.Client) {
	o.client = client
}
