// This file based on hashicorp/terraform-plugin-framework v1.14.1 types/basetypes/set_type.go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.SetTypable = (*SetWithSemanticEqualsType)(nil)

type SetWithSemanticEqualsType struct {
	basetypes.SetType
}

func (o SetWithSemanticEqualsType) Equal(in attr.Type) bool {
	other, ok := in.(SetWithSemanticEqualsType)
	if !ok {
		return false
	}

	return o.ElementType().Equal(other.ElementType())
}

func (o SetWithSemanticEqualsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewSetWithSemanticEqualsNull(o.ElementType()), nil
	}

	if !in.Type().Equal(o.TerraformType(ctx)) {
		return nil, fmt.Errorf("can't use %s as value of Set with ElementType %T, can only use %s values", in.String(), o.ElementType(), o.ElementType().TerraformType(ctx).String())
	}
	if !in.IsKnown() {
		return NewSetWithSemanticEqualsUnknown(o.ElementType()), nil
	}
	if in.IsNull() {
		return NewSetWithSemanticEqualsNull(o.ElementType()), nil
	}
	val := []tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	elems := make([]attr.Value, 0, len(val))
	for _, elem := range val {
		av, err := o.ElementType().ValueFromTerraform(ctx, elem)
		if err != nil {
			return nil, err
		}
		elems = append(elems, av)
	}
	// ValueFromTerraform above on each element should make this safe.
	// Otherwise, this will need to do some Diagnostics to error conversion.
	return NewSetWithSemanticEqualsValueMust(o.ElementType(), elems), nil
}

func (o SetWithSemanticEqualsType) ValueType(_ context.Context) attr.Value {
	return SetWithSemanticEqualsValue{basetypes.NewSetNull(o.ElemType)}
}

func (o SetWithSemanticEqualsType) String() string {
	return "types.SetWithSemanticEqualsType[" + o.ElementType().String() + "]"
}

func (o SetWithSemanticEqualsType) ValueFromSet(_ context.Context, set basetypes.SetValue) (basetypes.SetValuable, diag.Diagnostics) {
	return SetWithSemanticEqualsValue{SetValue: set}, nil
}

func NewSetWithSemanticEqualsType(elemType attr.Type) basetypes.SetTypable {
	result := SetWithSemanticEqualsType{}
	result.ElemType = elemType
	return result
}
