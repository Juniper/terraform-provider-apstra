package apstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// useStateForUnknownNull is a clone of resource/UseStateForUnknown from
// github.com/hashicorp/terraform-plugin-framework. Where the original ignores
// null values, this resource modifier will copy null values from the state
// into the plan.
func useStateForUnknownNull() tfsdk.AttributePlanModifier {
	return useStateForUnknownNullModifier{}
}

// useStateForUnknownNullModifier implements the useStateForUnknownNull
// AttributePlanModifier.
type useStateForUnknownNullModifier struct{}

// Modify copies the attribute's prior state to the attribute plan.
func (r useStateForUnknownNullModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	// if it's not planned to be the unknown value, stick with the concrete plan
	if !resp.AttributePlan.IsUnknown() {
		return
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.AttributeConfig.IsUnknown() {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r useStateForUnknownNullModifier) Description(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it is Null."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r useStateForUnknownNullModifier) MarkdownDescription(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it is Null."
}
