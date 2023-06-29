package apstravalidator

import (
	"context"
	"fmt"
	"github.com/IBM/netaddr"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

var _ validator.String = fallsWithinCidrValidator{}

type fallsWithinCidrValidator struct {
	expression path.Expression
	allZerosOk bool
	allOnesOk  bool
}

func (o fallsWithinCidrValidator) Description(_ context.Context) string {
	switch {
	case o.allZerosOk && o.allOnesOk:
		return fmt.Sprintf("Ensures that the supplied IP address falls "+
			"within the CIDR block specified at attribute %q.",
			o.expression.Resolve().String())
	case o.allZerosOk:
		return fmt.Sprintf("Ensures that the supplied IP address falls "+
			"within the CIDR block specified at attribute %q. and is not the "+
			"\"all ones\" address.", o.expression.Resolve().String())
	case o.allOnesOk:
		return fmt.Sprintf("Ensures that the supplied IP address falls "+
			"within the CIDR block specified at attribute %q. and is not the "+
			"\"all zeros\" address.", o.expression.Resolve().String())
	default:
		return fmt.Sprintf("Ensures that the supplied IP address falls "+
			"within the CIDR block specified at attribute %q. and is neither "+
			" the \"all zeros\" address nor the \"all ones\" address.",
			o.expression.Resolve().String())
	}
}

func (o fallsWithinCidrValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o fallsWithinCidrValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(o.expression)
	for i := range expressions {
		expressions[i] = expressions[i].Resolve()
	}

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, mp := range matchedPaths {
			var mpVal attr.Value
			diags = req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			var allZeros net.IP
			var subnet *net.IPNet
			var err error
			if mpString, ok := mpVal.(types.String); ok {
				allZeros, subnet, err = net.ParseCIDR(mpString.ValueString())
				if err != nil {
					resp.Diagnostics.AddAttributeError(
						mp, fmt.Sprintf("error parsing CIDR block %q",
							mpString.ValueString()), err.Error(),
					)
					return
				}
			} else {
				resp.Diagnostics.Append(validatordiag.BugInProviderDiagnostic(
					fmt.Sprintf("attribute at %q must be types.String got %t",
						o.expression, mpVal)))
				return
			}

			ip := net.ParseIP(req.ConfigValue.ValueString())
			if !subnet.Contains(ip) {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
					req.Path,
					fmt.Sprintf("value must fall within %s", subnet.String()),
					req.ConfigValue.ValueString()))
				return
			}

			if ip.Equal(allZeros) && !o.allZerosOk {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
					req.Path,
					"value must not be the all-zeros address %s",
					req.ConfigValue.ValueString()))
				return
			}

			allOnes := netaddr.BroadcastAddr(subnet)
			if ip.Equal(allOnes) && !o.allOnesOk {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
					req.Path,
					"value must not be the all-ones address %s",
					req.ConfigValue.ValueString()))
				return
			}
		}
	}
}

// FallsWithinCidr determines whether this attribute's value falls within the
// CIDR block specified by the attribute at expression. Arguments allZerosOk and
// allOnesOk modify the notion of "within" to include (true) or exclude (false)
// the first (all zeros) and last (all ones) addresses in the block.
func FallsWithinCidr(e path.Expression, allZerosOk bool, allOnesOk bool) validator.String {
	return fallsWithinCidrValidator{
		expression: e,
		allZerosOk: allZerosOk,
		allOnesOk:  allOnesOk,
	}
}
