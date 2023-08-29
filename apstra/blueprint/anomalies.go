package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Anomalies struct {
	BlueprintId      types.String `tfsdk:"blueprint_id"`
	Details          types.Set    `tfsdk:"details"`
	SummaryByNode    types.Set    `tfsdk:"summary_by_node"`
	SummaryByService types.Set    `tfsdk:"summary_by_service"`
}

func (o Anomalies) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Apstra Blueprint ID.",
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"details": dataSourceSchema.SetNestedAttribute{
			NestedObject:        dataSourceSchema.NestedAttributeObject{Attributes: anomalyDetail{}.dataSourceAttributes()},
			Computed:            true,
			MarkdownDescription: "Each current Anomaly is represented by an object in this set.",
		},
		"summary_by_node": dataSourceSchema.SetNestedAttribute{
			NestedObject:        dataSourceSchema.NestedAttributeObject{Attributes: anomalyNodeSummary{}.dataSourceAttributes()},
			Computed:            true,
			MarkdownDescription: "Set of Anomaly summaries organized by Node.",
		},
		"summary_by_service": dataSourceSchema.SetNestedAttribute{
			NestedObject:        dataSourceSchema.NestedAttributeObject{Attributes: anomalyServiceSummary{}.dataSourceAttributes()},
			Computed:            true,
			MarkdownDescription: "Set of Anomaly summaries organized by Fabric Service.",
		},
	}
}

func (o *Anomalies) ReadFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	o.ReadAnomalyDetailsFromApi(ctx, client, diags)
	if diags.HasError() {
		return
	}

	o.ReadAnomalyNodeSummariesFromApi(ctx, client, diags)
	if diags.HasError() {
		return
	}

	o.ReadAnomalyServiceSummariesFromApi(ctx, client, diags)
	if diags.HasError() {
		return
	}
}

func (o *Anomalies) ReadAnomalyDetailsFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	anomalies, err := client.GetBlueprintAnomalies(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			diags.AddAttributeError(
				path.Root("blueprint_id"),
				"not found",
				fmt.Sprintf("Blueprint %s not found", o.BlueprintId),
			)
			return
		}
		diags.AddError(fmt.Sprintf("unable to fetch Blueprint %s Anomalies from API", o.BlueprintId), err.Error())
		return
	}

	o.Details = newAnomalyDetailSet(ctx, anomalies, diags)
}

func (o *Anomalies) ReadAnomalyNodeSummariesFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	nodeSummaries, err := client.GetBlueprintNodeAnomalyCounts(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			diags.AddAttributeError(
				path.Root("blueprint_id"),
				"not found",
				fmt.Sprintf("Blueprint %s not found", o.BlueprintId),
			)
			return
		}
		diags.AddError(fmt.Sprintf("unable to fetch Blueprint %s per-Node Anomaly Summaries from API", o.BlueprintId), err.Error())
		return
	}

	o.SummaryByNode = newAnomalyNodeSummarySet(ctx, nodeSummaries, diags)
}

func (o *Anomalies) ReadAnomalyServiceSummariesFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	serviceSummaries, err := client.GetBlueprintServiceAnomalyCounts(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			diags.AddAttributeError(
				path.Root("blueprint_id"),
				"not found",
				fmt.Sprintf("Blueprint %s not found", o.BlueprintId),
			)
			return
		}
		diags.AddError(fmt.Sprintf("unable to fetch Blueprint %s per-Service Anomaly Summaries from API", o.BlueprintId), err.Error())
		return
	}

	o.SummaryByService = newAnomalyServiceSummarySet(ctx, serviceSummaries, diags)
}
