package apstra

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = switchLagInfoValidator{}

type switchLagInfoValidator struct {
	redundancyProtocol string
}

func (o switchLagInfoValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures the sibling attribute 'redundancy_protocol' is set to '%s'.", o.redundancyProtocol)
}

func (o switchLagInfoValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o switchLagInfoValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	redundancyProtocolPath := req.Path.ParentPath().AtName("redundancy_protocol")

	var redundancyProtocol types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, redundancyProtocolPath, &redundancyProtocol)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if redundancyProtocol.IsNull() || redundancyProtocol.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			req.Path, errInvalidConfig,
			fmt.Sprintf("configuring '%s' requires '%s' = '%s'", req.Path.String(), redundancyProtocolPath.String(), o.redundancyProtocol),
		)
		return
	}

	if redundancyProtocol.ValueString() != o.redundancyProtocol {
		resp.Diagnostics.AddAttributeError(
			req.Path, errInvalidConfig,
			fmt.Sprintf("configuring '%s' requires '%s' = '%s'", req.Path.String(), redundancyProtocolPath.String(), o.redundancyProtocol),
		)
	}
}

func validateSwitchLagInfo(m string) validator.Object {
	return switchLagInfoValidator{
		redundancyProtocol: m,
	}
}
