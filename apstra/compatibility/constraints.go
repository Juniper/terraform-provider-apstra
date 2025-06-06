package compatibility

import (
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/chrismarget-j/version-constraints"
)

var (
	ApiNotSupportsSetLoopbackIps       = versionconstraints.New(apiversions.LtApstra500)
	BpIbaDashboardOk                   = versionconstraints.New(apiversions.LtApstra500)
	BpIbaProbeOk                       = versionconstraints.New(apiversions.LtApstra500)
	BpIbaWidgetOk                      = versionconstraints.New(apiversions.LtApstra500)
	ChangeVnRzIdForbidden              = versionconstraints.New(apiversions.LeApstra422)
	FabricSettingsSetInCreate          = versionconstraints.New(apiversions.GeApstra421)
	TemplateRequiresAntiAffinityPolicy = versionconstraints.New(apiversions.Apstra420)
	VnDescriptionOk                    = versionconstraints.New(apiversions.GeApstra500)
	VnEmptyBindingsOk                  = versionconstraints.New(apiversions.GeApstra500)
	VnTagsOk                           = versionconstraints.New(apiversions.GeApstra500)
)
