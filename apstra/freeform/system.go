package freeform

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
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

type System struct {
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	DeviceProfileId types.String `tfsdk:"device_profile_id"`
	Hostname        types.String `tfsdk:"hostname"`
	Type            types.String `tfsdk:"type"`
	SystemId        types.String `tfsdk:"system_id"`
	DeployMode      types.String `tfsdk:"deploy_mode"`
	Tags            types.Set    `tfsdk:"tags"`
}

func (o System) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
		"hostname": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Hostname of the System",
			Computed:            true,
		},
		"deploy_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "deploy mode of the System",
			Computed:            true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "type of the System, either Internal or External",
			Computed:            true,
		},
		"device_profile_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "device profile ID of the System",
			Computed:            true,
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Device System ID assigned to the System",
			Computed:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

func (o System) ResourceAttributes() map[string]resourceSchema.Attribute {
	hostnameRegexp := "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$"
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
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"), "name may consist only of the following characters : a-zA-Z0-9.-_"),
			},
		},
		"hostname": resourceSchema.StringAttribute{
			MarkdownDescription: "Hostname of the Freeform System.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(hostnameRegexp), "must match regex "+hostnameRegexp),
			},
		},
		"deploy_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "Deploy mode of the System",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllNodeDeployModes()...)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Type of the System. Must be one of `%s` or `%s`",
				utils.StringersToFriendlyString(apstra.SystemTypeInternal),
				utils.StringersToFriendlyString(apstra.SystemTypeExternal),
			),
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{stringvalidator.OneOf(
				utils.StringersToFriendlyString(apstra.SystemTypeInternal),
				utils.StringersToFriendlyString(apstra.SystemTypeExternal),
			)},
		},
		"device_profile_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Device profile ID of the System. Required when `type` is %q.",
				utils.StringersToFriendlyString(apstra.SystemTypeInternal)),
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.SystemTypeExternal))),
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.SystemTypeInternal))),
			},
		},
		"system_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID (usually serial number) of the Managed Device to associate with this System",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Optional:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *System) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.FreeformSystemData {
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil
	}

	var systemType apstra.SystemType
	switch o.Type.ValueString() {
	case utils.StringersToFriendlyString(apstra.SystemTypeExternal):
		systemType = apstra.SystemTypeExternal
	case utils.StringersToFriendlyString(apstra.SystemTypeInternal):
		systemType = apstra.SystemTypeInternal
	default:
		diags.AddError("unexpected system type", "got: "+o.Type.ValueString())
	}

	return &apstra.FreeformSystemData{
		SystemId:        (*apstra.ObjectId)(o.SystemId.ValueStringPointer()),
		Type:            systemType,
		Label:           o.Name.ValueString(),
		Hostname:        o.Hostname.ValueString(),
		Tags:            tags,
		DeviceProfileId: (*apstra.ObjectId)(o.DeviceProfileId.ValueStringPointer()),
	}
}

func (o *System) LoadApiData(ctx context.Context, in *apstra.FreeformSystemData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Hostname = types.StringValue(in.Hostname)
	o.Type = types.StringValue(utils.StringersToFriendlyString(in.Type))
	o.DeviceProfileId = types.StringPointerValue((*string)(in.DeviceProfileId))
	o.SystemId = types.StringPointerValue((*string)(in.SystemId))
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags) // safe to ignore diagnostic here
}
