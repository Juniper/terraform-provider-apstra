package compatibility

import (
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/chrismarget-j/version-constraints"
)

var (
	ApiNotSupportsSetLoopbackIps       = versionconstraints.New(apiversions.LtApstra500)
	BPDefaultRoutingZoneAddressingOK   = versionconstraints.New(apiversions.GeApstra610)
	BpIbaDashboardOk                   = versionconstraints.New(apiversions.LtApstra500)
	BpIbaProbeOk                       = versionconstraints.New(apiversions.LtApstra500)
	BpIbaWidgetOk                      = versionconstraints.New(apiversions.LtApstra500)
	BlueprintIPv6ApplicationsOK        = versionconstraints.New(apiversions.LtApstra610)
	ChangeVnRzIdForbidden              = versionconstraints.New(apiversions.LeApstra422)
	FabricSettingsSetInCreate          = versionconstraints.New(apiversions.GeApstra421)
	RoutingPolicyExportL3EdgeServerOK  = versionconstraints.New(apiversions.LeApstra422)
	TemplateRequiresAntiAffinityPolicy = versionconstraints.New(apiversions.Apstra420)
	VnDescriptionOk                    = versionconstraints.New(apiversions.GeApstra500)
	VnEmptyBindingsOk                  = versionconstraints.New(apiversions.GeApstra500)
	VnTagsOk                           = versionconstraints.New(apiversions.GeApstra500)
)
