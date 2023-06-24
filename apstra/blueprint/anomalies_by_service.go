package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type anomalyServiceSummary struct {
	AnomalyType types.String `tfsdk:"type"`
	Role        types.String `tfsdk:"role"`
	Count       types.Int64  `tfsdk:"count"`
}

func (o anomalyServiceSummary) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":  types.StringType,
		"role":  types.StringType,
		"count": types.Int64Type,
	}
}

func (o anomalyServiceSummary) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Fabric Service experiencing Anomalies.",
			Computed:            true,
		},
		"role": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Further context about the Fabric Service Anomalies.",
			Computed:            true,
		},
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of Anomalies related to the Fabric Service and Role.",
			Computed:            true,
		},
	}
}

func (o *anomalyServiceSummary) loadApiData(_ context.Context, in *apstra.BlueprintServiceAnomalyCount, _ *diag.Diagnostics) {
	o.AnomalyType = types.StringValue(in.AnomalyType)
	o.Role = types.StringValue(in.Role)
	o.Count = types.Int64Value(int64(in.Count))
}

func newAnomalyServiceSummarySet(ctx context.Context, in []apstra.BlueprintServiceAnomalyCount, diags *diag.Diagnostics) types.Set {
	serviceSummaries := make([]anomalyServiceSummary, len(in))
	for i, serviceSummary := range in {
		serviceSummaries[i].loadApiData(ctx, &serviceSummary, diags)
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: anomalyServiceSummary{}.attrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: anomalyServiceSummary{}.attrTypes()}, serviceSummaries, diags)
}
