package apiversions

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	errSummary         = "Configuration not valid with Apstra %s"
	errAttributeDetail = "This configuration requires Apstra %s"
	errOtherDetail     = "%s\nThis configuration requires Apstra %s"
)

type ValidateConstraintsRequest struct {
	Version     *version.Version
	Constraints Constraints
}

func ValidateConstraints(_ context.Context, req ValidateConstraintsRequest) diag.Diagnostics {
	var response diag.Diagnostics

	for _, constraint := range req.Constraints.attributeConstraints {
		if !constraint.constraints.Check(req.Version) { // un-met version constraint?
			response.AddAttributeError(
				constraint.path,
				fmt.Sprintf(errSummary, req.Version),
				fmt.Sprintf(errAttributeDetail, constraint.constraints),
			)
		}
	}

	for _, constraint := range req.Constraints.otherConstraints {
		if !constraint.constraints.Check(req.Version) {
			response.AddError(
				fmt.Sprintf(errSummary, req.Version),
				fmt.Sprintf(errOtherDetail, constraint.message, constraint.constraints),
			)
		}
	}

	return response
}
