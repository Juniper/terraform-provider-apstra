package sysredundancycache

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	mainMutex         = new(sync.Mutex)
	bpToMutex         = make(map[string]*sync.RWMutex)
	bpToGroupToSystem = make(map[string]map[string][2]string)
	bpToSystemToGroup = make(map[string]map[string]*string)
)

// rLockBP invokes RLock() on the sync.RWMutex for the given blueprint ID.
// If the per-blueprint mutex does not exist, it is created along with the cache entries
// for that blueprint. The function returns a function that, when invoked, will release
// the read lock on the per-blueprint mutex.
func rLockBP(bpID string) func() {
	mainMutex.Lock()
	defer mainMutex.Unlock()

	bpMutex := bpToMutex[bpID]
	if bpMutex == nil {
		// Per-BP mutex not found. We can assume that maps for this BP also have not been created.
		bpMutex = new(sync.RWMutex)                          // create a per-BP mutex
		bpToMutex[bpID] = bpMutex                            // store it in the map
		bpToGroupToSystem[bpID] = make(map[string][2]string) // create group->sys map for this BP
		bpToSystemToGroup[bpID] = make(map[string]*string)   // create sys->group map for this BP
	}

	bpMutex.RLock()
	return func() {
		bpMutex.RUnlock()
	}
}

// lockBP invokes Lock() on the sync.RWMutex for the given blueprint ID.
// If the per-blueprint mutex does not exist, it is created along with the cache entries
// for that blueprint. The function returns a function that, when invoked, will release
// the lock on the per-blueprint mutex.
func lockBP(bpID string) func() {
	mainMutex.Lock()
	defer mainMutex.Unlock()

	bpMutex := bpToMutex[bpID]
	if bpMutex == nil {
		// Per-BP mutex not found. We can assume that maps for this BP also have not been created.
		bpMutex = new(sync.RWMutex)                          // create a per-BP mutex
		bpToMutex[bpID] = bpMutex                            // store it in the map
		bpToGroupToSystem[bpID] = make(map[string][2]string) // create group->sys map for this BP
		bpToSystemToGroup[bpID] = make(map[string]*string)   // create sys->group map for this BP
	}

	bpMutex.Lock()
	return func() {
		bpMutex.Unlock()
	}
}

// refresh queries the given blueprint for all switch
func refresh(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// match(
	//  node(type='system', system_type='switch', name='n_sys'),
	//  optional(
	//    node(name='n_sys')
	//      .out(type='part_of_redundancy_group')
	//      .node(type='redundancy_group', name='n_grp')
	//  )
	//)
	query := new(apstra.MatchQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				// We filter on system_type='switch' to reduce the cache size. Generic Systems are also "systems"
				// in the DC refdesign graph, but there's lots of them and they're not interesting to us.
				// But not all switches can be part of a redundancy group. DC refdesign has a validation called
				// RG_SUPPORTS_LEAF_ACCESS which ensures that only those with role=is_in(['leaf', 'access']) can
				// be part of a redundancy group, so we *could* disregard spines and superspines as well.
				// We're not doing that because the count of spines and superspines will be low/insignificant
				// and this simplifies the error diagnostic returned by the lookup functions in case of an unknown
				// system ID. Rather than saying "no such switch of type leaf or access with that ID" (we won't
				// know which type we're looking for), we can return "no switch with that ID" because we'll have
				// cache entries for all switches, regardless of their role.
				{Key: "system_type", Value: apstra.QEStringVal(apstra.SystemTypeSwitch.String())},
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

	// target collects only the system ID and group ID (if any) for each system in the blueprint. The group
	// ID is a pointer to ensure we notice when a system is not part of a redundancy group (nil pointer)
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

	// Run the query.
	err := query.Do(ctx, &target)
	if err != nil {
		diags.AddError("failed to query for system redundancy groups", err.Error())
		return
	}

	// Populate a new system ID -> group ID map.
	systemToGroup := make(map[string]*string, len(target.Items))
	for _, item := range target.Items {
		systemToGroup[item.System.ID] = item.Group.ID // nil if system is not part of a redundancy group
	}

	// groupToSystemSet is a temporary map that collects system IDs for each redundancy group ID.
	// The value associated with each redundancy group ID is a set (map[string]struct{}) of system IDs.
	groupToSystemSet := make(map[string]map[string]struct{})
	for sysID, rgID := range systemToGroup {
		if rgID == nil {
			continue // Skip systems which are not part of a redundancy group.
		}

		if _, ok := groupToSystemSet[*rgID]; !ok {
			groupToSystemSet[*rgID] = make(map[string]struct{}) // First system for this redundancy group, so create the set.
		}
		groupToSystemSet[*rgID][sysID] = struct{}{} // Add the system ID to the set for this redundancy group.
	}

	// Having created a set of system IDs for each redundancy group, we can now
	// create the final map of redundancy group ID to system ID pair ([2]string).
	// If a redundancy group does not have exactly 2 members, we add an error to diags and return.
	groupToSystem := make(map[string][2]string)
	for rgID, sysIDSet := range groupToSystemSet {
		if len(sysIDSet) != 2 {
			diags.AddError(
				"failed to find redundancy group members",
				fmt.Sprintf("Redundancy Group ID %q in Blueprint %q does not have exactly 2 members.", rgID, bp.Id()),
			)
			return
		}

		// Convert the set of system IDs to an [2]string system ID pair.
		var pair [2]string
		for i, sysID := range slices.Collect(maps.Keys(sysIDSet)) {
			pair[i] = sysID
		}

		groupToSystem[rgID] = pair // Store the pair at the redundancy group ID key.
	}

	// Store both blueprint-specific maps in the global cache.
	bpToSystemToGroup[bp.Id().String()] = systemToGroup
	bpToGroupToSystem[bp.Id().String()] = groupToSystem
}
