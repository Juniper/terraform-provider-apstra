package apstravalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ validator.Bool = reserveVlanOkValidator{}

type reserveVlanOkValidator struct {
	expression path.Expression
	value      string
}

func (o reserveVlanOkValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that no value is supplied when attribute at %q has value %q", o.expression.String(), o.value)
}

func (o reserveVlanOkValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o reserveVlanOkValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// nothing to do when value is false, null or unknown (the last two rely on bool zero value)
	if !req.ConfigValue.ValueBool() {
		return
	}

	bindingPath := req.Path.ParentPath().AtName("bindings")
	var bindingsVal types.Map
	req.Config.GetAttribute(ctx, bindingPath, &bindingsVal)

	foundVlans := make(map[int64]struct{})
	for _, val := range bindingsVal.Elements() {
		var binding struct {
			VlanId    types.Int64 `tfsdk:"vlan_id"`
			AccessIds types.Set   `tfsdk:"access_ids"`
		}

		resp.Diagnostics.Append(val.(types.Object).As(ctx, &binding, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		if binding.VlanId.IsNull() {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
				bindingPath.AtSetValue(val), "binding \"vlan_id\" cannot be null when \"reserve_vlan\" is set"))
		}

		foundVlans[binding.VlanId.ValueInt64()] = struct{}{}
	}

	if len(foundVlans) <= 1 {
		return
	}

	vlanIds := make([]string, len(foundVlans))
	var i int
	for k := range foundVlans {
		vlanIds[i] = fmt.Sprintf("%d", k)
		i++
	}

	resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
		bindingPath, fmt.Sprintf("bindings must all use the same \"vlan_id\" value when %q is set", req.Path),
	))
}

// ReserveVlanOK looks through the `bindings` to ensure that each binding
// specifies the same VLAN ID. This is the only condition under which VLAN
// reservation is permitted by Apstra.
func ReserveVlanOK() validator.Bool {
	return reserveVlanOkValidator{}
}
