package resources

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstraplanmodifier "github.com/Juniper/terraform-provider-apstra/apstra/plan_modifier"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AsnPool struct {
	Id             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Ranges         types.Set     `tfsdk:"ranges"`
	Total          types.Int64   `tfsdk:"total"`
	Status         types.String  `tfsdk:"status"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}

func (o AsnPool) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the desired ASN Pool. Required when `name` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Display name of the ASN Pool. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ranges": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Detailed info about individual ASN Pool Ranges within the ASN Pool.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: AsnPoolRange{}.DataSourceAttributes(),
			},
		},
		"total": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of ASNs in the ASN Pool.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the ASN Pool.",
			Computed:            true,
		},
		"used": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used ASNs in the ASN Pool.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used ASNs in the ASN Pool.",
			Computed:            true,
		},
	}
}

func (o AsnPool) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID number of the pool",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Pool name displayed in the Apstra web UI",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ranges": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Ranges mark the begin/end AS numbers available from the pool",
			Required:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: AsnPoolRange{}.ResourceAttributes(),
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

func (o *AsnPool) LoadApiData(ctx context.Context, in *apstra.AsnPool, diags *diag.Diagnostics) {
	ranges := make([]AsnPoolRange, len(in.Ranges))
	for i, r := range in.Ranges {
		ranges[i].LoadApiData(ctx, &r, diags)
		if diags.HasError() {
			return
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.DisplayName)
	o.Status = types.StringValue(in.Status.String())
	o.Used = types.Int64Value(int64(in.Used))
	o.UsedPercentage = types.Float64Value(float64(in.UsedPercentage))
	o.Total = types.Int64Value(int64(in.Total))
	o.Ranges = utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: AsnPoolRange{}.AttrTypes()}, ranges, diags)
}

func (o *AsnPool) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.AsnPoolRequest {
	response := apstra.AsnPoolRequest{
		DisplayName: o.Name.ValueString(),
		Ranges:      make([]apstra.IntfIntRange, len(o.Ranges.Elements())),
	}

	poolRanges := make([]AsnPoolRange, len(o.Ranges.Elements()))
	d := o.Ranges.ElementsAs(ctx, &poolRanges, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	for i, poolRange := range poolRanges {
		response.Ranges[i] = poolRange.Request(ctx, diags)
	}

	return &response
}

func (o *AsnPool) SetMutablesToNull(ctx context.Context, diags *diag.Diagnostics) {
	o.Status = types.StringNull()
	o.Total = types.Int64Null()
	o.Used = types.Int64Null()
	o.UsedPercentage = types.Float64Null()

	var ranges []AsnPoolRange
	diags.Append(o.Ranges.ElementsAs(ctx, &ranges, false)...)
	if diags.HasError() {
		return
	}

	for i := range ranges {
		ranges[i].setMutablesToNull()
	}

	o.Ranges = utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: AsnPoolRange{}.AttrTypes()}, ranges, diags)
}
