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
	_ validator.String = rackFabricConnectivityDesignMustBeValidator{}
	_ validator.Int64  = rackFabricConnectivityDesignMustBeValidator{}
)

type rackFabricConnectivityDesignMustBeValidator struct {
	fcd goapstra.FabricConnectivityDesign
}

type rackFabricConnectivityDesignMustBeValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type rackFabricConnectivityDesignMustBeValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o rackFabricConnectivityDesignMustBeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that the parent 'rack_type' has 'fabric_connectivity_design' %q", o.fcd.String())
}

func (o rackFabricConnectivityDesignMustBeValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o rackFabricConnectivityDesignMustBeValidator) Validate(ctx context.Context, req rackFabricConnectivityDesignMustBeValidatorRequest, resp *rackFabricConnectivityDesignMustBeValidatorResponse) {
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

func (o rackFabricConnectivityDesignMustBeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := rackFabricConnectivityDesignMustBeValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &rackFabricConnectivityDesignMustBeValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o rackFabricConnectivityDesignMustBeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := rackFabricConnectivityDesignMustBeValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &rackFabricConnectivityDesignMustBeValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func int64FabricConnectivityDesignMustBe(fcd goapstra.FabricConnectivityDesign) validator.Int64 {
	return rackFabricConnectivityDesignMustBeValidator{
		fcd: fcd,
	}
}

func stringFabricConnectivityDesignMustBe(fcd goapstra.FabricConnectivityDesign) validator.String {
	return rackFabricConnectivityDesignMustBeValidator{
		fcd: fcd,
	}
}
