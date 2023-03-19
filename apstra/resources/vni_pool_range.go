package resources

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VniPoolRange struct {
	Status         types.String  `tfsdk:"status"`
	First          types.Int64   `tfsdk:"first"`
	Last           types.Int64   `tfsdk:"last"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

const (
	VniMin = 4096
	VniMax = 16777214
)

func (o VniPoolRange) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":          types.StringType,
		"first":           types.Int64Type,
		"last":            types.Int64Type,
		"total":           types.Int64Type,
		"used":            types.Int64Type,
		"used_percentage": types.Float64Type,
	}
}

func (o VniPoolRange) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"first": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Lowest numbered ID in this VNI Pool Range.",
			Computed:            true,
		},
		"last": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Highest numbered ID in this VNI Pool Range.",
			Computed:            true,
		},
		"total": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of IDs in the VNI Pool Range.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the VNI Pool Range, as reported by Apstra.",
			Computed:            true,
		},
		"used": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used VNIs in the VNI Pool Range.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used IDs in the VNI Pool Range.",
			Computed:            true,
		},
	}
}

func (o VniPoolRange) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"first": resourceSchema.Int64Attribute{
			Required:   true,
			Validators: []validator.Int64{int64validator.Between(VniMin-1, VniMax+1)},
		},
		"last": resourceSchema.Int64Attribute{
			Required: true,
			Validators: []validator.Int64{
				int64validator.Between(VniMin-1, VniMax+1),
				int64validator.AtLeastSumOf(path.MatchRelative().AtParent().AtName("first")),
			},
		},
		"total": resourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of IDs in the VNI Pool Range.",
			Computed:            true,
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Status of the VNI Pool Range, as reported by Apstra." +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
		"used": resourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used IDs in the VNI Pool Range." +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
		"used_percentage": resourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used IDs in the VNI Pool Range." +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
	}
}

func (o *VniPoolRange) LoadApiData(_ context.Context, in *goapstra.IntRange, _ *diag.Diagnostics) {
	o.Status = types.StringValue(in.Status)
	o.First = types.Int64Value(int64(in.First))
	o.Last = types.Int64Value(int64(in.Last))
	o.Total = types.Int64Value(int64(in.Total))
	o.Used = types.Int64Value(int64(in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
}

func (o *VniPoolRange) Request(_ context.Context, _ *diag.Diagnostics) goapstra.IntfIntRange {
	return &goapstra.IntRangeRequest{
		First: uint32(o.First.ValueInt64()),
		Last:  uint32(o.Last.ValueInt64()),
	}
}
