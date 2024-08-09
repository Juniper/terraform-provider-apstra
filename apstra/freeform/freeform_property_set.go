package freeform

import (
	"context"
	"encoding/json"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

type FreeformPropertySet struct {
	Id          types.String         `tfsdk:"id"`
	BlueprintId types.String         `tfsdk:"blueprint_id"`
	Name        types.String         `tfsdk:"name"`
	SystemId    types.String         `tfsdk:"system_id"`
	Values      jsontypes.Normalized `tfsdk:"values"`
}

func (o FreeformPropertySet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Property Set lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Freeform Property Set by ID. Required when `name` is omitted.",
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
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The system ID where the Property Set is associated.",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up an imported Property Set by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"values": dataSourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the Property Set in JSON format.",
			CustomType:          jsontypes.NormalizedType{},
			Computed:            true,
		},
	}
}

func (o FreeformPropertySet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Property Set.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"system_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The system ID where the Property Set is associated.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Property Set name as shown in the Web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"values": resourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the Property Set in JSON format.",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
	}
}

func (o *FreeformPropertySet) Request(_ context.Context, _ *diag.Diagnostics) *apstra.FreeformPropertySetData {
	return &apstra.FreeformPropertySetData{
		SystemId: (*apstra.ObjectId)(o.SystemId.ValueStringPointer()),
		Label:    o.Name.ValueString(),
		Values:   json.RawMessage(o.Values.ValueString()),
	}
}

func (o *FreeformPropertySet) LoadApiData(_ context.Context, in *apstra.FreeformPropertySetData, _ *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Values = jsontypes.NewNormalizedValue(string(in.Values))
	if in.SystemId != nil {
		o.SystemId = types.StringValue(string(*in.SystemId))
	}
}
