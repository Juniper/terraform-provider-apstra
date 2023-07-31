package apstraplanmodifier

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func BoolRequiresReplaceWhenSwitchingTo(to bool) boolplanmodifier.RequiresReplaceIfFunc {
	return func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
		if req.StateValue.IsUnknown() || req.StateValue.IsNull() {
			// if no prior state, there's no way to determine whether it's being changed
			return
		}

		if req.PlanValue.ValueBool() != to {
			// plan doesn't match the "switching to" trigger
			return
		}

		// plan does match the "switching to" trigger
		if req.StateValue.ValueBool() != to {
			// state does not mach the "switching to" trigger. We're switching to the trigger value.
			resp.RequiresReplace = true
		}
	}
}

//func requiresReplaceWhenSwitchingTo() func(context.Context, planmodifier.BoolRequest, *RequiresReplaceIfFuncResponse)
