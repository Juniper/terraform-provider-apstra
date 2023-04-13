package apstraplanmodifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ NineTypesPlanModifier = &UnknownSetter{}

type NineTypesPlanModifier interface {
	planmodifier.String
	planmodifier.Bool
	planmodifier.Float64
	planmodifier.Int64
	planmodifier.List
	planmodifier.Map
	planmodifier.Number
	planmodifier.Object
	planmodifier.Set
	planmodifier.String
}

type UnknownSetter struct{}

// based on a comment by @SBGoods:
//
//	The subject of what goes on behind the scenes of Terraform plan with
//	regards to providers is pretty nuanced. Without going too much into the
//	weeds, the behavior for Terraform for Optional + Computed attributes is to
//	copy the prior state if there is no configuration for it.
//
//	To handle your situation in the framework, you can create an attribute or
//	resource plan modifier as detailed in my previous reply except you would
//	set the plan value to unknown if the request state value is not null and
//	the request configuration value is null. This should then give you a diff
//	in the terraform plan.
//
// https://discuss.hashicorp.com/t/schema-for-optional-computed-to-support-correct-removal-plan-in-framework/49055/5?u=hqnvylrx
func (o UnknownSetter) removed(configValue attr.Value, stateValue attr.Value) bool {
	if !stateValue.IsNull() && configValue.IsNull() {
		return true
	}
	return false
}

func (o UnknownSetter) Description(_ context.Context) string {
	return "value reverts to <unknown> when removed from config"
}

func (o UnknownSetter) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o UnknownSetter) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.BoolUnknown()
	}
}

func (o UnknownSetter) PlanModifyFloat64(_ context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.Float64Unknown()
	}
}

func (o UnknownSetter) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.Int64Unknown()
	}
}

func (o UnknownSetter) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.ListUnknown(req.ConfigValue.Type(ctx))
	}
}

func (o UnknownSetter) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.MapUnknown(req.ConfigValue.Type(ctx))
	}
}

func (o UnknownSetter) PlanModifyNumber(_ context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.NumberUnknown()
	}
}

func (o UnknownSetter) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.ObjectUnknown(req.ConfigValue.AttributeTypes(ctx))
	}
}

func (o UnknownSetter) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.SetUnknown(req.ConfigValue.Type(ctx))
	}
}

func (o UnknownSetter) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if o.removed(req.ConfigValue, req.StateValue) {
		resp.PlanValue = types.StringUnknown()
	}
}

// UnknownWhenRemoved is useful for Optional + Computed attributes. The default
// behavior for these attributes is for them to *remain set* in the plan when
// they're removed from the configuration, which is probably not what the
// Terraform user intended. This plan modifier reverts them to <unknown> in
// that case, which is probably closer to the expected behavior.
func UnknownWhenRemoved() NineTypesPlanModifier {
	return UnknownSetter{}
}
