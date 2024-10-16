package apiversions

import (
	"github.com/hashicorp/go-version"
)

var (
	Eq410 = version.MustConstraints(version.NewConstraint(Apstra410))
	Ge411 = version.MustConstraints(version.NewConstraint(">=" + Apstra411))
	Ge420 = version.MustConstraints(version.NewConstraint(">=" + Apstra420))
	Ge421 = version.MustConstraints(version.NewConstraint(">=" + Apstra421))
)
