package design

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
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

type PropertySet struct {
	Id         types.String `tfsdk:"id"`
	Label      types.String `tfsdk:"name"`
	Values     types.String `tfsdk:"keys"`
	Blueprints types.Set    `tfsdk:"blueprints"`
}

func (o PropertySet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Property Set by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up a property set by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"keys": dataSourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the property set",
			Computed:            true,
		},
		"blueprints": dataSourceSchema.SetAttribute{
			MarkdownDescription: "List of blueprints that this property set might be associated with.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o PropertySet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a PropertySet by ID. Required when `name` is omitted.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a property set by name. Required when `id` is omitted.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"keys": resourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the property set",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"blueprints": resourceSchema.SetAttribute{
			MarkdownDescription: "List of blueprints that this property set might be associated with.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o *PropertySet) LoadApiData(ctx context.Context, in *apstra.PropertySetData, diags *diag.Diagnostics) {
	o.Label = types.StringValue(in.Label)
	var d diag.Diagnostics
	o.Blueprints, d = types.SetValueFrom(ctx, types.StringType, in.Blueprints)
	diags.Append(d...)
	o.Values = types.StringValue(string(in.Values))
}

func (o *PropertySet) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.PropertySetData {
	if !utils.IsJSON(o.Values) {
		diags.AddError("error Marshalling to JSON", o.Values.ValueString())
	}
	return &apstra.PropertySetData{
		Label:  o.Label.ValueString(),
		Values: []byte(o.Values.ValueString()),
	}
}
