package apstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func stringUseStateForUnknownNull() planmodifier.String {
	return useStateForUnknownNullModifier{}
}

// useStateForUnknownNullModifier implements the plan modifier.
type useStateForUnknownNullModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (r useStateForUnknownNullModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it is Null."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r useStateForUnknownNullModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it is Null."
}

// PlanModifyString implements the plan modification logic.
func (r useStateForUnknownNullModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

//var _ tfsdk.AttributePlanModifier = tagDataTrackTagIdsModifier{}
//
//func tagDataTrackTagIds() tfsdk.AttributePlanModifier {
//	return tagDataTrackTagIdsModifier{}
//}
//
//type tagDataTrackTagIdsModifier struct{}
//
//// Modify doesn't work currently. The intent:
//// - It should be called on any tag_data element (set of tag_data objects)
//// - It should find the state and plan of it's sibling tag_ids object
//// - if state.tag_ids == plan.tag_ids : resp.AttributePlan = req.AttributeState
////   (use state)
//// - otherwise return having done nothing. tag_data will be left "unknown",
////   allowing it to be computed by the provider.
////
//// The problem: can't figure out how to grab the plan and state for the sibling
//// tag_ids, so cannot compare them.
//// options:
//// - figure out how to find ../tag_ids in the plan and state
//// - get ../tag_ids from the plan, fetch tag_data from the API
//// - switch tagged objects from sets to lists to simplify the path
////   string. JP suggests this should be safe to do:
////     Yes most API should be ordering things reliably and consistently, but
////     the rules for those sort orders can be unique to those api endpoints,
////     semantics may be different for some list elements across the product
////     This is a pretty important part of UI/UX experience, and even internal
////     to unit test framework where we have to have determinism in the API
////     response to embed those unit test payloads reliably
//func (r tagDataTrackTagIdsModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
//	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
//		return
//	}
//	return
//
//	plan := &rRackType{}
//	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	state := &rRackType{}
//	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	//tagDataPath := req.AttributePath
//	//parentPath := tagDataPath.ParentPath()
//	//tagIdsPath := parentPath.AtName("tag_ids")
//	//resp.Diagnostics.AddWarning("tag data path", tagDataPath.String())
//	//resp.Diagnostics.AddWarning("parent path", parentPath.String())
//	//resp.Diagnostics.AddWarning("tag ids path", tagIdsPath.String())
//	//resp.Diagnostics.AddWarning("state leaf_switches", state.LeafSwitches.String())
//	//resp.Diagnostics.AddWarning("plan leaf_switches", plan.LeafSwitches.String())
//
//	//thisObjByPath := types.Set{}
//	//resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, tagDataPath, &thisObjByPath)...)
//	//resp.Diagnostics.AddWarning("this object by path", thisObjByPath.String())
//
//	//planParent := types.Object{}
//	//resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, parentPath, &planParent)...)
//	//if resp.Diagnostics.HasError() {
//	//	return
//	//}
//	//resp.Diagnostics.AddWarning("parent", planParent.String())
//
//	return
//	var lsp, lss types.Set
//	req.State.Get(ctx, &lss)
//	req.Plan.Get(ctx, &lsp)
//	//resp.Diagnostics.AddWarning("state", lss.String())
//	//resp.Diagnostics.AddWarning("plan", lsp.String())
//	resp.Diagnostics.AddWarning("state", req.AttributeState.String())
//	resp.Diagnostics.AddWarning("plan", req.AttributePlan.String())
//	return
//
//	//resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, parentPath, types.Obj)...)
//	if resp.Diagnostics.HasError() {
//		resp.Diagnostics.AddWarning("welp ls", "that blew up in my face")
//		return
//	}
//	//dump, _ := json.MarshalIndent(ls, "", "  ")
//	//resp.Diagnostics.AddWarning("ls", string(dump))
//	//resp.Diagnostics.AddWarning("pah", parentPath.String())
//
//	// set tagData to "unknown"
//	resp.AttributePlan = types.SetUnknown(tagData{}.attrType())
//	return
//
//	// determine the path of 'tag_data' and 'tag_ids'
//	//tagDataPlanPath := req.AttributePath
//	//tagIdsPlanPath := tagDataPlanPath.ParentPath().AtName("tag_ids")
//
//	//var planTagIdStrings, stateTagIdStrings []string
//	//resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, tagIdsPlanPath, &planTagIdStrings)...)
//	//resp.Diagnostics.Append(req.State.GetAttribute(ctx, tagIdsPlanPath, &stateTagIdStrings)...)
//	//if resp.Diagnostics.HasError() {
//	//	resp.Diagnostics.AddWarning("welp strings", "that blew up in my face")
//	//	return
//	//}
//	//resp.Diagnostics.AddWarning("planTagIdStrings", strings.Join(planTagIdStrings, ","))
//	//resp.Diagnostics.AddWarning("stateTagIdStrings", strings.Join(stateTagIdStrings, ","))
//
//	//resp.Diagnostics.AddWarning("tagDataPlanPath", "\n"+req.AttributePath.String())
//	//resp.Diagnostics.AddWarning("tagIdsPlanPath", "\n"+tagIdsPlanPath.String())
//	//
//	//resp.Diagnostics.AddWarning("attribute state", req.AttributeState.String())
//
//	//resp.Diagnostics.AddWarning("req.AttributePath.String()", req.AttributePath.String())
//
//	// if tag_ids have no planned change, use state for tag_data
//	//if setStringEqual(planTagIdStrings, stateTagIdStrings) {
//	//	resp.AttributePlan = req.AttributeState
//	//	return
//	//}
//
//	// tag_ids have a planned change, so tag_data becomes "unknown"
//	//resp.Diagnostics.AddWarning("tag_data update", "setting tag_data to unknown")
//	unknownSet := types.SetUnknown(tagData{}.attrType())
//
//	resp.AttributePlan = unknownSet
//}
//
//// Description returns a human-readable description of the plan modifier.
//func (r tagDataTrackTagIdsModifier) Description(ctx context.Context) string {
//	return "Once set, the value of this attribute in state will not change, even if it is Null."
//}
//
//// MarkdownDescription returns a markdown description of the plan modifier.
//func (r tagDataTrackTagIdsModifier) MarkdownDescription(ctx context.Context) string {
//	return "Once set, the value of this attribute in state will not change, even if it is Null."
//}
//
////
