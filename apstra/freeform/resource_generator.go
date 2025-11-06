package freeform

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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

type ResourceGenerator struct {
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Id              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	Scope           types.String `tfsdk:"scope"`
	AllocatedFrom   types.String `tfsdk:"allocated_from"`
	ContainerId     types.String `tfsdk:"container_id"`
	SubnetPrefixLen types.Int64  `tfsdk:"subnet_prefix_len"`
}

func (o ResourceGenerator) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Resource lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Resource Generator by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Type of the Resource Generator",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up Resource Generator by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"scope": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Scope is a graph query which selects target nodes for which Resources should be generated.\n" +
				"Example: `node('system', name='target', label=aeq('*prod*'))`",
			Computed: true,
		},
		"allocated_from": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Selects the Allocation Group, parent Resource, or Local Resource Pool from which to " +
				"source generated Resources. In the case of a Local Resource Pool, this value must be the name (label) " +
				"of the pool. Allocation Groups and parent Resources are specified by ID.",
			Computed: true,
		},
		"container_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group used to organize the generated resources",
			Computed:            true,
		},
		"subnet_prefix_len": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Length of the subnet for the generated resources, if any.",
			Computed:            true,
		},
	}
}

func (o ResourceGenerator) ResourceAttributes() map[string]resourceSchema.Attribute {
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
			MarkdownDescription: "Freeform Resource Generator name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(apstraregexp.StdNameConstraint, apstraregexp.StdNameConstraintMsg),
			},
		},
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope is a graph query which selects target nodes for which Resources should be generated.\n" +
				"Example: `node('system', name='target', label=aeq('*prod*'))`\n" +
				"Required when `container_id` references a `apstra_freeform_resource_group` object. Must be `null` when " +
				"`container_id` references a `apstra_freeform_resource_group` object. `scope` will be inherited in that case.",
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"allocated_from": resourceSchema.StringAttribute{
			MarkdownDescription: "Selects the Allocation Group, parent Resource, or Local Resource Pool from which to " +
				"source generated Resources. In the case of a Local Resource Pool, this value must be the name (label) " +
				"of the pool. Allocation Groups and parent Resources are specified by ID.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"container_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group where Resources are generated. ",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"subnet_prefix_len": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("Length of the subnet for the generated Resources. "+
				"Only applicable when `type` is `%s` or `%s`",
				rosetta.StringersToFriendlyString(enum.FFResourceTypeIpv4),
				rosetta.StringersToFriendlyString(enum.FFResourceTypeIpv6),
			),
			Optional: true,
			Validators: []validator.Int64{
				int64validator.Between(1, 127),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeAsn))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeHostIpv4))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeHostIpv6))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeInt))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeVlan))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeVni))),
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeIpv4))),
				apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(rosetta.StringersToFriendlyString(enum.FFResourceTypeIpv6))),
			},
		},
	}
}

func (o *ResourceGenerator) Request(_ context.Context, diags *diag.Diagnostics) *apstra.FreeformResourceGeneratorData {
	var resourceType enum.FFResourceType
	err := rosetta.ApiStringerFromFriendlyString(&resourceType, o.Type.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing type %q", o.Type.ValueString()), err.Error())
	}

	var scopeNodePoolLabel *string
	var allocatedFrom *apstra.ObjectId
	if resourceType == enum.FFResourceTypeVlan {
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

func (o *ResourceGenerator) LoadApiData(ctx context.Context, in *apstra.FreeformResourceGeneratorData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Scope = value.StringOrNull(ctx, in.Scope, diags)
	o.Type = types.StringValue(rosetta.StringersToFriendlyString(in.ResourceType))
	if in.ResourceType == enum.FFResourceTypeVlan {
		o.AllocatedFrom = types.StringPointerValue(in.ScopeNodePoolLabel)
	} else {
		o.AllocatedFrom = types.StringPointerValue((*string)(in.AllocatedFrom))
	}
	o.ContainerId = types.StringValue(string(in.ContainerId))
	o.SubnetPrefixLen = value.Int64FromPointer(in.SubnetPrefixLen)
}
