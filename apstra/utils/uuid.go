package utils

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewUuidStringVal(diags *diag.Diagnostics) types.String {
	id, err := uuid.NewUUID()
	if err != nil {
		diags.AddError("failed to generate UUID", err.Error())
		return types.StringNull()
	}

	return types.StringValue(id.String())
}
