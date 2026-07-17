package customtypes

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringValuable    = (*RouteTarget)(nil)
	_ xattr.ValidateableAttribute = (*RouteTarget)(nil)
)

type RouteTarget struct {
	basetypes.StringValue
}

func (v RouteTarget) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if !v.Type(ctx).TerraformType(ctx).Is(tftypes.String) {
		err := fmt.Errorf("expected String value, received %T with value: %v", v, v)
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"RouteTarget Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)
	}

	if v.IsNull() || v.IsUnknown() {
		return
	}

	_, err := datacenter.NewRouteTarget(v.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Route Target String Value",
			"A string value was provided that is not a valid Route Type string format.\n\n"+
				"Given Value: "+v.ValueString()+"\n"+
				"Error: "+err.Error()+"\n",
		)
	}
}

func (v RouteTarget) Type(_ context.Context) attr.Type {
	return RouteTargetType{}
}

func (v RouteTarget) Equal(o attr.Value) bool {
	other, ok := o.(RouteTarget)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func NewRouteTargetNull() RouteTarget {
	return RouteTarget{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewRouteTargetUnknown() RouteTarget {
	return RouteTarget{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewRouteTargetValue(value string) RouteTarget {
	return RouteTarget{
		StringValue: basetypes.NewStringValue(value),
	}
}

func NewRouteTargetPointerValue(value *string) RouteTarget {
	return RouteTarget{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}
