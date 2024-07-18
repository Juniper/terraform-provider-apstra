package utils

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FabricAddressing(_ context.Context, s types.String, path *path.Path, diags *diag.Diagnostics) *apstra.AddressingScheme {
	if !HasValue(s) {
		return nil
	}

	var result apstra.AddressingScheme
	err := result.FromString(s.ValueString())
	if err != nil {
		if path != nil {
			diags.AddAttributeError(
				*path,
				fmt.Sprintf("failed parsing addressing scheme %q", s.ValueString()),
				err.Error(),
			)
		} else {
			diags.AddError(fmt.Sprintf(
				"failed parsing addressing scheme %q", s.ValueString()),
				err.Error(),
			)
		}
	}

	return &result
}
