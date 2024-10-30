package freeform

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
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

type GroupGenerator struct {
	Id          types.String `tfsdk:"id"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
	GroupId     types.String `tfsdk:"group_id"`
	Name        types.String `tfsdk:"name"`
	Scope       types.String `tfsdk:"scope"`
}

func (o GroupGenerator) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Group Generator by ID. Required when `name` is omitted.",
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
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Group lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"group_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Resource Group the Group Generator belongs to.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up Group Generator by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"scope": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Scope is a graph query which selects target nodes for which Groups should be generated.\n" +
				"Example: `node('system', name='target', label=aeq('*prod*'))`",
			Computed: true,
		},
	}
}

func (o GroupGenerator) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Group Generator within the Freeform Blueprint.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"group_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Resource Group the Group Generator belongs to. Omit to create at the `root` level.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform Group Generator name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(apstraregexp.StdNameConstraint, apstraregexp.StdNameConstraintMsg),
			},
		},
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope is a graph query which selects target nodes for which Group Generators should " +
				"be generated.\nExample: `node('system', name='target', label=aeq('*prod*'))`",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o *GroupGenerator) Request(_ context.Context, _ *diag.Diagnostics) *apstra.FreeformGroupGeneratorData {
	return &apstra.FreeformGroupGeneratorData{
		ParentId: (*apstra.ObjectId)(o.GroupId.ValueStringPointer()),
		Label:    o.Name.ValueString(),
		Scope:    o.Scope.ValueString(),
	}
}

func (o *GroupGenerator) LoadApiData(_ context.Context, in *apstra.FreeformGroupGeneratorData, _ *diag.Diagnostics) {
	o.GroupId = types.StringPointerValue((*string)(in.ParentId))
	o.Name = types.StringValue(in.Label)
	o.Scope = types.StringValue(in.Scope)
}
