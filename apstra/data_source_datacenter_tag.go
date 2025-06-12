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

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterTag{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterTag{}
)

type dataSourceDatacenterTag struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterTag) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_tag"
}

func (o *dataSourceDatacenterTag) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterTag) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns details of a specific Tag within a Blueprint.",
		Attributes:          blueprint.Tag{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterTag) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config blueprint.Tag
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

	tags, err := bp.GetAllTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to get blueprint tags", err.Error())
		return
	}

	var dcTag *apstra.TwoStageL3ClosTag
	for _, v := range tags {
		if v.Data.Label == config.Name.ValueString() {
			dcTag = &v
			break
		}
	}
	if dcTag == nil {
		resp.Diagnostics.AddError(
			"Tag not found",
			fmt.Sprintf("No Tag with name %s found in Blueprint %s", config.Name, config.BlueprintId),
		)
		return
	}

	config.LoadApiData(ctx, dcTag.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterTag) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
