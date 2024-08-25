package freeform

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

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
)

type ResourceGroup struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ParentId    types.String `tfsdk:"parent_id"`
	// Tags        types.Set            `tfsdk:"tags"`
	Data        jsontypes.Normalized `tfsdk:"data"`
	GeneratorId types.String         `tfsdk:"generator_id"`
}

func (o ResourceGroup) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Resource Group lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Group by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up the Freeform Group by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"parent_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group node that is present as a parent of the current one in a " +
				"parent/child relationship. If this is a top-level (root) node, then `parent_id` will be `null`.",
			Computed: true,
		},
		//"tags": dataSourceSchema.SetAttribute{
		//	MarkdownDescription: "Set of Tag labels",
		//	ElementType:         types.StringType,
		//	Computed:            true,
		//},
		"data": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Arbitrary key-value mapping that is useful in a context of this group. " +
				"For example, you can store some VRF-related data there or add properties that are useful " +
				"only in context of resource allocation, but not systems or interfaces.",
			Computed:   true,
			CustomType: jsontypes.NormalizedType{},
		},
		"generator_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group generator that created the group, if any.",
			Computed:            true,
		},
	}
}

func (o ResourceGroup) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Freeform Resource Group.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform Resource Group name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile("^[a-zA-Z0-9.-_]+$"),
					"name may consist only of the following characters : a-zA-Z0-9.-_"),
			},
		},
		"parent_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the parent Freeform Resource Group, if this group is to be nested.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		//"tags": resourceSchema.SetAttribute{
		//	MarkdownDescription: "Set of Tag labels",
		//	ElementType:         types.StringType,
		//	Optional:            true,
		//	Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		//},
		"data": resourceSchema.StringAttribute{
			MarkdownDescription: "Arbitrary JSON-encoded key-value mapping that is useful in a context of this " +
				"group. For example, you can store some VRF-related data there or add properties that are useful " +
				"only in context of resource allocation, but not systems or interfaces.",
			Optional:   true,
			Computed:   true,
			Default:    stringdefault.StaticString("{}"),
			CustomType: jsontypes.NormalizedType{},
		},
		"generator_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Generator that created Resource Group. " +
				"Always `null` because groups created via resource declaration were not generated.",
			Computed: true,
		},
	}
}

func (o *ResourceGroup) Request(_ context.Context, _ *diag.Diagnostics) *apstra.FreeformRaGroupData {
	//var tags []string
	//diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	//if diags.HasError() {
	//	return nil
	//}

	return &apstra.FreeformRaGroupData{
		ParentId: (*apstra.ObjectId)(o.ParentId.ValueStringPointer()),
		Label:    o.Name.ValueString(),
		// Tags:        tags,
		Data:        json.RawMessage(o.Data.ValueString()),
		GeneratorId: (*apstra.ObjectId)(o.GeneratorId.ValueStringPointer()),
	}
}

func (o *ResourceGroup) LoadApiData(_ context.Context, in *apstra.FreeformRaGroupData, _ *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	if in.ParentId != nil {
		o.ParentId = types.StringValue(string(*in.ParentId))
	}

	o.Data = jsontypes.NewNormalizedValue(string(in.Data))
	// o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags) // safe to ignore diagnostic here
	o.GeneratorId = types.StringPointerValue((*string)(in.GeneratorId))
}
