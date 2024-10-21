package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

type prefixFilter struct {
	Prefix types.String `tfsdk:"prefix"`
	GeMask types.Int64  `tfsdk:"ge_mask"`
	LeMask types.Int64  `tfsdk:"le_mask"`
	Action types.String `tfsdk:"action"`
}

func (o prefixFilter) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"prefix": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 or IPv6 network address specified in the form of network/prefixlen.",
			Required:            true,
			Validators:          []validator.String{apstravalidator.ParseCidr(false, false)},
		},
		"ge_mask": resourceSchema.Int64Attribute{
			MarkdownDescription: "Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix " +
				"length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the " +
				"prefix-list entry should be an exact match. The option can be optionally be used in combination " +
				"with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` " +
				"are both specified, then `le_mask` must be greater than `ge_mask`.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.AtLeast(1)},
		},
		"le_mask": resourceSchema.Int64Attribute{
			MarkdownDescription: "Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. " +
				"Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be " +
				"an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must " +
				"be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then " +
				"`le_mask` must be greater than `ge_mask`.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.AtLeast(1)},
		},
		"action": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("If the action is %q, match the route. If the action is %q, do "+
				"not match the route. For composing complex policies, all prefix-list items will be processed in the "+
				"order specified, top-down. This allows the user to deny a subset of a route that may otherwise be "+
				"permitted.", apstra.PrefixFilterActionPermit, apstra.PrefixFilterActionDeny),
			Computed:   true,
			Optional:   true,
			Default:    stringdefault.StaticString(apstra.PrefixFilterActionPermit.String()),
			Validators: []validator.String{stringvalidator.OneOf(utils.AllValidPrefixFilterActions()...)},
		},
	}
}

func (o prefixFilter) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"prefix": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 or IPv6 network address specified in the form of network/prefixlen.",
			Computed:            true,
		},
		"ge_mask": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix " +
				"length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the " +
				"prefix-list entry should be an exact match. The option can be optionally be used in combination " +
				"with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` " +
				"are both specified, then `le_mask` must be greater than `ge_mask`.",
			Computed: true,
		},
		"le_mask": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. " +
				"Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be " +
				"an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must " +
				"be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then " +
				"`le_mask` must be greater than `ge_mask`.",
			Computed: true,
		},
		"action": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("If the action is %q, match the route. If the action is %q, do "+
				"not match the route. For composing complex policies, all prefix-list items will be processed in the "+
				"order specified, top-down. This allows the user to deny a subset of a route that may otherwise be "+
				"permitted.", apstra.PrefixFilterActionPermit, apstra.PrefixFilterActionDeny),
			Computed: true,
		},
	}
}

func (o prefixFilter) dataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"prefix": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 or IPv6 network address specified in the form of network/prefixlen.",
			Optional:            true,
			Validators:          []validator.String{apstravalidator.ParseCidr(false, false)},
		},
		"ge_mask": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix " +
				"length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the " +
				"prefix-list entry should be an exact match. The option can be optionally be used in combination " +
				"with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` " +
				"are both specified, then `le_mask` must be greater than `ge_mask`.",
			Optional: true,
		},
		"le_mask": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. " +
				"Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be " +
				"an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must " +
				"be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then " +
				"`le_mask` must be greater than `ge_mask`.",
			Optional: true,
		},
		"action": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("If the action is %q, match the route. If the action is %q, do "+
				"not match the route. For composing complex policies, all prefix-list items will be processed in the "+
				"order specified, top-down. This allows the user to deny a subset of a route that may otherwise be "+
				"permitted.", apstra.PrefixFilterActionPermit, apstra.PrefixFilterActionDeny),
			Optional: true,
		},
	}
}

func (o prefixFilter) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"prefix":  types.StringType,
		"ge_mask": types.Int64Type,
		"le_mask": types.Int64Type,
		"action":  types.StringType,
	}
}

func (o *prefixFilter) loadApiData(ctx context.Context, in *apstra.PrefixFilter, diags *diag.Diagnostics) {
	o.Prefix = types.StringValue(in.Prefix.String())
	o.GeMask = utils.Int64ValueOrNull(ctx, in.GeMask, diags)
	o.LeMask = utils.Int64ValueOrNull(ctx, in.LeMask, diags)
	o.Action = types.StringValue(in.Action.String())
}

func (o *prefixFilter) request(_ context.Context, diags *diag.Diagnostics) *apstra.PrefixFilter {
	var action apstra.PrefixFilterAction
	err := action.FromString(o.Action.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing prefix filter action string %q", o.Action.ValueString()), err.Error())
		return nil
	}

	_, prefix, err := net.ParseCIDR(o.Prefix.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing prefix %q", o.Prefix.ValueString()), err.Error())
		return nil
	}

	var geMask, leMask *int
	if !o.GeMask.IsNull() && !o.GeMask.IsUnknown() {
		m := int(o.GeMask.ValueInt64())
		geMask = &m
	}
	if !o.LeMask.IsNull() && !o.LeMask.IsUnknown() {
		m := int(o.LeMask.ValueInt64())
		leMask = &m
	}

	return &apstra.PrefixFilter{
		Action: action,
		Prefix: *prefix,
		GeMask: geMask,
		LeMask: leMask,
	}
}

func (o *prefixFilter) filterMatch(_ context.Context, candidate *prefixFilter, diags *diag.Diagnostics) bool {
	if !o.Prefix.IsNull() {
		_, filterPrefix, _ := net.ParseCIDR(o.Prefix.ValueString()) // ignore errors already caught in input validation

		candidateIp, candidatePrefix, err := net.ParseCIDR(candidate.Prefix.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing filter prefix %s", candidate.Prefix.String()), err.Error())
			return false
		}
		if !candidateIp.Equal(candidatePrefix.IP) {
			diags.AddError("unexpected API response", fmt.Sprintf(
				"API response for routing policy prefix misaligned with CIDR boundary - %q", candidatePrefix.String()))
			return false
		}

		if filterPrefix.String() != candidatePrefix.String() {
			return false
		}
	}

	if !o.GeMask.IsNull() && !o.GeMask.Equal(candidate.GeMask) {
		return false
	}

	if !o.LeMask.IsNull() && !o.LeMask.Equal(candidate.LeMask) {
		return false
	}

	if !o.Action.IsNull() && !o.Action.Equal(candidate.Action) {
		return false
	}

	return true
}
