package tfapstra

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	ResourceAgentProfile               = resourceAgentProfile{}
	ResourceAsnPool                    = resourceAsnPool{}
	ResourceDatacenterGenericSystem    = resourceDatacenterGenericSystem{}
	ResourceDatacenterIpLinkAddressing = resourceDatacenterIpLinkAddressing{}
	ResourceDatacenterRoutingZone      = resourceDatacenterRoutingZone{}
	ResourceFreeformBlueprint          = resourceFreeformBlueprint{}
	ResourceFreeformConfigTemplate     = resourceFreeformConfigTemplate{}
	ResourceFreeformLink               = resourceFreeformLink{}
	ResourceFreeformPropertySet        = resourceFreeformPropertySet{}
	ResourceFreeformResourceGroup      = resourceFreeformResourceGroup{}
	ResourceFreeformResource           = resourceFreeformResource{}
	ResourceFreeformSystem             = resourceFreeformSystem{}
	ResourceIntegerPool                = resourceIntegerPool{}
	ResourceIpv4Pool                   = resourceIpv4Pool{}
	ResourceIpv6Pool                   = resourceIpv6Pool{}
	ResourceTemplateCollapsed          = resourceTemplateCollapsed{}
	ResourceTemplatePodBased           = resourceTemplatePodBased{}
	ResourceVniPool                    = resourceVniPool{}
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
