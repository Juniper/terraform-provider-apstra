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
	_ validator.String = rackFabricConnectivityDesignMustBeWhenNullValidator{}
	_ validator.Int64  = rackFabricConnectivityDesignMustBeWhenNullValidator{}
)

type rackFabricConnectivityDesignMustBeWhenNullValidator struct {
	fcd goapstra.FabricConnectivityDesign
}

type rackFabricConnectivityDesignMustBeWhenNullValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type rackFabricConnectivityDesignMustBeWhenNullValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o rackFabricConnectivityDesignMustBeWhenNullValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that when this value is null, the parent 'rack_type' has 'fabric_connectivity_design' %q", o.fcd.String())
}

func (o rackFabricConnectivityDesignMustBeWhenNullValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o rackFabricConnectivityDesignMustBeWhenNullValidator) Validate(ctx context.Context, req rackFabricConnectivityDesignMustBeWhenNullValidatorRequest, resp *rackFabricConnectivityDesignMustBeWhenNullValidatorResponse) {
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

func (o rackFabricConnectivityDesignMustBeWhenNullValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := rackFabricConnectivityDesignMustBeWhenNullValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &rackFabricConnectivityDesignMustBeWhenNullValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o rackFabricConnectivityDesignMustBeWhenNullValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := rackFabricConnectivityDesignMustBeWhenNullValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &rackFabricConnectivityDesignMustBeWhenNullValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func int64FabricConnectivityDesignMustBeWhenNull(fcd goapstra.FabricConnectivityDesign) validator.Int64 {
	return rackFabricConnectivityDesignMustBeWhenNullValidator{
		fcd: fcd,
	}
}

func stringFabricConnectivityDesignMustBeWhenNull(fcd goapstra.FabricConnectivityDesign) validator.String {
	return rackFabricConnectivityDesignMustBeWhenNullValidator{
		fcd: fcd,
	}
}
