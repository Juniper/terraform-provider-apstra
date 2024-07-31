package blueprint

import (
	"context"
	"fmt"
	"regexp"

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

type FreeformAllocGroup struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	PoolId      types.String `tfsdk:"pool_id"`
}

func (o FreeformAllocGroup) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the System lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform System by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up System by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "type of the Resource Pool, either Internal or External",
			Computed:            true,
		},
		"pool_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Pool ID assigned to the allocation group",
			Computed:            true,
		},
	}
}

func (o FreeformAllocGroup) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Freeform System.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform System name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"), "name may consist only of the following characters : a-zA-Z0-9.-_")},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Type of the System. Must be one of `%s` or `%s`", apstra.SystemTypeInternal, apstra.SystemTypeExternal),
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.OneOf(apstra.SystemTypeInternal.String(), apstra.SystemTypeExternal.String())},
		},
		"pool_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID (usually serial number) of the Managed Device to associate with this System",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o *FreeformAllocGroup) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.FreeformAllocGroupData {

	return &apstra.FreeformAllocGroupData{
		Name:    "",
		Type:    apstra.ResourcePoolType{},
		PoolIds: nil,
	}
}

func (o *FreeformAllocGroup) LoadApiData(ctx context.Context, in *apstra.FreeformAllocGroupData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Name)
	o.Type = types.StringValue(in.Type.String())
}
