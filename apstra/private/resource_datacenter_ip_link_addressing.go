package private

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ResourceDatacenterIpLinkAddressingInterfaceAddressing is stored in private state by
// ResourceDatacenterIpLinkAddressing.Create(). It is the record of the original numbering scheme
// on a logical link, and is restored by ResourceDatacenterIpLinkAddressing.Delete()
type ResourceDatacenterIpLinkAddressingInterfaceAddressing struct {
	SwitchIpv4  enum.InterfaceNumberingIpv4Type `json:"switch_ipv4"`
	SwitchIpv6  enum.InterfaceNumberingIpv6Type `json:"switch_ipv6"`
	GenericIpv4 enum.InterfaceNumberingIpv4Type `json:"generic_ipv4"`
	GenericIpv6 enum.InterfaceNumberingIpv6Type `json:"generic_ipv6"`
}

func (o *ResourceDatacenterIpLinkAddressingInterfaceAddressing) LoadApiData(_ context.Context, link *apstra.TwoStageL3ClosSubinterfaceLink, diags *diag.Diagnostics) {
	switchEp := epBySystemType(apstra.SystemTypeSwitch, link.Endpoints, diags)
	if diags.HasError() {
		return
	}

	genericEp := epBySystemType(apstra.SystemTypeServer, link.Endpoints, diags)
	if diags.HasError() {
		return
	}

	o.SwitchIpv4 = switchEp.Subinterface.Ipv4AddrType
	o.SwitchIpv6 = switchEp.Subinterface.Ipv6AddrType
	o.GenericIpv4 = genericEp.Subinterface.Ipv4AddrType
	o.GenericIpv6 = genericEp.Subinterface.Ipv6AddrType
}

func (o *ResourceDatacenterIpLinkAddressingInterfaceAddressing) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, d := ps.GetKey(ctx, "ResourceDatacenterIpLinkAddressingInterfaceAddressing")
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	err := json.Unmarshal(b, &o)
	if err != nil {
		diags.AddError("failed to unmarshal private state", err.Error())
		return
	}
}

func (o *ResourceDatacenterIpLinkAddressingInterfaceAddressing) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, "ResourceDatacenterIpLinkAddressingInterfaceAddressing", b)...)
}

// ResourceDatacenterIpLinkAddressingInterfaceIds contains the logical interfaces associated with
// a logical link. It turns out that these interface IDs are NOT immutable. The interfaces, along
// with the logical link are created as a side-effect of associating a CT containing IP Link
// primitives with a physical switch port. Modifying the CT may cause the logical link and logical
// interface pair to be replaced. The logical link is indistinguishable from immutable because
// its ID is constructed by encoding other information. The interfaces, on the other hand, use
// random IDs, so they may be found to have changed from one run to the next. As a result, this
// "private data" struct is no longer committed to private state. It's now relegated to merely
// unpacking the API response via the LoadApiData() method.
type ResourceDatacenterIpLinkAddressingInterfaceIds struct {
	SwitchInterface  apstra.ObjectId `json:"switch_interface"`
	GenericInterface apstra.ObjectId `json:"generic_interface"`
}

func (o *ResourceDatacenterIpLinkAddressingInterfaceIds) LoadApiData(_ context.Context, link *apstra.TwoStageL3ClosSubinterfaceLink, diags *diag.Diagnostics) {
	switchEp := epBySystemType(apstra.SystemTypeSwitch, link.Endpoints, diags)
	if diags.HasError() {
		return
	}

	genericEp := epBySystemType(apstra.SystemTypeServer, link.Endpoints, diags)
	if diags.HasError() {
		return
	}

	o.SwitchInterface = switchEp.SubinterfaceId
	o.GenericInterface = genericEp.SubinterfaceId
}

func (o *ResourceDatacenterIpLinkAddressingInterfaceIds) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, d := ps.GetKey(ctx, "ResourceDatacenterIpLinkAddressingInterfaceIds")
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	err := json.Unmarshal(b, o)
	if err != nil {
		diags.AddError("failed to unmarshal private state", err.Error())
		return
	}
}

func (o *ResourceDatacenterIpLinkAddressingInterfaceIds) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, "ResourceDatacenterIpLinkAddressingInterfaceIds", b)...)
}

func epBySystemType(sysType apstra.SystemType, eps []apstra.TwoStageL3ClosSubinterfaceLinkEndpoint, diags *diag.Diagnostics) *apstra.TwoStageL3ClosSubinterfaceLinkEndpoint {
	var systemRoles []apstra.SystemRole

	switch sysType {
	case apstra.SystemTypeSwitch:
		systemRoles = []apstra.SystemRole{apstra.SystemRoleSuperSpine, apstra.SystemRoleSpine, apstra.SystemRoleLeaf, apstra.SystemRoleAccess}
	case apstra.SystemTypeServer:
		systemRoles = []apstra.SystemRole{apstra.SystemRoleGeneric}
	default:
		diags.AddError(constants.ErrProviderBug, fmt.Sprintf("unexpected system type %q", sysType))
		return nil
	}

	var result *apstra.TwoStageL3ClosSubinterfaceLinkEndpoint
	for _, ep := range eps {
		ep := ep
		if utils.SliceContains(ep.System.Role, systemRoles) {
			if result != nil {
				diags.AddError(
					"Unexpected API response",
					fmt.Sprintf("Logical link has multiple endpoints on systems with %q roles", sysType),
				)
				return nil
			}

			result = &ep
		}
	}

	if result == nil {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("Logical link has no endpoints on systems with %q roles", sysType),
		)
	}

	return result
}
