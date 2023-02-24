package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ip4PoolSubnet struct {
	Status         types.String  `tfsdk:"status"`
	Network        types.String  `tfsdk:"network"`
	Total          types.Number  `tfsdk:"total"`
	Used           types.Number  `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

func (o ip4PoolSubnet) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":          types.StringType,
		"network":         types.StringType,
		"total":           types.NumberType,
		"used":            types.NumberType,
		"used_percentage": types.Float64Type,
	}
}

func (o *ip4PoolSubnet) loadApiData(_ context.Context, in *goapstra.IpSubnet, _ *diag.Diagnostics) {
	o.Status = types.StringValue(in.Status)
	o.Network = types.StringValue(in.Network.String())
	o.Total = types.NumberValue(bigIntToBigFloat(&in.Total))
	o.Used = types.NumberValue(bigIntToBigFloat(&in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
}
