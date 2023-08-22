package resources

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
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
	"terraform-provider-apstra/apstra/utils"
)

type IntegerPool struct {
	Id          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Ranges      types.Set     `tfsdk:"ranges"`
	Total       types.Int64   `tfsdk:"total"`
	Status      types.String  `tfsdk:"status"`
	Used        types.Int64   `tfsdk:"used"`
	UsedPercent types.Float64 `tfsdk:"used_percentage"`
}

func (o IntegerPool) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Integer Pool. Required when `name` is omitted.",
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
			MarkdownDescription: "Web UI name of the Integer Pool. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ranges": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Detailed info about individual Integer Pool Ranges within the Integer Pool.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: IntegerPoolRange{}.DataSourceAttributes(),
			},
		},
		"total": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of Integers in the Integer Pool.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the Integer Pool.",
			Computed:            true,
		},
		"used": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used Integers in the Integer Pool.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used Integers in the Integer Pool.",
			Computed:            true,
		},
	}
}

func (o IntegerPool) ResourceAttributes() map[string]resourceSchema.Attribute {
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
			MarkdownDescription: "Ranges mark the begin/end Integers available from the pool",
			Required:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: IntegerPoolRange{}.ResourceAttributes(),
			},
		},
		"total": resourceSchema.Int64Attribute{
			MarkdownDescription: "Total number of Integers in the Integer Pool.",
			Computed:            true,
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Status of the Integer Pool. " +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
		"used": resourceSchema.Int64Attribute{
			MarkdownDescription: "Count of used Integers in the Integer Pool. " +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
		"used_percentage": resourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used Integers in the Integer Pool. " +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
	}
}

func (o *IntegerPool) LoadApiData(ctx context.Context, in *apstra.IntPool, diags *diag.Diagnostics) {
	ranges := make([]IntegerPoolRange, len(in.Ranges))
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
	o.UsedPercent = types.Float64Value(float64(in.UsedPercentage))
	o.Total = types.Int64Value(int64(in.Total))
	o.Ranges = utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: IntegerPoolRange{}.AttrTypes()}, ranges, diags)
}

func (o *IntegerPool) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.IntPoolRequest {
	response := apstra.IntPoolRequest{
		DisplayName: o.Name.ValueString(),
		Ranges:      make([]apstra.IntfIntRange, len(o.Ranges.Elements())),
	}

	poolRanges := make([]IntegerPoolRange, len(o.Ranges.Elements()))
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
