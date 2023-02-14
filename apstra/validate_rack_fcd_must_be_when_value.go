package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ validator.String = rackFabricConnectivityDesignMustBeWhenValueValidator{}
)

type rackFabricConnectivityDesignMustBeWhenValueValidator struct {
	fcd   goapstra.FabricConnectivityDesign
	value string
}

type rackFabricConnectivityDesignMustBeWhenValueValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type rackFabricConnectivityDesignMustBeWhenValueValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o rackFabricConnectivityDesignMustBeWhenValueValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that when this value is %q, the parent 'rack_type' has 'fabric_connectivity_design' = %q", o.value, o.fcd.String())
}

func (o rackFabricConnectivityDesignMustBeWhenValueValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o rackFabricConnectivityDesignMustBeWhenValueValidator) Validate(ctx context.Context, req rackFabricConnectivityDesignMustBeWhenValueValidatorRequest, resp *rackFabricConnectivityDesignMustBeWhenValueValidatorResponse) {
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

func (o rackFabricConnectivityDesignMustBeWhenValueValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := rackFabricConnectivityDesignMustBeWhenValueValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &rackFabricConnectivityDesignMustBeWhenValueValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func stringFabricConnectivityDesignMustBeWhenValue(fcd goapstra.FabricConnectivityDesign, value string) validator.String {
	return rackFabricConnectivityDesignMustBeWhenValueValidator{
		fcd:   fcd,
		value: value,
	}
}
