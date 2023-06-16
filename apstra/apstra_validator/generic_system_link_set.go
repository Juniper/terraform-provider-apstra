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

	switchIdToInterfaces := make(map[string]map[string]bool, len(links))
	for _, link := range links {
		sid := link.TargetSwitchId.ValueString()
		ifn := link.TargetSwitchIfName.ValueString()

		if switchIdToInterfaces[sid] == nil {
			// first time we've seen this switch, create a new interface map to keep track of it.
			switchIdToInterfaces[sid] = map[string]bool{ifn: true}
			continue
		}

		if switchIdToInterfaces[sid][ifn] {
			// this is the second link claiming this combination of switch + interface!
			resp.Diagnostics.Append(
				validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("multiple links use system %s and interface %s", link.TargetSwitchId, link.TargetSwitchIfName),
				),
			)
			return
		}

		// this is the first link claiming this interface
		switchIdToInterfaces[sid][ifn] = true
	}
}

func GenericSystemLinksNoOverlap() validator.Set {
	return genericSystemLinkSetValidator{}
}
