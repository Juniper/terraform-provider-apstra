package apstravalidator

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = LeafSwitchMlagInfoValidator{}

type LeafSwitchMlagInfoValidator struct{}

func (o LeafSwitchMlagInfoValidator) Description(_ context.Context) string {
	return "Validates sibling attributes 'mlag_info' and 'redundancy_mode' are aligned."
}

func (o LeafSwitchMlagInfoValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o LeafSwitchMlagInfoValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	mlagInfoPath := req.Path.ParentPath().AtName("mlag_info")
	var miObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, mlagInfoPath, &miObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !miObj.IsNull() && req.ConfigValue.ValueString() != apstra.LeafRedundancyProtocolMlag.String() {
		resp.Diagnostics.AddAttributeError(req.Path, errInvalidConfig,
			fmt.Sprintf("setting '%s' at '%s' is incompatible with '%s'",
				req.ConfigValue.ValueString(), req.Path.String(), mlagInfoPath.String()))
		return
	}

	if miObj.IsNull() && req.ConfigValue.ValueString() == apstra.LeafRedundancyProtocolMlag.String() {
		resp.Diagnostics.AddAttributeError(req.Path, errInvalidConfig,
			fmt.Sprintf("setting '%s' at '%s' requires setting '%s'",
				req.ConfigValue.ValueString(), req.Path.String(), mlagInfoPath.String()))
	}
}

func ValidateLeafSwitchRedundancyMode() validator.String {
	return LeafSwitchMlagInfoValidator{}
}
