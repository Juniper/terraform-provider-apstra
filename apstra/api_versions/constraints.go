package apiversions

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type attributeConstraints struct {
	path        path.Path
	constraints version.Constraints
}

type otherConstraint struct {
	message     string
	constraints version.Constraints
}

type Constraints struct {
	attributeConstraints []attributeConstraints
	otherConstraints     []otherConstraint
}

// AddAttributeConstraints should be used to add version constraints imposed by
// the presence of an attribute or attribute value in the configuration.
func (o *Constraints) AddAttributeConstraints(path path.Path, constraints version.Constraints) {
	o.attributeConstraints = append(o.attributeConstraints, attributeConstraints{
		path:        path,
		constraints: constraints,
	})
}

// AddConstraints should be used to add non-attribute-value-based version
// constraints, like those imposed by the absence of an attribute in the
// configuration.
func (o *Constraints) AddConstraints(message string, constraints version.Constraints) {
	o.otherConstraints = append(o.otherConstraints, otherConstraint{
		message:     message,
		constraints: constraints,
	})
}
