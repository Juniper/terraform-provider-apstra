package customtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ attr.Type            = (*SetWithSemanticEqualsType)(nil)
	_ basetypes.SetTypable = (*SetWithSemanticEqualsType)(nil)
)

type SetWithSemanticEqualsType struct {
	basetypes.SetType
}

func NewSetWithSemanticEqualsType(elemType attr.Type) basetypes.SetTypable {
	result := SetWithSemanticEqualsType{}
	result.ElemType = elemType
	return result
}
