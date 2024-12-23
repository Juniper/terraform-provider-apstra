package tfapstra

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	DataSourceBlueprintNodeConfig   = dataSourceBlueprintNodeConfig{}
	DataSourceDatacenterSystemNodes = dataSourceDatacenterSystemNodes{}

	ResourceAgentProfile                            = resourceAgentProfile{}
	ResourceAsnPool                                 = resourceAsnPool{}
	ResourceDatacenterConnectivityTemplateInterface = resourceDatacenterConnectivityTemplateInterface{}
	ResourceDatacenterConnectivityTemplateLoopback  = resourceDatacenterConnectivityTemplateLoopback{}
	ResourceDatacenterConnectivityTemplateSvi       = resourceDatacenterConnectivityTemplateSvi{}
	ResourceDatacenterConnectivityTemplateSystem    = resourceDatacenterConnectivityTemplateSystem{}
	ResourceDatacenterGenericSystem                 = resourceDatacenterGenericSystem{}
	ResourceDatacenterIpLinkAddressing              = resourceDatacenterIpLinkAddressing{}
	ResourceDatacenterRoutingZone                   = resourceDatacenterRoutingZone{}
	ResourceDatacenterRoutingZoneConstraint         = resourceDatacenterRoutingZoneConstraint{}
	ResourceDatacenterVirtualNetwork                = resourceDatacenterVirtualNetwork{}
	ResourceFreeformAllocGroup                      = resourceFreeformAllocGroup{}
	ResourceFreeformBlueprint                       = resourceFreeformBlueprint{}
	ResourceFreeformConfigTemplate                  = resourceFreeformConfigTemplate{}
	ResourceFreeformDeviceProfile                   = resourceFreeformDeviceProfile{}
	ResourceFreeformGroupGenerator                  = resourceFreeformGroupGenerator{}
	ResourceFreeformLink                            = resourceFreeformLink{}
	ResourceFreeformPropertySet                     = resourceFreeformPropertySet{}
	ResourceFreeformResourceGenerator               = resourceFreeformResourceGenerator{}
	ResourceFreeformResourceGroup                   = resourceFreeformResourceGroup{}
	ResourceFreeformResource                        = resourceFreeformResource{}
	ResourceFreeformSystem                          = resourceFreeformSystem{}
	ResourceIntegerPool                             = resourceIntegerPool{}
	ResourceIpv4Pool                                = resourceIpv4Pool{}
	ResourceIpv6Pool                                = resourceIpv6Pool{}
	ResourceTelemetryServiceRegistryEntry           = resourceTelemetryServiceRegistryEntry{}
	ResourceTemplateCollapsed                       = resourceTemplateCollapsed{}
	ResourceTemplatePodBased                        = resourceTemplatePodBased{}
	ResourceVniPool                                 = resourceVniPool{}
)

func DatasourceName(ctx context.Context, d datasource.DataSource) string {
	var pMdReq provider.MetadataRequest
	var pMdResp provider.MetadataResponse
	NewProvider().Metadata(ctx, pMdReq, &pMdResp)

	var dMdReq datasource.MetadataRequest
	var dMdResp datasource.MetadataResponse

	dMdReq.ProviderTypeName = pMdResp.TypeName
	d.Metadata(ctx, dMdReq, &dMdResp)

	return dMdResp.TypeName
}

func ResourceName(ctx context.Context, r resource.Resource) string {
	var pMdReq provider.MetadataRequest
	var pMdResp provider.MetadataResponse
	NewProvider().Metadata(ctx, pMdReq, &pMdResp)

	var rMdReq resource.MetadataRequest
	var rMdResp resource.MetadataResponse

	rMdReq.ProviderTypeName = pMdResp.TypeName
	r.Metadata(ctx, rMdReq, &rMdResp)

	return rMdResp.TypeName
}
