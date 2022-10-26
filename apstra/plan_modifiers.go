package apstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ tfsdk.AttributePlanModifier = useStateForUnknownNullModifier{}

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

//

var _ tfsdk.AttributePlanModifier = tagDataTrackTagIdsModifier{}

func tagDataTrackTagIds() tfsdk.AttributePlanModifier {
	return tagDataTrackTagIdsModifier{}
}

type tagDataTrackTagIdsModifier struct{}

func (r tagDataTrackTagIdsModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	tagDataPath := req.AttributePath
	tagIdsPath := tagDataPath.ParentPath().AtName("tag_ids")

	var planTagIds, stateTagIds types.Set
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, tagIdsPath, &planTagIds)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, tagIdsPath, &stateTagIds)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planTagIds.Equal(stateTagIds) {
		return
	}

	unknownSet := newTagDataSet(0)
	unknownSet.Unknown = true

	resp.AttributePlan = unknownSet
}

// Description returns a human-readable description of the plan modifier.
func (r tagDataTrackTagIdsModifier) Description(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it is Null."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r tagDataTrackTagIdsModifier) MarkdownDescription(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it is Null."
}

//
