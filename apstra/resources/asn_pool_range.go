package resources

import (
	"context"
	"math"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstraplanmodifier "github.com/Juniper/terraform-provider-apstra/apstra/plan_modifier"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	minAsn = 1              // rfc4893 says 0 is okay, but apstra says "Must be between 1 and 4294967295"
	maxAsn = math.MaxUint32 // 4294967295 rfc4893
)

type AsnPoolRange struct {
	Status         types.String  `tfsdk:"status"`
	First          types.Int64   `tfsdk:"first"`
	Last           types.Int64   `tfsdk:"last"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

func (o AsnPoolRange) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"status":          types.StringType,
		"first":           types.Int64Type,
		"last":            types.Int64Type,
		"total":           types.Int64Type,
		"used":            types.Int64Type,
		"used_percentage": types.Float64Type,
	}
}

func (o AsnPoolRange) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"first": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Lowest numbered AS in this ASN Pool Range.",
			Computed:            true,
		},
		"last": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Highest numbered AS in this ASN Pool Range.",
			Computed:            true,
		},
		"total": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of ASNs in the ASN Pool Range.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the ASN Pool Range, as reported by Apstra.",
			Computed:            true,
		},
		"used": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used ASNs in the ASN Pool Range.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used ASNs in the ASN Pool Range.",
			Computed:            true,
		},
	}
}

func (o AsnPoolRange) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"first": resourceSchema.Int64Attribute{
			Required:   true,
			Validators: []validator.Int64{int64validator.Between(minAsn, maxAsn)},
		},
		"last": resourceSchema.Int64Attribute{
			Required: true,
			Validators: []validator.Int64{
				int64validator.Between(minAsn, maxAsn),
				int64validator.AtLeastSumOf(path.MatchRelative().AtParent().AtName("first")),
			},
		},
		"total": resourceSchema.Int64Attribute{
			MarkdownDescription: "Mutable read-only attribute is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{apstraplanmodifier.UseNullStateForUnknown()},
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Mutable read-only attribute is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{apstraplanmodifier.UseNullStateForUnknown()},
		},
		"used": resourceSchema.Int64Attribute{
			MarkdownDescription: "Mutable read-only attribute is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{apstraplanmodifier.UseNullStateForUnknown()},
		},
		"used_percentage": resourceSchema.Float64Attribute{
			MarkdownDescription: "Mutable read-only attribute is always null in a Resource. Use the matching Data Source for this information.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Float64{apstraplanmodifier.UseNullStateForUnknown()},
		},
	}
}

func (o *AsnPoolRange) LoadApiData(_ context.Context, in *apstra.IntRange, _ *diag.Diagnostics) {
	o.Status = types.StringValue(in.Status)
	o.First = types.Int64Value(int64(in.First))
	o.Last = types.Int64Value(int64(in.Last))
	o.Total = types.Int64Value(int64(in.Total))
	o.Used = types.Int64Value(int64(in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
}

func (o *AsnPoolRange) Request(_ context.Context, _ *diag.Diagnostics) apstra.IntfIntRange {
	return &apstra.IntRangeRequest{
		First: uint32(o.First.ValueInt64()),
		Last:  uint32(o.Last.ValueInt64()),
	}
}

func (o *AsnPoolRange) setMutablesToNull() {
	o.Status = types.StringNull()
	o.Total = types.Int64Null()
	o.Used = types.Int64Null()
	o.UsedPercentage = types.Float64Null()
}
