package private

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ResourceDatacenterVirtualNetworkBindings contains a map[string]int64 which
// records the VLAN specified for each leaf switch. Value 0 indicates that the
// user did not specify a VLAN. This object is stored in private state by
// ResourceDatacenterIpLinkAddressing.Create() and .Update().
// The .Read() method, relies on it to ignore bindings not previously created by
// and by Request() (Create(), Update(), Delete()) where it is used as a record
// of previously-created bindings which may need to be deleted.
type ResourceDatacenterVirtualNetworkBindings struct {
	SystemIdToVlan map[string]int64 `json:"system_id_to_vlan"`
	//RedundancyGroupIdToSystemIDs map[string][]string `json:"redundancy_group_id_to_system_ids"`
}

func (o *ResourceDatacenterVirtualNetworkBindings) LoadSystemIdToVlanApiData(_ context.Context, bindings []apstra.VnBinding, _ *diag.Diagnostics) {
	o.SystemIdToVlan = make(map[string]int64, len(bindings))
	for _, binding := range bindings {
		if binding.VlanId == nil {
			o.SystemIdToVlan[binding.SystemId.String()] = 0 // no VLAN specified
		} else {
			o.SystemIdToVlan[binding.SystemId.String()] = int64(*binding.VlanId)
		}
	}
}

//func (o *ResourceDatacenterVirtualNetworkBindings) LoadRedundancyGroupIdToSystemIDsApiData(_ context.Context, rgiMap map[string]*apstra.RedundancyGroupInfo, ids []string, _ *diag.Diagnostics) {
//	o.RedundancyGroupIdToSystemIDs = make(map[string][]string)
//	for _, id := range ids {
//		if entry, ok := rgiMap[id]; ok {
//			o.RedundancyGroupIdToSystemIDs[entry.Id.String()] = []string{entry.SystemIds[0].String(), entry.SystemIds[1].String()}
//		}
//	}
//}

func (o *ResourceDatacenterVirtualNetworkBindings) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, d := ps.GetKey(ctx, fmt.Sprintf("%T", *o))
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	if len(b) == 0 {
		return
	}

	err := json.Unmarshal(b, &o)
	if err != nil {
		diags.AddError("failed to unmarshal private state", err.Error())
		return
	}
}

func (o *ResourceDatacenterVirtualNetworkBindings) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(o)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, fmt.Sprintf("%T", *o), b)...)
}
