package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

const (
	minimumFreeFormVersion        = "4.1.1"
	twoStageL3ClosRefDesignUiName = "datacenter"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprintIds{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceBlueprintIds{}

type dataSourceBlueprintIds struct {
	client *goapstra.Client
}

func (o *dataSourceBlueprintIds) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_ids"
}

func (o *dataSourceBlueprintIds) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceBlueprintIds) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source returns a list of blueprint IDs configured on Apstra.",
		Attributes: map[string]tfsdk.Attribute{
			"ids": {
				MarkdownDescription: "ID of the desired ASN Resource Pool.",
				Computed:            true,
				Type:                types.ListType{ElemType: types.StringType},
			},
			"reference_design": {
				MarkdownDescription: "Optional filter for bluepirnts of the specified reference design.",
				Optional:            true,
				Type:                types.StringType,
				Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(
					twoStageL3ClosRefDesignUiName,
					goapstra.RefDesignFreeform.String(),
				)},
			},
		},
	}, nil
}

func (o *dataSourceBlueprintIds) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if o.client == nil {
		return
	}

	var config dBlueprintIds
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.RefDesign.IsNull() && config.RefDesign.Value == goapstra.RefDesignFreeform.String() {
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

func (o *dataSourceBlueprintIds) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		resp.Diagnostics.AddWarning("got some blueprints", strconv.Itoa(len(objectIds)))
	} else {
		var refDesign string
		// substitute UI name for API name
		switch config.RefDesign.Value {
		case twoStageL3ClosRefDesignUiName:
			refDesign = goapstra.RefDesignDatacenter.String()
		default:
			refDesign = config.RefDesign.Value
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

	elems := make([]attr.Value, len(objectIds))
	for i, id := range objectIds {
		elems[i] = types.String{Value: string(id)}
	}

	// Set state
	diags = resp.State.Set(ctx, &dBlueprintIds{
		RefDesign: config.RefDesign,
		Ids:       types.List{ElemType: types.StringType, Elems: elems},
	})
	resp.Diagnostics.Append(diags...)
}

type dBlueprintIds struct {
	Ids       types.List   `tfsdk:"ids"`
	RefDesign types.String `tfsdk:"reference_design"`
}
