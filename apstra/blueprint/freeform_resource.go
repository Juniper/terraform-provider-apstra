package blueprint

import (
	"context"
	"fmt"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"regexp"
	"strconv"
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

type FreeformResource struct {
	BlueprintId   types.String         `tfsdk:"blueprint_id"`
	Id            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	GroupId       types.String         `tfsdk:"group_id"`
	Type          types.String         `tfsdk:"type"`
	AllocatedFrom types.String         `tfsdk:"allocated_from"`
	IntValue      types.Int64          `tfsdk:"integer_value"`
	Ipv4Value     cidrtypes.IPv4Prefix `tfsdk:"ipv4_value"`
	Ipv6Value     cidrtypes.IPv6Prefix `tfsdk:"ipv6_value"`
	GeneratorId   types.String         `tfsdk:"generator_id"`
}

func (o FreeformResource) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up Resource by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"group_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Group the Resource belongs to",
			Computed:            true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "type of the Resource, either asn | ipv6 | host_ip | host_ipv6 | vni | integer | ip | vlan",
			Computed:            true,
		},
		"integer_value": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "integer value of the Resource",
			Computed:            true,
		},
		"allocated_from": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the node that works as a source for this resource. This could be an ID " +
				"of resource allocation group or another resource (in case of IP/Host IP allocation). " +
				"This also can be empty. In that case it is required that value for this resource is provided by thex user.",
			Computed: true,
		},
		"subnet_prefix_len": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Length of subnet prefix",
			Computed:            true,
		},
		"generator_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the group generator that created the group, if any.",
			Computed:            true,
		},
	}
}

func (o FreeformResource) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Freeform Resource.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform Resource name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"), "name may consist only of the following characters : a-zA-Z0-9.-_")},
		},
		"group_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Group the Resource belongs to.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "type of the Resource, must be one of :\n  - " +
				strings.Join(utils.AllResourceTypes(), "\n  - ") + "\n",
			Optional:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllResourceTypes()...)},
		},
		"integer_value": resourceSchema.Int64Attribute{
			MarkdownDescription: "integer value of the Resource",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeAsn))),
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeInt))),
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeIpv4))),
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeIpv6))),
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeVlan))),
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeVni))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeHostIpv4))),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeHostIpv6))),
			},
		},
		"ipv4_value": resourceSchema.StringAttribute{
			MarkdownDescription: "ipv4 value of the Resource in CIDR notation.",
			Optional:            true,
			Computed:            true,
			CustomType:          cidrtypes.IPv4PrefixType{},
			Validators:          []validator.String{
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeIpv4))),
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeHostIpv4))),
			},
		},
		"ipv6_value": resourceSchema.StringAttribute{
			MarkdownDescription: "ipv6 value of the Resource in CIDR notation.",
			Optional:            true,
			Computed:            true,
			CustomType:          cidrtypes.IPv6PrefixType{},
			Validators:          []validator.String{
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeIpv6))),
				//apstravalidator.RequiredWhenValueIs(path.MatchRoot("type"), types.StringValue(utils.StringersToFriendlyString(apstra.FFResourceTypeHostIpv6))),
			},
		},
		"allocated_from": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the node that works as a source for this resource. This could be an ID " +
				"of resource allocation group or another resource (in case of IP/Host IP allocation). " +
				"This also can be empty. In that case it is required that value for this resource is provided by the user.",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				//stringvalidator.ConflictsWith(
				//	path.MatchRoot("integer_value"),
				//	path.MatchRoot("ipv4_value"),
				//	path.MatchRoot("ipv6_value"),
				//),
			},
		},
		"generator_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Generator that created Resource Allocation Group. " +
				"Always `null` because groups created via resource declaration were not generated.",
			Computed: true,
		},
	}
}

func (o *FreeformResource) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.FreeformRaResourceData {
	var resourceType apstra.FFResourceType
	err := utils.ApiStringerFromFriendlyString(&resourceType, o.Type.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing type %q", o.Type.ValueString()), err.Error())
	}

	typeIpv4 := o.Type.ValueString() == utils.StringersToFriendlyString(apstra.FFResourceTypeIpv4)
	typeIpv6 := o.Type.ValueString() == utils.StringersToFriendlyString(apstra.FFResourceTypeIpv6)

	var subnetPrefixLen *int
	if typeIpv4 || typeIpv6 {
		subnetPrefixLen = utils.ToPtr(int(o.IntValue.ValueInt64()))
	}

	var value *string
	switch {
	case utils.HasValue(o.Ipv4Value):
		value = o.Ipv4Value.ValueStringPointer()
	case utils.HasValue(o.Ipv6Value):
		value = o.Ipv6Value.ValueStringPointer()
	case utils.HasValue(o.IntValue) && !typeIpv4 && !typeIpv6:
		value = utils.ToPtr(strconv.FormatInt(o.IntValue.ValueInt64(), 10))
	}

	return &apstra.FreeformRaResourceData{
		ResourceType:    resourceType,
		Label:           o.Name.ValueString(),
		Value:           value,
		AllocatedFrom:   (*apstra.ObjectId)(o.AllocatedFrom.ValueStringPointer()),
		GroupId:         apstra.ObjectId(o.GroupId.ValueString()),
		SubnetPrefixLen: subnetPrefixLen,
	}
}

func (o *FreeformResource) LoadApiData(ctx context.Context, in *apstra.FreeformRaResourceData, diags *diag.Diagnostics) {
	o.Type = types.StringValue(utils.StringersToFriendlyString(in.ResourceType))
	o.Name = types.StringValue(in.Label)
	o.AllocatedFrom = types.StringPointerValue((*string)(in.AllocatedFrom))
	o.GroupId = types.StringValue(string(in.GroupId))
	o.IntValue = types.Int64Null()
	o.Ipv4Value = cidrtypes.NewIPv4PrefixNull()
	o.Ipv6Value = cidrtypes.NewIPv6PrefixNull()
	o.GeneratorId = types.StringPointerValue((*string)(in.GeneratorId))

	var err error
	var i int64
	if in.Value != nil {
		switch in.ResourceType {
		case apstra.FFResourceTypeAsn,
			apstra.FFResourceTypeVni,
			apstra.FFResourceTypeVlan,
			apstra.FFResourceTypeInt:
			i, err = strconv.ParseInt(*in.Value, 10, 64)
			o.IntValue = types.Int64Value(i)
		case apstra.FFResourceTypeHostIpv4:
			o.Ipv4Value = cidrtypes.NewIPv4PrefixValue(*in.Value)
		case apstra.FFResourceTypeHostIpv6:
			o.Ipv6Value = cidrtypes.NewIPv6PrefixValue(*in.Value)
		case apstra.FFResourceTypeIpv4:
			o.Ipv4Value = cidrtypes.NewIPv4PrefixValue(*in.Value)
			if in.SubnetPrefixLen != nil {
				o.IntValue = types.Int64Value(int64(*in.SubnetPrefixLen))
			}
		case apstra.FFResourceTypeIpv6:
			o.Ipv6Value = cidrtypes.NewIPv6PrefixValue(*in.Value)
			if in.SubnetPrefixLen != nil {
				o.IntValue = types.Int64Value(int64(*in.SubnetPrefixLen))
			}
		}
		if err != nil {
			diags.AddError("failed parsing integer value from API", err.Error())
		}
	}
}
