package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterSecurityPolicyRulePortRange struct {
	FromPort types.Int64 `tfsdk:"from_port"`
	ToPort   types.Int64 `tfsdk:"to_port"`
}

func (o DatacenterSecurityPolicyRulePortRange) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"from_port": types.Int64Type,
		"to_port":   types.Int64Type,
	}
}

func (o DatacenterSecurityPolicyRulePortRange) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"from_port": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "First (low) port number in a range of ports matched by the policy rule.",
			Computed:            true,
		},
		"to_port": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Last (high) port number in a range of ports matched by the policy rule.",
			Computed:            true,
		},
	}
}

func (o DatacenterSecurityPolicyRulePortRange) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"from_port": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "First (low) port number in a range of ports matched by the policy rule. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
		"to_port": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Last (high) port number in a range of ports matched by the policy rule. Do you need this? " +
				"Let us know by [opening an issue](https://github.com/Juniper/terraform-provider-apstra/issues/new)!",
			Computed: true,
		},
	}
}

func (o DatacenterSecurityPolicyRulePortRange) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"from_port": resourceSchema.Int64Attribute{
			MarkdownDescription: "First (low) port number in a range of ports matched by the policy rule.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.Between(1, 65535)},
		},
		"to_port": resourceSchema.Int64Attribute{
			MarkdownDescription: "Last (high) port number in a range of ports matched by the policy rule.",
			Required:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, 65535),
				int64validator.AtLeastSumOf(path.MatchRelative().AtParent().AtName("from_port")),
			},
		},
	}
}

func (o *DatacenterSecurityPolicyRulePortRange) loadApiData(_ context.Context, data *apstra.PortRange, _ *diag.Diagnostics) {
	o.FromPort = types.Int64Value(int64(data.First))
	o.ToPort = types.Int64Value(int64(data.Last))
}

func (o *DatacenterSecurityPolicyRulePortRange) request(_ context.Context, _ path.Path, _ *diag.Diagnostics) *apstra.PortRange {
	return &apstra.PortRange{
		First: uint16(o.FromPort.ValueInt64()),
		Last:  uint16(o.ToPort.ValueInt64()),
	}
}

func newDatacenterPolicyRulePortRangeSet(ctx context.Context, in []apstra.PortRange, diags *diag.Diagnostics) types.Set {
	if len(in) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRulePortRange{}.attrTypes()})
	}

	portRanges := make([]attr.Value, len(in))
	for i, inRange := range in {
		var portRange DatacenterSecurityPolicyRulePortRange
		portRange.loadApiData(ctx, &inRange, diags)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRulePortRange{}.attrTypes()})
		}

		var d diag.Diagnostics
		portRanges[i], d = types.ObjectValueFrom(ctx, DatacenterSecurityPolicyRulePortRange{}.attrTypes(), portRange)
		diags.Append(d...)
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRulePortRange{}.attrTypes()})
	}

	return types.SetValueMust(types.ObjectType{AttrTypes: DatacenterSecurityPolicyRulePortRange{}.attrTypes()}, portRanges)
}

func portRangeSetToApstraPortRanges(ctx context.Context, portSet types.Set, diags *diag.Diagnostics) apstra.PortRanges {
	var portRangeSlice []DatacenterSecurityPolicyRulePortRange
	diags.Append(portSet.ElementsAs(ctx, &portRangeSlice, false)...)
	if diags.HasError() {
		return nil
	}

	if len(portRangeSlice) == 0 {
		return nil
	}

	result := make(apstra.PortRanges, len(portRangeSlice))
	for i, portRange := range portRangeSlice {
		result[i] = apstra.PortRange{
			First: uint16(portRange.FromPort.ValueInt64()),
			Last:  uint16(portRange.ToPort.ValueInt64()),
		}
	}

	return result
}
