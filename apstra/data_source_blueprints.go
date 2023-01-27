package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	minimumFreeFormVersion        = "4.1.1"
	twoStageL3ClosRefDesignUiName = "datacenter"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprints{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceBlueprints{}

type dataSourceBlueprints struct {
	client *goapstra.Client
}

func (o *dataSourceBlueprints) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprints"
}

func (o *dataSourceBlueprints) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataDetail,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *dataSourceBlueprints) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns a list of blueprint IDs configured on Apstra.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "ID of the desired ASN Resource Pool.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"reference_design": schema.StringAttribute{
				MarkdownDescription: "Optional filter for blueprints of the specified reference design.",
				Optional:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					twoStageL3ClosRefDesignUiName,
					goapstra.RefDesignFreeform.String(),
				)},
			},
		},
	}
}

func (o *dataSourceBlueprints) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if o.client == nil {
		return
	}

	var config dBlueprintIds
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.RefDesign.IsNull() && config.RefDesign.ValueString() == goapstra.RefDesignFreeform.String() {
		minVer, err := version.NewVersion(minimumFreeFormVersion)
		if err != nil {
			resp.Diagnostics.AddError("error parsing minimum freeform version", err.Error())
		}

		thisVer, err := version.NewVersion(o.client.ApiVersion())
		if err != nil {
			resp.Diagnostics.AddError("error parsing reported apstra version", err.Error())
		}

		if thisVer.LessThan(minVer) {
			resp.Diagnostics.AddError("Apstra API version error",
				fmt.Sprintf("Apstra %s doesn't support reference design '%s'",
					o.client.ApiVersion(), goapstra.RefDesignFreeform.String()))
		}
	}
}

func (o *dataSourceBlueprints) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dBlueprintIds
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var objectIds []goapstra.ObjectId
	var err error
	if config.RefDesign.IsNull() {
		objectIds, err = o.client.ListAllBlueprintIds(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error listing blueprint IDs", err.Error())
			return
		}
	} else {
		var refDesign string
		// substitute UI name for API name
		switch config.RefDesign.ValueString() {
		case twoStageL3ClosRefDesignUiName:
			refDesign = goapstra.RefDesignDatacenter.String()
		default:
			refDesign = config.RefDesign.ValueString()
		}

		bpStatuses, err := o.client.GetAllBlueprintStatus(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving blueprint statuses", err.Error())
			return
		}
		for _, bpStatus := range bpStatuses {
			if bpStatus.Design.String() == refDesign {
				objectIds = append(objectIds, bpStatus.Id)
			}
		}
	}

	ids := make([]attr.Value, len(objectIds))
	for i, id := range objectIds {
		ids[i] = types.StringValue(string(id))
	}

	idSet, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &dBlueprintIds{
		RefDesign: config.RefDesign,
		Ids:       idSet,
	})
	resp.Diagnostics.Append(diags...)
}

type dBlueprintIds struct {
	Ids       types.Set    `tfsdk:"ids"`
	RefDesign types.String `tfsdk:"reference_design"`
}
