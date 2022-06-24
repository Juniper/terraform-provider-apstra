package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceAgentProfile struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	//Packages []types.String `tfsdk:"packages"`
	Platform types.String `tfsdk:"platform"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	//OpenOptions types.Map      `tfsdk:"open_options"`
}

type ResourceAsnPool struct {
	Id          types.String   `tfsdk:"id"`
	DisplayName types.String   `tfsdk:"display_name"`
	Tags        []types.String `tfsdk:"tags"`
}

type ResourceAsnPoolRange struct {
	PoolId types.String `tfsdk:"pool_id"`
	First  types.Int64  `tfsdk:"first"`
	Last   types.Int64  `tfsdk:"last"`
}

type DataAsnPool struct {
	Id             types.String   `tfsdk:"id"`
	DisplayName    types.String   `tfsdk:"display_name"`
	Status         types.String   `tfsdk:"status"`
	Tags           []types.String `tfsdk:"tags"`
	Used           types.Int64    `tfsdk:"used"`
	UsedPercent    types.Float64  `tfsdk:"used_percentage"`
	CreatedAt      types.String   `tfsdk:"created_at"`
	LastModifiedAt types.String   `tfsdk:"last_modified_at"`
	Total          types.Int64    `tfsdk:"total"`
	Ranges         []AsnRange     `tfsdk:"ranges"`
}

type AsnRange struct {
	Status         types.String  `tfsdk:"status"`
	First          types.Int64   `tfsdk:"first"`
	Last           types.Int64   `tfsdk:"last"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

type DataAsnPoolIds struct {
	Ids []types.String `tfsdk:"ids"`
}

type DataAsnPoolId struct {
	Id          types.String   `tfsdk:"id"`
	DisplayName types.String   `tfsdk:"display_name"`
	Tags        []types.String `tfsdk:"tags""`
}

type DataAgentProfileIds struct {
	Ids []types.String `tfsdk:"ids"`
}

type DataAgentProfileId struct {
	Id    types.String   `tfsdk:"id"`
	Label types.String   `tfsdk:"label"`
	Tags  []types.String `tfsdk:"tags""`
}
