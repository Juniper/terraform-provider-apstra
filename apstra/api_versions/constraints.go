package apiversions

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var Ge420 = version.MustConstraints(version.NewConstraint(">=" + Apstra420))

type AttributeConstraint struct {
	Path        path.Path
	Constraints version.Constraints
}

type OtherConstraint struct {
	Message     string
	Constraints version.Constraints
}

type Constraints struct {
	attributeConstraints []AttributeConstraint
	otherConstraints     []OtherConstraint
}

// AddAttributeConstraints should be used to add version constraints imposed by
// the presence of an attribute or attribute value in the configuration.
func (o *Constraints) AddAttributeConstraints(in AttributeConstraint) {
	o.attributeConstraints = append(o.attributeConstraints, in)
}

// AddOtherConstraints should be used to add non-attribute-value-based version
// constraints, like those imposed by the absence of an attribute in the
// configuration.
func (o *Constraints) AddOtherConstraints(in OtherConstraint) {
	o.otherConstraints = append(o.otherConstraints, in)
}

func (o *Constraints) AttributeConstraints() []AttributeConstraint {
	return o.attributeConstraints
}

func (o *Constraints) OtherConstraints() []OtherConstraint {
	return o.otherConstraints
}
