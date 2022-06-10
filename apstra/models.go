package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AsnPool struct {
	Id          types.String   `tfsdk:"id"`
	DisplayName types.String   `tfsdk:"display_name"`
	Status      types.String   `tfsdk:"status"`
	Tags        []types.String `tfsdk:"tags"`
	//Used        types.Int64    `tfsdk:"used"`
	//UsedPercent types.Float64  `tfsdk:"used_percentage"`
	//Created     types.String   `tfsdk:"created"`
	//Modified    types.String   `tfsdk:"modified"`
	//Ranges      []AsnRange     `tfsdk:"ranges"`
}

type AsnRange struct {
	Status         types.String  `tfsdk:"status"`
	First          types.Int64   `tfsdk:"first"`
	Last           types.Int64   `tfsdk:"last"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}
