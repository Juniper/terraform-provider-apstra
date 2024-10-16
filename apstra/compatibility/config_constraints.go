package compatibility

import (
	"context"
	"fmt"

	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

const (
	errSummary         = "Configuration not compatible with Apstra %s"
	errAttributeDetail = "This configuration requires Apstra %s"
	errOtherDetail     = "%s\nThis configuration requires Apstra %s"
)

type AttributeConstraint struct {
	Path        path.Path
	Constraints versionconstraints.Constraints
}

type OtherConstraint struct {
	Message     string
	Constraints versionconstraints.Constraints
}

type ConfigConstraints struct {
	attributeConstraints []AttributeConstraint
	otherConstraints     []OtherConstraint
}

// AddAttributeConstraints should be used to add version constraints imposed by
// the presence of an attribute or attribute value in the configuration.
func (o *ConfigConstraints) AddAttributeConstraints(in AttributeConstraint) {
	o.attributeConstraints = append(o.attributeConstraints, in)
}

// AddOtherConstraints should be used to add non-attribute-value-based version
// constraints, like those imposed by the absence of an attribute in the
// configuration.
func (o *ConfigConstraints) AddOtherConstraints(in OtherConstraint) {
	o.otherConstraints = append(o.otherConstraints, in)
}

func (o *ConfigConstraints) AttributeConstraints() []AttributeConstraint {
	return o.attributeConstraints
}

func (o *ConfigConstraints) OtherConstraints() []OtherConstraint {
	return o.otherConstraints
}

type ValidateConfigConstraintsRequest struct {
	Version     *version.Version
	Constraints ConfigConstraints
}

func ValidateConfigConstraints(_ context.Context, req ValidateConfigConstraintsRequest) diag.Diagnostics {
	var response diag.Diagnostics

	for _, constraint := range req.Constraints.attributeConstraints {
		if !constraint.Constraints.Check(req.Version) { // un-met version constraint?
			response.AddAttributeError(
				constraint.Path,
				fmt.Sprintf(errSummary, req.Version),
				fmt.Sprintf(errAttributeDetail, constraint.Constraints),
			)
		}
	}

	for _, constraint := range req.Constraints.otherConstraints {
		if !constraint.Constraints.Check(req.Version) {
			response.AddError(
				fmt.Sprintf(errSummary, req.Version),
				fmt.Sprintf(errOtherDetail, constraint.Message, constraint.Constraints),
			)
		}
	}

	return response
}
