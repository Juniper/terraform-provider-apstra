package apstradefault

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ defaults.Set = PortRolesDefault{}

type PortRolesDefault struct{}

func (o PortRolesDefault) Description(_ context.Context) string {
	return "Enables all port roles."
}

func (o PortRolesDefault) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o PortRolesDefault) DefaultSet(ctx context.Context, _ defaults.SetRequest, resp *defaults.SetResponse) {
	var allRoleFlagsSet apstra.LogicalDevicePortRoleFlags
	allRoleFlagsSet.SetAll()

	var d diag.Diagnostics
	resp.PlanValue, d = types.SetValueFrom(ctx, types.StringType, allRoleFlagsSet.Strings())
	resp.Diagnostics.Append(d...)
}
