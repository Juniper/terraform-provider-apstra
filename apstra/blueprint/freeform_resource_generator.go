package blueprint

import (
	"context"
	"fmt"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"regexp"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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

type FreeformResourceGenerator struct {
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Id              types.String `tfsdk:"id"`
	ResourceType    types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	Scope           types.String `tfsdk:"scope"`
	AllocatedFrom   types.String `tfsdk:"allocated_from"`
	ContainerId     types.String `tfsdk:"container_id"`
	SubnetPrefixLen types.Int64  `tfsdk:"subnet_prefix_len"`
}

func (o FreeformResourceGenerator) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Resource lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Resource by ID. Required when `name` is omitted.",
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
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Type of the Resource",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up Resource by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"scope": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Scope the Resource Generator uses for resource generation",
			Computed:            true,
		},
		"allocated_from": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the node from which this resource generator has been sourced. This could be an ID " +
				"of resource generator or another resource (in case of IP or Host IP allocations). " +
				"This also can be empty. In that case it is required that value for this resource is provided by the user.",
			Computed: true,
		},
		"container_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group generator that created the group, if any.",
			Computed:            true,
		},
		"subnet_prefix_len": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Length of the subnet for the generated resources, if any.",
			Computed:            true,
		},
	}
}

func (o FreeformResourceGenerator) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Resource Generator within the Freeform Blueprint.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "type of the Resource Generator, must be one of :\n  - `" +
				strings.Join(utils.AllFFResourceTypes(), "`\n  - `") + "`\n",
			Required:      true,
			Validators:    []validator.String{stringvalidator.OneOf(utils.AllFFResourceTypes()...)},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform Resource name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"), "name may consist only of the following characters : a-zA-Z0-9.-_"),
			},
		},
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope the Resource Generator uses for resource generation.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"allocated_from": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the node to be used as a source for this resource. This could be an ID " +
				"of resource group or another resource (in case of IP or Host IP allocations). " +
				"This also can be empty. In that case it is required that value for this resource is provided by the user.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"container_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group where resources are generated. ",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"subnet_prefix_len": resourceSchema.Int64Attribute{
			MarkdownDescription: "Length of the subnet for the generated resources, if any.",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, 127),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeAsn))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeHostIpv4))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeHostIpv6))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeInt))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeVlan))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeVni))),
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeIpv4))),
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("allocated_from"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeIpv6))),
			},
		},
	}
}

func (o *FreeformResourceGenerator) Request(_ context.Context, diags *diag.Diagnostics) *apstra.FreeformResourceGeneratorData {
	var resourceType apstra.FFResourceType
	err := utils.ApiStringerFromFriendlyString(&resourceType, o.ResourceType.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing type %q", o.ResourceType.ValueString()), err.Error())
	}

	var scopeNodePoolLabel *string
	var allocatedFrom *apstra.ObjectId

	if resourceType == apstra.FFResourceTypeVlan {
		scopeNodePoolLabel = o.AllocatedFrom.ValueStringPointer()
	} else {
		allocatedFrom = (*apstra.ObjectId)(o.AllocatedFrom.ValueStringPointer())
	}

	var subnetPrefixLen *int
	if !o.SubnetPrefixLen.IsNull() {
		l := int(o.SubnetPrefixLen.ValueInt64())
		subnetPrefixLen = &l
	}

	return &apstra.FreeformResourceGeneratorData{
		ResourceType:       resourceType,
		Label:              o.Name.ValueString(),
		Scope:              o.Scope.ValueString(),
		AllocatedFrom:      allocatedFrom,
		ScopeNodePoolLabel: scopeNodePoolLabel,
		ContainerId:        apstra.ObjectId(o.ContainerId.ValueString()),
		SubnetPrefixLen:    subnetPrefixLen,
	}
}

func (o *FreeformResourceGenerator) LoadApiData(_ context.Context, in *apstra.FreeformResourceGeneratorData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Scope = types.StringValue(in.Scope)
	o.ResourceType = types.StringValue(utils.StringersToFriendlyString(in.ResourceType))
	if in.ResourceType == apstra.FFResourceTypeVlan {
		o.AllocatedFrom = types.StringPointerValue(in.ScopeNodePoolLabel)
	} else {
		o.AllocatedFrom = types.StringPointerValue((*string)(in.AllocatedFrom))
	}
	o.ContainerId = types.StringValue(string(in.ContainerId))
	o.SubnetPrefixLen = int64AttrValueFromPtr(in.SubnetPrefixLen)
}
