package tfapstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	ResourceAgentProfile            = resourceAgentProfile{}
	ResourceDatacenterGenericSystem = resourceDatacenterGenericSystem{}
	ResourceDatacenterRoutingZone   = resourceDatacenterRoutingZone{}
	ResourceFreeformConfigTemplate  = resourceFreeformConfigTemplate{}
	ResourceFreeformLink            = resourceFreeformLink{}
	ResourceFreeformSystem          = resourceFreeformSystem{}
	ResourceFreeformPropertySet     = resourceFreeformPropertySet{}
	ResourceIpv4Pool                = resourceIpv4Pool{}
	ResourceTemplatePodBased        = resourceTemplatePodBased{}
	ResourceTemplateCollapsed       = resourceTemplateCollapsed{}
)

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
