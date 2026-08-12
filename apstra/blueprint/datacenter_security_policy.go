package blueprint

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterSecurityPolicy struct {
	BlueprintId   types.String `tfsdk:"blueprint_id"`
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	IPVersion     types.String `tfsdk:"ip_version"`
	SrcAppPointId types.String `tfsdk:"source_application_point_id"`
	DstAppPointId types.String `tfsdk:"destination_application_point_id"`
	Rules         types.List   `tfsdk:"rules"`
	Tags          types.Set    `tfsdk:"tags"`
}

func (o DatacenterSecurityPolicy) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Security Policy by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up a Security Policy by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Description of the Security Policy as seen in the Web UI.",
			Computed:            true,
		},
		"enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the Security Policy is enabled.",
			Computed:            true,
		},
		"ip_version": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy IP version.",
			Computed:            true,
		},
		"source_application_point_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the source Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Computed:            true,
		},
		"destination_application_point_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the destination Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Computed:            true,
		},
		"rules": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "List of Rules associated with the Security Policy.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRule{}.DataSourceAttributes(),
			},
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags associated with the Security Policy.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o DatacenterSecurityPolicy) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy name.",
			Optional:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy description, as seen in the Web UI.",
			Optional:            true,
		},
		"enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the Security Policy is enabled.",
			Optional:            true,
		},
		"ip_version": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy IP version.",
			Optional:            true,
		},
		"source_application_point_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the source Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Optional:            true,
		},
		"destination_application_point_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the destination Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Optional:            true,
		},
		"rules": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Not currently supported for use in a filter. Do you need this? Let us know by " +
				"[opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRule{}.DataSourceFilterAttributes(),
			},
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags. All tags supplied here are used to match the Security Policy, " +
				"but a matching Security Policy may have additional tags not enumerated in this set.",
			Optional:    true,
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
	}
}

func (o DatacenterSecurityPolicy) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy name.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy description, as seen in the Web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the Security Policy is enabled. Default value: `true`",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},
		"ip_version": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(
				"Security Policy IP version. One of: `%s`. Required with Apstra %s, not compatible with Apstra %s.",
				strings.Join(enum.PolicyAddressFamilies.Values(), "`, `"),
				compatibility.DatacenterPolicyAddressFamilyRequired,
				compatibility.DatacenterPolicyAddressFamilyForbidden,
			),
			Optional:   true,
			Validators: []validator.String{stringvalidator.OneOf(enum.PolicyAddressFamilies.Values()...)},
		},
		"source_application_point_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the source Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"destination_application_point_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the destination Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"rules": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered list of policy rules.",
			Optional:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRule{}.ResourceAttributes(),
			},
			Validators: []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags applied to the Security Policy.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *DatacenterSecurityPolicy) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) error {
	var api datacenter.Policy
	var err error

	if o.Id.IsNull() {
		api, err = bp.GetPolicyByLabel(ctx, o.Name.ValueString())
	} else {
		api, err = bp.GetPolicy(ctx, o.Id.ValueString())
	}
	if err != nil {
		return err
	}

	o.Id = types.StringPointerValue(api.ID())
	o.loadApiData(ctx, api, diags)

	return nil
}

func (o *DatacenterSecurityPolicy) loadApiData(ctx context.Context, data datacenter.Policy, diags *diag.Diagnostics) {
	var ipVersion, srcAppPointId, dstAppPointId types.String
	if data.SrcApplicationPoint != nil {
		srcAppPointId = types.StringPointerValue(data.SrcApplicationPoint)
	}
	if data.DstApplicationPoint != nil {
		dstAppPointId = types.StringPointerValue(data.DstApplicationPoint)
	}
	if data.AddressFamily != nil {
		ipVersion = types.StringValue(data.AddressFamily.String())
	}

	o.Name = types.StringValue(data.Label)
	o.Description = value.StringOrNull(ctx, data.Description, diags)
	o.Enabled = types.BoolValue(data.Enabled)
	o.IPVersion = ipVersion
	o.SrcAppPointId = srcAppPointId
	o.DstAppPointId = dstAppPointId
	o.Rules = newPolicyRuleList(ctx, data.Rules, diags)
	o.Tags = value.SetOrNull(ctx, types.StringType, data.Tags, diags)
}

func (o *DatacenterSecurityPolicy) Query(resultName string) apstra.QEQuery {
	nodeAttributes := []apstra.QEEAttribute{
		apstra.NodeTypePolicy.QEEAttribute(),
		{Key: "policy_type", Value: apstra.QEStringVal("security")},
		{Key: "name", Value: apstra.QEStringVal(resultName)},
	}

	if !o.Name.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "label",
			Value: apstra.QEStringVal(o.Name.ValueString()),
		})
	}

	if !o.Description.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "description",
			Value: apstra.QEStringVal(o.Description.ValueString()),
		})
	}

	if !o.Enabled.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "enabled",
			Value: apstra.QEBoolVal(o.Enabled.ValueBool()),
		})
	}

	if !o.IPVersion.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "ip_version",
			Value: apstra.QEStringVal(o.IPVersion.ValueString()),
		})
	}

	// Begin the query with thw SP node using the attributes we've collected so far
	spNodeQuery := new(apstra.PathQuery).Node(nodeAttributes)

	// Add tag matchers for the policy node's embedded tags as needed
	if !o.Tags.IsNull() {
		for _, attrVal := range o.Tags.Elements() {
			tag := attrVal.(types.String).ValueString()
			where := fmt.Sprintf("lambda %s: '%s' in (%s.tags or [])", resultName, tag, resultName)
			spNodeQuery.Where(where)
		}
	}

	// lump the spNodeQuery into a match query so we can attach graph traversals to the application points
	matchQuery := new(apstra.MatchQuery).
		Match(spNodeQuery)

	// prepare a graph traversal to the source application point if needed
	if !o.SrcAppPointId.IsNull() {
		matchQuery.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypePolicy.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(resultName)},
			}).
			In([]apstra.QEEAttribute{
				apstra.RelationshipTypeSecurityPolicy.QEEAttribute(),
				{Key: "policy_direction", Value: apstra.QEStringVal("from")},
			}).
			Node([]apstra.QEEAttribute{
				{Key: "name", Value: apstra.QEStringVal("src_app_point_id")},
			}))
	}

	// prepare a graph traversal to the destination application point if needed
	if !o.DstAppPointId.IsNull() {
		matchQuery.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypePolicy.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(resultName)},
			}).
			In([]apstra.QEEAttribute{
				apstra.RelationshipTypeSecurityPolicy.QEEAttribute(),
				{Key: "policy_direction", Value: apstra.QEStringVal("to")},
			}).
			Node([]apstra.QEEAttribute{
				{Key: "name", Value: apstra.QEStringVal("dst_app_point_id")},
			}))
	}

	return matchQuery
}

func (o *DatacenterSecurityPolicy) Request(ctx context.Context, diags *diag.Diagnostics) datacenter.Policy {
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if tags == nil {
		tags = make([]string, 0) // we must send an empty slice to wipe out current tags
	}

	result := datacenter.Policy{
		Enabled:             o.Enabled.ValueBool(),
		Label:               o.Name.ValueString(),
		Description:         o.Description.ValueString(),
		SrcApplicationPoint: o.SrcAppPointId.ValueStringPointer(),
		DstApplicationPoint: o.DstAppPointId.ValueStringPointer(),
		AddressFamily:       nil, // potentially overwritten below
		Rules:               policyRuleListToApstraPolicyRuleSlice(ctx, o.Rules, diags),
		Tags:                make([]string, 0), // potentially overwritten below
	}
	if !o.IPVersion.IsNull() {
		result.AddressFamily = new(enum.PolicyAddressFamily)
		err := result.AddressFamily.FromString(o.IPVersion.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("parsing ip_version value %s", o.IPVersion), err.Error())
		}
	}
	if !o.Tags.IsNull() {
		diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)
	}

	_ = result.SetID(o.Id.ValueString()) // skipping error check b/c we know that result.id is currently empty
	return result
}

func (o DatacenterSecurityPolicy) VersionConstraints(_ context.Context, _ *diag.Diagnostics) compatibility.ConfigConstraints {
	var response compatibility.ConfigConstraints

	if utils.HasValue(o.IPVersion) {
		response.AddAttributeConstraints(compatibility.AttributeConstraint{
			Path:        path.Root("ip_version"),
			Constraints: compatibility.DatacenterPolicyAddressFamilyOK,
		})
	} else {
		response.AddAttributeConstraints(compatibility.AttributeConstraint{
			Path:        path.Root("ip_version"),
			Constraints: compatibility.DatacenterPolicyAddressFamilyNotRequired,
		})
	}

	return response
}
