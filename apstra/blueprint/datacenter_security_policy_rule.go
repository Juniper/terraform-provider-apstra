package blueprint

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			MarkdownDescription: fmt.Sprintf("Security Policy Rule Protocol; one of: '%s'", strings.Join(friendlyPolicyRuleProtocols(), "', '")),
			Computed:            true,
		},
		"action": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Security Policy Rule Action; one of: %s", enum.PolicyRuleActions),
			Computed:            true,
		},
		"source_ports": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Set of TCP/UDP source ports matched by this rule. A `null` "+
				"set matches any port. Applies only when `protocol` is `%s` or `%s`.",
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolTcp),
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolUdp),
			),
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRulePortRange{}.DataSourceAttributes(),
			},
		},
		"destination_ports": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Set of TCP/UDP destination ports matched by this rule. A `null` "+
				"set matches any port. Applies only when `protocol` is `%s` or `%s`.",
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolTcp),
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolUdp),
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

func (o DatacenterSecurityPolicyRule) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule ID.",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Name.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Security Policy Rule Description.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"protocol": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Security Policy Rule Protocol; one of: '%s'", strings.Join(friendlyPolicyRuleProtocols(), "', '")),
			Required:            true,
			Validators:          []validator.String{stringvalidator.OneOf(friendlyPolicyRuleProtocols()...)},
		},
		"action": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Action - One of: %s", enum.PolicyRuleActions),
			Required:            true,
			Validators:          []validator.String{stringvalidator.OneOf(enum.PolicyRuleActions.Values()...)},
		},
		"source_ports": resourceSchema.SetNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Set of TCP/UDP source ports matched by this rule. A `null` "+
				"set matches any port. Valid only when `protocol` is `%s` or `%s`.",
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolTcp),
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolUdp),
			),
			Optional: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				// todo: validate protocol has tcp or udp - maybe a nOfValidators() validator?
			},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRulePortRange{}.ResourceAttributes(),
			},
		},
		"destination_ports": resourceSchema.SetNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Set of TCP/UDP destination ports matched by this rule. A `null` "+
				"set matches any port. Valid only when `protocol` is `%s` or `%s`.",
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolTcp),
				utils.StringersToFriendlyString(enum.PolicyRuleProtocolUdp),
			),
			Optional: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				// todo: validate protocol has tcp or udp - maybe a nOfValidators() validator?
			},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRulePortRange{}.ResourceAttributes(),
			},
		},
		"established": resourceSchema.BoolAttribute{
			MarkdownDescription: "When `true`, the rendered rule will use the NOS `established` or `tcp-established` " +
				"keyword/feature for TCP access control list entries.",
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueAtMustBeBool(
					path.MatchRelative().AtParent().AtName("protocol").Resolve(),
					types.StringValue(enum.PolicyRuleProtocolTcp.Value),
					apstravalidator.ValueAtMustBeBool(
						path.MatchRelative().AtParent().AtName("protocol").Resolve(),
						types.StringValue(enum.PolicyRuleProtocolTcp.Value),
						true,
					),
				),
			},
		},
	}
}

func (o *DatacenterSecurityPolicyRule) loadApiData(ctx context.Context, in *apstra.PolicyRuleData, diags *diag.Diagnostics) {
	var established types.Bool
	if in.Protocol == enum.PolicyRuleProtocolTcp {
		if in.TcpStateQualifier == nil {
			established = types.BoolValue(false)
		} else {
			established = types.BoolValue(in.TcpStateQualifier.Value == enum.TcpStateQualifierEstablished.Value)
		}
	}

	o.Name = types.StringValue(in.Label)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diags)
	o.Protocol = types.StringValue(utils.StringersToFriendlyString(in.Protocol))
	o.Action = types.StringValue(in.Action.Value)
	o.SrcPorts = newDatacenterPolicyRulePortRangeSet(ctx, in.SrcPort, diags)
	o.DstPorts = newDatacenterPolicyRulePortRangeSet(ctx, in.DstPort, diags)
	o.Established = established
}

func (o *DatacenterSecurityPolicyRule) request(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.PolicyRuleData {
	var protocol enum.PolicyRuleProtocol
	err := utils.ApiStringerFromFriendlyString(&protocol, o.Protocol.ValueString())
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed to parse policy rule protocol %s", o.Protocol), err.Error())
		return nil
	}

	action := enum.PolicyRuleActions.Parse(o.Action.ValueString())
	if action == nil {
		diags.AddAttributeError(
			path.AtName("action"),
			constants.ErrStringParse,
			fmt.Sprintf("failed to parse action %s", o.Action))
		return nil
	}

	srcPort := portRangeSetToApstraPortRanges(ctx, o.SrcPorts, diags)
	dstPort := portRangeSetToApstraPortRanges(ctx, o.DstPorts, diags)
	if diags.HasError() {
		return nil
	}

	var tcpStateQualifier *enum.TcpStateQualifier
	if o.Established.ValueBool() {
		tcpStateQualifier = &enum.TcpStateQualifierEstablished
	}

	return &apstra.PolicyRuleData{
		Label:             o.Name.ValueString(),
		Description:       o.Description.ValueString(),
		Protocol:          protocol,
		Action:            *action,
		SrcPort:           srcPort,
		DstPort:           dstPort,
		TcpStateQualifier: tcpStateQualifier,
	}
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

func policyRuleListToApstraPolicyRuleSlice(ctx context.Context, ruleList types.List, diags *diag.Diagnostics) []apstra.PolicyRule {
	var ruleSlice []DatacenterSecurityPolicyRule
	diags.Append(ruleList.ElementsAs(ctx, &ruleSlice, false)...)
	if diags.HasError() {
		return nil
	}

	if len(ruleSlice) == 0 {
		return nil
	}

	result := make([]apstra.PolicyRule, len(ruleSlice))
	for i, rule := range ruleSlice {
		result[i] = apstra.PolicyRule{
			Data: rule.request(ctx, path.Root("rules").AtListIndex(i), diags),
		}
	}

	return result
}
