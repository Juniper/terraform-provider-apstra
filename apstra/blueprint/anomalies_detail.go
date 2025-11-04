package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type anomalyDetail struct {
	AnomalyId   types.String `tfsdk:"anomaly_id"`
	Severity    types.String `tfsdk:"severity"`
	AnomalyType types.String `tfsdk:"type"`
	Expected    types.String `tfsdk:"expected"`
	Actual      types.String `tfsdk:"actual"`
	Identity    types.String `tfsdk:"identity"`
	Role        types.String `tfsdk:"role"`
	Anomalous   types.String `tfsdk:"anomalous"`
}

func (o anomalyDetail) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"anomaly_id": types.StringType,
		"severity":   types.StringType,
		"type":       types.StringType,
		"expected":   types.StringType,
		"actual":     types.StringType,
		"identity":   types.StringType,
		"role":       types.StringType,
		"anomalous":  types.StringType,
	}
}

func (o anomalyDetail) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"anomaly_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Anomaly ID.",
			Computed:            true,
		},
		"severity": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Severity of Anomaly.",
			Computed:            true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Anomaly Type.",
			Computed:            true,
		},
		"expected": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Extended Anomaly attribute describing the expected value/state/condition in JSON format.",
			Computed:            true,
		},
		"actual": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Extended Anomaly attribute describing the actual value/state/condition in JSON format.",
			Computed:            true,
		},
		"identity": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Extended Anomaly attribute which identifies the anomalous value/state/condition in JSON format.",
			Computed:            true,
		},
		"role": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Anomaly role further contextualizes `type`.",
			Computed:            true,
		},
		"anomalous": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Extended Anomaly attribute which further contextualizes the Anomaly.",
			Computed:            true,
		},
	}
}

func (o *anomalyDetail) loadApiData(ctx context.Context, in *apstra.BlueprintAnomaly, diags *diag.Diagnostics) {
	var role string
	if in.Role != nil {
		role = *in.Role
	}

	o.AnomalyId = types.StringValue(in.Id.String())
	o.Severity = types.StringValue(in.Severity)
	o.AnomalyType = types.StringValue(in.AnomalyType)
	o.Expected = value.StringOrNull(ctx, string(in.Expected), diags)
	o.Actual = value.StringOrNull(ctx, string(in.Actual), diags)
	o.Identity = value.StringOrNull(ctx, string(in.Identity), diags)
	o.Role = value.StringOrNull(ctx, role, diags)
	o.Anomalous = value.StringOrNull(ctx, string(in.Anomalous), diags)
}

func newAnomalyDetailSet(ctx context.Context, in []apstra.BlueprintAnomaly, diags *diag.Diagnostics) types.Set {
	anomalyDetails := make([]anomalyDetail, len(in))
	for i, anomalyDetail := range in {
		anomalyDetails[i].loadApiData(ctx, &anomalyDetail, diags)
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: anomalyDetail{}.attrTypes()})
	}

	return value.SetOrNull(ctx, types.ObjectType{AttrTypes: anomalyDetail{}.attrTypes()}, anomalyDetails, diags)
}
