package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
)

type DatacenterPropertySet struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Id          types.String `tfsdk:"id"`
	Label       types.String `tfsdk:"name"`
	Values      types.String `tfsdk:"data"`
	Stale       types.Bool   `tfsdk:"stale"`
	//	Keys        types.SetType `tfsdk:"keys"`
}

func (o DatacenterPropertySet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the blueprint that the property set has been imported into.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up an imported Property Set by ID. Required when `name` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up an imported Property Set by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"data": dataSourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the Property Set in JSON format",
			Computed:            true,
		},
		"stale": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "This is true if the imported Property Set does not match the global property set",
			Computed:            true,
		},
	}
}

func (o DatacenterPropertySet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the blueprint that the property set is imported into.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up an imported Property Set by ID. Required when `name` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up an imported Property Set by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"keys": resourceSchema.SetAttribute{
			MarkdownDescription: "Subset of Keys to import. Empty set implies all keys imported.",
			Required:            true,
			ElementType:         types.StringType,
			PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		},
		"data": resourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the Property Set in JSON format",
			Computed:            true,
			Validators:          []validator.String{apstravalidator.ParseJson()},
		},
		"stale": resourceSchema.BoolAttribute{
			MarkdownDescription: "This is true if the Property Set does not match the global property set",
			Computed:            true,
		},
	}
}

func (o *DatacenterPropertySet) LoadApiData(ctx context.Context, in *apstra.TwoStageL3ClosPropertySet, diags *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Label = types.StringValue(in.Label)
	var d diag.Diagnostics
	diags.Append(d...)
	o.Values = types.StringValue(string(in.Values))
	o.Stale = types.BoolValue(in.Stale)
}

func (o *DatacenterPropertySet) Request(_ context.Context, _ *diag.Diagnostics) *apstra.TwoStageL3ClosPropertySet {
	return &apstra.TwoStageL3ClosPropertySet{
		Label:  o.Label.ValueString(),
		Values: []byte(o.Values.ValueString()),
	}
}

func (o *DatacenterPropertySet) ImportRequest(_ context.Context, _ *diag.Diagnostics) *apstra.TwoStageL3ClosPropertySet {
	return &apstra.TwoStageL3ClosPropertySet{
		Label:  o.Label.ValueString(),
		Values: []byte(o.Values.ValueString()),
	}
}
