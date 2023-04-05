package apstravalidator

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ validator.String = RackFabricConnectivityDesignMustBeValidator{}
	_ validator.Int64  = RackFabricConnectivityDesignMustBeValidator{}
)

type RackFabricConnectivityDesignMustBeValidator struct {
	fcd apstra.FabricConnectivityDesign
}

type RackFabricConnectivityDesignMustBeValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type RackFabricConnectivityDesignMustBeValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o RackFabricConnectivityDesignMustBeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that the parent 'rack_type' has 'fabric_connectivity_design' %q", o.fcd.String())
}

func (o RackFabricConnectivityDesignMustBeValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o RackFabricConnectivityDesignMustBeValidator) Validate(ctx context.Context, req RackFabricConnectivityDesignMustBeValidatorRequest, resp *RackFabricConnectivityDesignMustBeValidatorResponse) {
	// don't force fabric connectivity design when values are null or unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
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
			fmt.Sprintf("%q valid only when %q has value %q", req.Path, fcdPath, o.fcd.String())))
	}
}

func (o RackFabricConnectivityDesignMustBeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := RackFabricConnectivityDesignMustBeValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &RackFabricConnectivityDesignMustBeValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RackFabricConnectivityDesignMustBeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := RackFabricConnectivityDesignMustBeValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &RackFabricConnectivityDesignMustBeValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func Int64FabricConnectivityDesignMustBe(fcd apstra.FabricConnectivityDesign) validator.Int64 {
	return RackFabricConnectivityDesignMustBeValidator{
		fcd: fcd,
	}
}

func StringFabricConnectivityDesignMustBe(fcd apstra.FabricConnectivityDesign) validator.String {
	return RackFabricConnectivityDesignMustBeValidator{
		fcd: fcd,
	}
}
