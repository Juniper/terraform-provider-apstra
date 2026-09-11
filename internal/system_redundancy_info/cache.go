package sysredundancyinfo

import (
	"context"
	"fmt"
	"sync"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	mutex          = new(sync.Mutex)
	bpsToSysToRGs  = make(map[string]map[string]*string)
	bpsToRGsToSyss = make(map[string]map[string][2]string)
)

func refresh(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	query := new(apstra.MatchQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("n_sys")},
			}),
		).
		Optional(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_sys")}}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRedundancyGroup.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeRedundancyGroup.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("n_grp")},
			}),
		)

	var target struct {
		Items []struct {
			System struct {
				ID string `json:"id"`
			} `json:"n_sys"`
			Group struct {
				ID *string `json:"id"`
			} `json:"n_grp"`
		} `json:"items"`
	}

	err := query.Do(ctx, &target)
	if err != nil {
		diags.AddError("failed to query for system redundancy groups", err.Error())
		return
	}

	bySysMap := make(map[string]*string, len(target.Items))
	for _, item := range target.Items {
		bySysMap[item.System.ID] = item.Group.ID
	}

	tempMap := make(map[string]map[string]struct{})
	for sysID, rgID := range bySysMap {
		if rgID == nil {
			continue
		}

		if _, ok := tempMap[*rgID]; !ok {
			tempMap[*rgID] = make(map[string]struct{})
		}
		tempMap[*rgID][sysID] = struct{}{}
	}

	byRGMap := make(map[string][2]string)
	for rgID, sysIDSet := range tempMap {
		if len(sysIDSet) != 2 {
			diags.AddError(
				"failed to find redundancy group members",
				fmt.Sprintf("Redundancy Group ID %q in Blueprint %q does not have exactly 2 members.", rgID, bp.Id()),
			)
			return
		}

		var arr [2]string
		i := 0
		for sysID := range sysIDSet {
			arr[i] = sysID
			i++
		}
		byRGMap[rgID] = arr
	}

	mutex.Lock()
	bpsToSysToRGs[bp.Id().String()] = bySysMap
	bpsToRGsToSyss[bp.Id().String()] = byRGMap
	mutex.Unlock()
}
