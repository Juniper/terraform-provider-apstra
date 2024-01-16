package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type DatacenterSecurityPolicyRule struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Protocol    types.String `tfsdk:"protocol"`
	Action      types.String `tfsdk:"action"`
	SrcPorts    types.Set    `tfsdk:"source_ports"`
	DstPorts    types.Set    `tfsdk:"destination_ports"`
	Established types.Bool   `tfsdk:"established"`
}

func (o DatacenterSecurityPolicyRule) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"name":              types.StringType,
		"description":       types.StringType,
		"protocol":          types.StringType,
		"action":            types.StringType,
		"source_ports":      types.SetType{ElemType: types.ObjectType{AttrTypes: DatacenterSecurityPolicyRulePortRange{}.attrTypes()}},
		"destination_ports": types.SetType{ElemType: types.ObjectType{AttrTypes: DatacenterSecurityPolicyRulePortRange{}.attrTypes()}},
		"established":       types.BoolType,
	}
}

func (o DatacenterSecurityPolicyRule) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule ID.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Name.",
			Computed:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Description.",
			Computed:            true,
		},
		"protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Security Policy Rule Protocol; one of: %s", strings.ToLower(fmt.Sprint(apstra.PolicyRuleProtocols))),
			Computed:            true,
		},
		"action": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Security Policy Rule Action; one of: %s", apstra.PolicyRuleActions),
			Computed:            true,
		},
		"source_ports": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Set of TCP/UDP source ports matched by this rule. A `null` "+
				"set matches any port. Applies only when `protocol` is `%s` or `%s`.",
				utils.StringersToFriendlyString(apstra.PolicyRuleProtocolTcp),
				utils.StringersToFriendlyString(apstra.PolicyRuleProtocolUdp),
			),
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRulePortRange{}.DataSourceAttributes(),
			},
		},
		"destination_ports": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Set of TCP/UDP destination ports matched by this rule. A `null` "+
				"set matches any port. Applies only when `protocol` is `%s` or `%s`.",
				utils.StringersToFriendlyString(apstra.PolicyRuleProtocolTcp),
				utils.StringersToFriendlyString(apstra.PolicyRuleProtocolUdp),
			),
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRulePortRange{}.DataSourceAttributes(),
			},
		},
		"established": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "When `true`, the rendered rule will use the NOS `established` or `tcp-established` " +
				"keyword/feature for TCP access control list entries.",
			Computed: true,
		},
	}
}

func (o DatacenterSecurityPolicyRule) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule ID. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Name. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Description. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
		"protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Protocol. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
		"action": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Action. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
		"source_ports": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of TCP/UDP source ports matched by this rule. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRulePortRange{}.DataSourceFilterAttributes(),
			},
		},
		"destination_ports": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of TCP/UDP destination ports matched by this rule. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRulePortRange{}.DataSourceFilterAttributes(),
			},
		},
		"established": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "When `true`, the rendered rule will use the NOS `established` or `tcp-established` " +
				"keyword/feature for TCP access control list entries. Not currently supported for use in a filter. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
	}
}

func (o *DatacenterSecurityPolicyRule) loadApiData(ctx context.Context, in *apstra.PolicyRuleData, diags *diag.Diagnostics) {
	var established types.Bool
	if in.TcpStateQualifier != nil {
		established = types.BoolValue(in.TcpStateQualifier.Value == apstra.TcpStateQualifierEstablished.Value)
	}

	o.Name = types.StringValue(in.Label)
	o.Description = types.StringValue(in.Description)
	o.Protocol = types.StringValue(utils.StringersToFriendlyString(in.Protocol))
	o.Action = types.StringValue(in.Action.Value)
	o.SrcPorts = newDatacenterPolicyRulePortRangeSet(ctx, in.SrcPort, diags)
	o.DstPorts = newDatacenterPolicyRulePortRangeSet(ctx, in.DstPort, diags)
	o.Established = established
}

func newPolicyRuleList(ctx context.Context, in []apstra.PolicyRule, diags *diag.Diagnostics) types.List {
	var d diag.Diagnostics

	if len(in) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRule{}.attrTypes()})
	}

	rules := make([]attr.Value, len(in))
	for i, inRule := range in {
		rule := DatacenterSecurityPolicyRule{Id: types.StringValue(inRule.Id.String())}
		rule.loadApiData(ctx, inRule.Data, diags)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRule{}.attrTypes()})
		}

		rules[i], d = types.ObjectValueFrom(ctx, DatacenterSecurityPolicyRule{}.attrTypes(), rule)
		diags.Append(d...)
	}
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRule{}.attrTypes()})
	}

	return types.ListValueMust(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRule{}.attrTypes()}, rules)
}
