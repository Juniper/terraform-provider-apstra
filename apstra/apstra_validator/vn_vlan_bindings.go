package apstravalidator

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Set = VnVlanBindingsValidator{}

type VnVlanBindingsValidator struct{}

func (o VnVlanBindingsValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensure that when the VN type is %q, only a single binding is configured", apstra.VnTypeVlan)
}

func (o VnVlanBindingsValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o VnVlanBindingsValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}

	typePath := req.Path.ParentPath().AtName("type")

	var typeVal types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, typePath, &typeVal)...)

	isVlanType := typeVal.ValueString() == apstra.VnTypeVlan.String()
	hasMultipleBindings := len(req.ConfigValue.Elements()) > 1

	if isVlanType && hasMultipleBindings {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("only 1 binding permitted when virtual network type is %q, got %d", typeVal, len(req.ConfigValue.Elements())),
		))
	}
}

func OneBindingWhenVlan() validator.Set {
	return VnVlanBindingsValidator{}
}
