package apstravalidator

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Map = ExactlyOneBindingWhenVnTypeVlanValidator{}

type ExactlyOneBindingWhenVnTypeVlanValidator struct{}

func (o ExactlyOneBindingWhenVnTypeVlanValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensure that when the VN type is %q, only a single binding is configured", enum.VnTypeVlan)
}

func (o ExactlyOneBindingWhenVnTypeVlanValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ExactlyOneBindingWhenVnTypeVlanValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}

	typePath := req.Path.ParentPath().AtName("type")

	var typeVal types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, typePath, &typeVal)...)

	isVlanType := typeVal.ValueString() == enum.VnTypeVlan.String()
	hasMultipleBindings := len(req.ConfigValue.Elements()) > 1

	if isVlanType && hasMultipleBindings {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("only 1 binding permitted when virtual network type is %q, got %d", typeVal, len(req.ConfigValue.Elements())),
		))
	}
}

// ExactlyOneBindingWhenVnTypeVlan ensures that exactly one leaf node binding
// exists when the attribute at the neighboring `type` attribute matches
// apstra.VnTypeVlan
func ExactlyOneBindingWhenVnTypeVlan() validator.Map {
	return ExactlyOneBindingWhenVnTypeVlanValidator{}
}
