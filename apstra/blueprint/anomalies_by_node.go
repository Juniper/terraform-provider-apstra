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

type anomalyNodeSummary struct {
	NodeName           types.String `tfsdk:"node_name"`
	SystemId           types.String `tfsdk:"system_id"`
	Total              types.Int64  `tfsdk:"total"`
	Arp                types.Int64  `tfsdk:"arp"`
	Bgp                types.Int64  `tfsdk:"bgp"`
	BlueprintRendering types.Int64  `tfsdk:"blueprint_rendering"`
	Cabling            types.Int64  `tfsdk:"cabling"`
	Config             types.Int64  `tfsdk:"config"`
	Counter            types.Int64  `tfsdk:"counter"`
	Deployment         types.Int64  `tfsdk:"deployment"`
	Hostname           types.Int64  `tfsdk:"hostname"`
	Interface          types.Int64  `tfsdk:"interface"`
	Lag                types.Int64  `tfsdk:"lag"`
	Liveness           types.Int64  `tfsdk:"liveness"`
	Mac                types.Int64  `tfsdk:"mac"`
	Mlag               types.Int64  `tfsdk:"mlag"`
	Probe              types.Int64  `tfsdk:"probe"`
	Route              types.Int64  `tfsdk:"route"`
	Series             types.Int64  `tfsdk:"series"`
	Streaming          types.Int64  `tfsdk:"streaming"`
}

func (o anomalyNodeSummary) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"node_name":           types.StringType,
		"system_id":           types.StringType,
		"total":               types.Int64Type,
		"arp":                 types.Int64Type,
		"bgp":                 types.Int64Type,
		"blueprint_rendering": types.Int64Type,
		"cabling":             types.Int64Type,
		"config":              types.Int64Type,
		"counter":             types.Int64Type,
		"deployment":          types.Int64Type,
		"hostname":            types.Int64Type,
		"interface":           types.Int64Type,
		"lag":                 types.Int64Type,
		"liveness":            types.Int64Type,
		"mac":                 types.Int64Type,
		"mlag":                types.Int64Type,
		"probe":               types.Int64Type,
		"route":               types.Int64Type,
		"series":              types.Int64Type,
		"streaming":           types.Int64Type,
	}
}

func (o anomalyNodeSummary) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"node_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Node experiencing Anomalies.",
			Computed:            true,
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "System ID of the Node experiencing Anomalies.",
			Computed:            true,
		},
		"total": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of Anomalies related to the Node.",
			Computed:            true,
		},
		"arp": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of ARP Anomalies related to the Node.",
			Computed:            true,
		},
		"bgp": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of BGP Anomalies related to the Node.",
			Computed:            true,
		},
		"blueprint_rendering": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Blueprint Rendering Anomalies related to the Node.",
			Computed:            true,
		},
		"cabling": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Cabling Anomalies related to the Node.",
			Computed:            true,
		},
		"config": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Config Anomalies related to the Node.",
			Computed:            true,
		},
		"counter": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Counter Anomalies related to the Node.",
			Computed:            true,
		},
		"deployment": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Deployment Anomalies related to the Node.",
			Computed:            true,
		},
		"hostname": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Hostname Anomalies related to the Node.",
			Computed:            true,
		},
		"interface": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Interface Anomalies related to the Node.",
			Computed:            true,
		},
		"lag": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of LAG Anomalies related to the Node.",
			Computed:            true,
		},
		"liveness": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Liveness Anomalies related to the Node.",
			Computed:            true,
		},
		"mac": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of MAC Anomalies related to the Node.",
			Computed:            true,
		},
		"mlag": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of MLAG Anomalies related to the Node.",
			Computed:            true,
		},
		"probe": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Probe Anomalies related to the Node.",
			Computed:            true,
		},
		"route": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Route Anomalies related to the Node.",
			Computed:            true,
		},
		"series": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Series Anomalies related to the Node.",
			Computed:            true,
		},
		"streaming": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Streaming Anomalies related to the Node.",
			Computed:            true,
		},
	}
}

func (o *anomalyNodeSummary) loadApiData(_ context.Context, in *apstra.BlueprintNodeAnomalyCounts, _ *diag.Diagnostics) {
	o.NodeName = types.StringValue(in.Node)
	o.SystemId = types.StringValue(in.SystemId.String())
	o.Total = types.Int64Value(int64(in.All))
	o.Arp = types.Int64Value(int64(in.Arp))
	o.Bgp = types.Int64Value(int64(in.Bgp))
	o.BlueprintRendering = types.Int64Value(int64(in.BlueprintRendering))
	o.Cabling = types.Int64Value(int64(in.Cabling))
	o.Config = types.Int64Value(int64(in.Config))
	o.Counter = types.Int64Value(int64(in.Counter))
	o.Deployment = types.Int64Value(int64(in.Deployment))
	o.Hostname = types.Int64Value(int64(in.Hostname))
	o.Interface = types.Int64Value(int64(in.Interface))
	o.Lag = types.Int64Value(int64(in.Lag))
	o.Liveness = types.Int64Value(int64(in.Liveness))
	o.Mac = types.Int64Value(int64(in.Mac))
	o.Mlag = types.Int64Value(int64(in.Mlag))
	o.Probe = types.Int64Value(int64(in.Probe))
	o.Route = types.Int64Value(int64(in.Route))
	o.Series = types.Int64Value(int64(in.Series))
	o.Streaming = types.Int64Value(int64(in.Streaming))
}

func newAnomalyNodeSummarySet(ctx context.Context, in []apstra.BlueprintNodeAnomalyCounts, diags *diag.Diagnostics) types.Set {
	nodeSummaries := make([]anomalyNodeSummary, len(in))
	for i, nodeSummary := range in {
		nodeSummaries[i].loadApiData(ctx, &nodeSummary, diags)
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: anomalyNodeSummary{}.attrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: anomalyNodeSummary{}.attrTypes()}, nodeSummaries, diags)
}
