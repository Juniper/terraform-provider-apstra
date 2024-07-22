package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"regexp"

	"github.com/Juniper/apstra-go-sdk/apstra"
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

//{
//  "parent_id": "string",
//  "label": "string",
//  "tags": [
//    "string"
//  ],
//  "data": {}
//}

type FreeformRaGroup struct {
	BlueprintId types.String         `tfsdk:"blueprint_id"`
	Id          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	ParentId    types.String         `tfsdk:"parent_id"`
	Tags        types.Set            `tfsdk:"tags"`
	Data        jsontypes.Normalized `tfsdk:"data"`
	GeneratorId types.String         `tfsdk:"generator_id"`
}

func (o FreeformRaGroup) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Resource Allocation Group lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Allocation Group by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up the Allocation Group by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"parent_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group node that is present as a parent of the current one in " +
				"parent/children relationship." +
				" If group is a top-level one, then 'parent_id' is equal to None/null.",
			Computed: true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"data": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Arbitrary key-value mapping that is useful in a context of this group. " +
				"For example, you can store some VRF-related data there or add properties that are useful " +
				"only in context of resource allocation, but not systems or interfaces.",
			Computed:   true,
			CustomType: jsontypes.NormalizedType{},
		},
		"generator_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group generator that created the group.",
			Computed:            true,
		},
	}
}

func (o FreeformRaGroup) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Freeform Resource Allocation Group.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform Resource Allocation Group name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"), "name may consist only of the following characters : a-zA-Z0-9.-_")},
		},
		"parent_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Type of the System. Must be one of `%s` or `%s`", apstra.SystemTypeInternal, apstra.SystemTypeExternal),
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Optional:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"data": resourceSchema.StringAttribute{
			MarkdownDescription: "Arbitrary key-value mapping that is useful in a context of this group. " +
				"For example, you can store some VRF-related data there or add properties that are useful" +
				" only in context of resource allocation, but not systems or interfaces. ",
			Optional:   true,
			Computed:   true,
			Default:    stringdefault.StaticString("{}"),
			CustomType: jsontypes.NormalizedType{},
		},
		"generator_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Generator that created Resource Allocation Group, " +
				"always `null` because groups created with this resource were not generated.",
			Computed: true,
		},
	}
}

func (o *FreeformRaGroup) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.FreeformRaGroupData {
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil
	}

	return &apstra.FreeformRaGroupData{
		ParentId:    (*apstra.ObjectId)(o.ParentId.ValueStringPointer()),
		Label:       o.Name.ValueString(),
		Tags:        tags,
		Data:        json.RawMessage(o.Data.ValueString()),
		GeneratorId: (*apstra.ObjectId)(o.GeneratorId.ValueStringPointer()),
	}
}

func (o *FreeformRaGroup) LoadApiData(ctx context.Context, in *apstra.FreeformRaGroupData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	if in.ParentId != nil {
		o.ParentId = types.StringValue(string(*in.ParentId))
	}

	o.Data = jsontypes.NewNormalizedValue(string(in.Data))
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags) // safe to ignore diagnostic here
	o.GeneratorId = types.StringPointerValue((*string)(in.GeneratorId))
}
