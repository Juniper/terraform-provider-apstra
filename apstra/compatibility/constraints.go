package compatibility

import (
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/chrismarget-j/version-constraints"
)

var (
	BpIbaDashboardOk = versionconstraints.New(apiversions.LtApstra500)
	BpIbaProbeOk     = versionconstraints.New(apiversions.LtApstra500)
	BpIbaWidgetOk    = versionconstraints.New(apiversions.LtApstra500)
)
