package apstravalidator

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = RackFabricConnectivityDesignMustBeWhenValueValidator{}

type RackFabricConnectivityDesignMustBeWhenValueValidator struct {
	fcd   enum.FabricConnectivityDesign
	value string
}

type RackFabricConnectivityDesignMustBeWhenValueValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type RackFabricConnectivityDesignMustBeWhenValueValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o RackFabricConnectivityDesignMustBeWhenValueValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that when this value is %q, the parent 'rack_type' has 'fabric_connectivity_design' = %q", o.value, o.fcd.String())
}

func (o RackFabricConnectivityDesignMustBeWhenValueValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o RackFabricConnectivityDesignMustBeWhenValueValidator) Validate(ctx context.Context, req RackFabricConnectivityDesignMustBeWhenValueValidatorRequest, resp *RackFabricConnectivityDesignMustBeWhenValueValidatorResponse) {
	// do nothing if no value
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// do nothing if value isn't the "when" value
	if !req.ConfigValue.Equal(types.StringValue(o.value)) {
		return
	}

	fcdPath := req.Path.ParentPath().ParentPath().ParentPath().AtName("fabric_connectivity_design")

	var fcdVal attr.Value
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, fcdPath, &fcdVal)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !types.StringValue(o.fcd.String()).Equal(fcdVal) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(req.Path,
			fmt.Sprintf("%q must be %q when %q has value %q", fcdPath, o.fcd.String(), req.Path, o.value)))
	}
}

func (o RackFabricConnectivityDesignMustBeWhenValueValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := RackFabricConnectivityDesignMustBeWhenValueValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &RackFabricConnectivityDesignMustBeWhenValueValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func StringFabricConnectivityDesignMustBeWhenValue(fcd enum.FabricConnectivityDesign, value string) validator.String {
	return RackFabricConnectivityDesignMustBeWhenValueValidator{
		fcd:   fcd,
		value: value,
	}
}
