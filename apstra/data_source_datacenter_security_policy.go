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
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterSecurityPolicy{}

type dataSourceDatacenterSecurityPolicy struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterSecurityPolicy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_security_policy"
}

func (o *dataSourceDatacenterSecurityPolicy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.getBpClientFunc = DataSourceGetTwoStageL3ClosClientFunc(ctx, req, resp)
}

func (o *dataSourceDatacenterSecurityPolicy) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source provides details of a specific Security " +
			"Policy within a Datacenter Blueprint.\n\nAt least one optional attribute is required.",
		Attributes: blueprint.DatacenterSecurityPolicy{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterSecurityPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.DatacenterSecurityPolicy
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

	err = config.Read(ctx, bp, &resp.Diagnostics)
	if err != nil {
		if utils.IsApstra404(err) {
			switch config.Id.IsNull() {
			case true:
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"Security Policy not found",
					fmt.Sprintf("Blueprint %q Security Policy with Name %s not found", bp.Id(), config.Name))
			case false:
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"Security Policy not found",
					fmt.Sprintf("Blueprint %q Security Policy with ID %s not found", bp.Id(), config.Id))
			}
			return
		}
		resp.Diagnostics.AddError("Failed to read Security Policy", err.Error())
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
