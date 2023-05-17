package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
)

type redundancyGroup struct {
	role      string   // access / generic
	id        string   // redundancy_group graph db node id
	memberIds []string // id of leaf/access nodes in the group
}

// redunancyGroupIdToRedundancyGroupInfo returns a map keyed by redundancy
// group ID to *redundancyGroup representing all redundancy groups found in the
// blueprint
func redunancyGroupIdToRedundancyGroupInfo(ctx context.Context, client *apstra.TwoStageL3ClosClient) (map[string]*redundancyGroup, error) {
	pathQuery := new(apstra.PathQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("redundancy_group")},
			{Key: "name", Value: apstra.QEStringVal("n_redundancy_group")},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("composed_of_systems")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
			{Key: "role", Value: apstra.QEStringValIsIn{"access", "leaf"}},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	queryResponse := &struct {
		Items []struct {
			RedundancyGroup struct {
				Id string `json:"id"`
			} `json:"n_redundancy_group"`
			System struct {
				Id   string `json:"id"`
				Role string `json:"role"`
			} `json:"n_system"`
		} `json:"items"`
	}{}

	err := new(apstra.MatchQuery).
		Match(pathQuery).
		Distinct(apstra.MatchQueryDistinct{"n_system"}).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Do(ctx, queryResponse)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*redundancyGroup)
	for _, item := range queryResponse.Items {
		id := item.RedundancyGroup.Id
		if rg, ok := result[id]; ok {
			rg.memberIds = append(rg.memberIds, item.System.Id)
			result[id] = rg
		} else {
			result[id] = &redundancyGroup{
				role:      item.System.Role,
				id:        item.RedundancyGroup.Id,
				memberIds: []string{item.System.Id},
			}
		}
	}

	return result, nil
}
