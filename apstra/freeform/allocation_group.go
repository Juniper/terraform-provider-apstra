package freeform

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

type AllocGroup struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	PoolIds     types.Set    `tfsdk:"pool_ids"`
}

func (o AllocGroup) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Allocation Group lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Allocation Group by ID. Required when `name` is omitted.",
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
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("type of the Allocation Group, must be one of :\n  - `" +
				strings.Join(utils.AllResourcePoolTypes(), "`\n  - `") + "`\n"),
			Computed: true,
		},
		"pool_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs of Resource Pools assigned to the allocation group",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o AllocGroup) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Freeform Allocation Group.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform Allocation Group name as shown in the Web UI.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"),
					"name may consist only of the following characters : a-zA-Z0-9.-_"),
			},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("type of the Allocation Group, must be one of :\n  - `" +
				strings.Join(utils.AllResourcePoolTypes(), "`\n  - `") + "`\n"),
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:    []validator.String{stringvalidator.OneOf(utils.FFResourceTypes()...)},
		},
		"pool_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "IDs of Resource Pools assigned to the allocation group",
			ElementType:         types.StringType,
			Required:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *AllocGroup) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.FreeformAllocGroupData {
	// unpack
	var allocGroupType apstra.ResourcePoolType
	err := utils.ApiStringerFromFriendlyString(&allocGroupType, o.Type.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing type %q", o.Type.ValueString()), err.Error())
	}

	var poolIds []apstra.ObjectId
	diags.Append(o.PoolIds.ElementsAs(ctx, &poolIds, false)...)
	if diags.HasError() {
		return nil
	}

	return &apstra.FreeformAllocGroupData{
		Name:    o.Name.ValueString(),
		Type:    allocGroupType,
		PoolIds: poolIds,
	}
}

func (o *AllocGroup) LoadApiData(ctx context.Context, in *apstra.FreeformAllocGroupData, diags *diag.Diagnostics) {
	// pack
	o.Name = types.StringValue(in.Name)
	o.Type = types.StringValue(utils.StringersToFriendlyString(in.Type))
	o.PoolIds = utils.SetValueOrNull(ctx, types.StringType, in.PoolIds, diags)
}
