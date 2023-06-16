package apstravalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ validator.Set = genericSystemLinkSetValidator{}

type genericSystemLinkSetValidator struct{}

func (o genericSystemLinkSetValidator) Description(_ context.Context) string {
	return "ensures that links each use a unique combination of system_id + interface_name"
}

func (o genericSystemLinkSetValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o genericSystemLinkSetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	var links []blueprint.DatacenterGenericSystemLink
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &links, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	linkDigests := make(map[string]bool, len(links))
	for _, link := range links {
		digest := link.Digest()
		if linkDigests[digest] {
			resp.Diagnostics.Append(
				validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("multiple links use system %s and interface %s", link.TargetSwitchId, link.TargetSwitchIfName),
				),
			)
		}
		linkDigests[link.Digest()] = true
	}
}

func GenericSystemLinksNoOverlap() validator.Set {
	return genericSystemLinkSetValidator{}
}
