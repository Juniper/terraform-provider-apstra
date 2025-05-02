package apstravalidator

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ validator.String = RackFabricConnectivityDesignMustBeWhenNullValidator{}
	_ validator.Int64  = RackFabricConnectivityDesignMustBeWhenNullValidator{}
)

type RackFabricConnectivityDesignMustBeWhenNullValidator struct {
	fcd enum.FabricConnectivityDesign
}

type RackFabricConnectivityDesignMustBeWhenNullValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type RackFabricConnectivityDesignMustBeWhenNullValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o RackFabricConnectivityDesignMustBeWhenNullValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that when this value is null, the parent 'rack_type' has 'fabric_connectivity_design' %q", o.fcd)
}

func (o RackFabricConnectivityDesignMustBeWhenNullValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o RackFabricConnectivityDesignMustBeWhenNullValidator) Validate(ctx context.Context, req RackFabricConnectivityDesignMustBeWhenNullValidatorRequest, resp *RackFabricConnectivityDesignMustBeWhenNullValidatorResponse) {
	// do nothing if the value is known
	if !req.ConfigValue.IsNull() {
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
			fmt.Sprintf("%q must be %q when %q is null", fcdPath, o.fcd.String(), req.Path)))
	}
}

func (o RackFabricConnectivityDesignMustBeWhenNullValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := RackFabricConnectivityDesignMustBeWhenNullValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &RackFabricConnectivityDesignMustBeWhenNullValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RackFabricConnectivityDesignMustBeWhenNullValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := RackFabricConnectivityDesignMustBeWhenNullValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &RackFabricConnectivityDesignMustBeWhenNullValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func Int64FabricConnectivityDesignMustBeWhenNull(fcd enum.FabricConnectivityDesign) validator.Int64 {
	return RackFabricConnectivityDesignMustBeWhenNullValidator{
		fcd: fcd,
	}
}

func StringFabricConnectivityDesignMustBeWhenNull(fcd enum.FabricConnectivityDesign) validator.String {
	return RackFabricConnectivityDesignMustBeWhenNullValidator{
		fcd: fcd,
	}
}
