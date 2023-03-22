package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
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
		},
		"ge_mask": resourceSchema.Int64Attribute{
			MarkdownDescription: "Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix " +
				"length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the " +
				"prefix-list entry should be an exact match. The option can be optionally be used in combination " +
				"with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` " +
				"are both specified, then `le_mask` must be greater than `ge_mask`.",
			Optional: true,
		},
		"le_mask": resourceSchema.Int64Attribute{
			MarkdownDescription: "Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. " +
				"Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be " +
				"an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must " +
				"be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then " +
				"`le_mask` must be greater than `ge_mask`.",
			Optional: true,
		},
		"action": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("If the action is %q, match the route. If the action is %q, do "+
				"not match the route. For composing complex policies, all prefix-list items will be processed in the "+
				"order specified, top-down. This allows the user to deny a subset of a route that may otherwise be "+
				"permitted.", goapstra.PrefixFilterActionPermit, goapstra.PrefixFilterActionDeny),
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllValidPrefixFilterActions()...)},
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
