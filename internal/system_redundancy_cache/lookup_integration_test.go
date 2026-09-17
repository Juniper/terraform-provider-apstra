package sysredundancycache_test

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	cache "github.com/Juniper/terraform-provider-apstra/internal/system_redundancy_cache"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestLookup(t *testing.T) {
	ctx := context.Background()

	bp := testutils.BlueprintC(t, ctx)
	bpID := string(bp.Id())
	expectedGroupCount := 4
	expectedSystemCount := 15

	clearBPToSystemToGroupCache := func() {
		for key := range cache.BPToSystemToGroup {
			delete(cache.BPToSystemToGroup, key)
		}
	}
	clearBPToGroupToSystemsCache := func() {
		for key := range cache.BPToGroupToSystem {
			delete(cache.BPToGroupToSystem, key)
		}
	}

	t.Run("lookup_systems_using_bogus_group_id", func(t *testing.T) {
		// Begin by clearing the cache.
		clearBPToGroupToSystemsCache()
		clearBPToSystemToGroupCache()

		var diags diag.Diagnostics
		systems := cache.LookupSystem(ctx, bp, "bogus-group-id", &diags)
		require.True(t, diags.HasError(), "expected error for bogus group ID, but got none")
		require.Empty(t, systems[0], "expected first member of bogus redundant system pair to have empty ID")
		require.Empty(t, systems[1], "expected second member of bogus redundant system pair to have empty ID")
		require.Equal(t, 1, len(cache.BPToGroupToSystem))                         // one blueprint in the group->system cache
		require.Equal(t, expectedGroupCount, len(cache.BPToGroupToSystem[bpID]))  // expectedGroupCount groups in the per-bp cache
		require.Equal(t, 1, len(cache.BPToSystemToGroup))                         // one blueprint in the system->group cache
		require.Equal(t, expectedSystemCount, len(cache.BPToSystemToGroup[bpID])) // expectedSystemCount systems in the per-bp cache
	})

	t.Run("lookup_group_using_bogus_system_id", func(t *testing.T) {
		// Begin by clearing the cache.
		clearBPToGroupToSystemsCache()
		clearBPToSystemToGroupCache()

		var diags diag.Diagnostics
		group := cache.LookupGroup(ctx, bp, "bogus-system-id", &diags)
		require.Nilf(t, group, "expected nil group for bogus system id")
		require.True(t, diags.HasError(), "expected error for bogus system ID, but got none")
		require.Equal(t, 1, len(cache.BPToGroupToSystem))                         // one blueprint in the group->system cache
		require.Equal(t, expectedGroupCount, len(cache.BPToGroupToSystem[bpID]))  // expectedGroupCount groups in the per-bp cache
		require.Equal(t, 1, len(cache.BPToSystemToGroup))                         // one blueprint in the system->group cache
		require.Equal(t, expectedSystemCount, len(cache.BPToSystemToGroup[bpID])) // expectedSystemCount systems in the per-bp cache
	})

	// Function which returns system IDs of switch nodes.
	switchIDs := func(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient) []string {
		q := new(apstra.PathQuery).
			SetClient(bp.Client()).
			SetBlueprintId(apstra.ObjectId(bpID)).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "system_type", Value: apstra.QEStringVal(apstra.SystemTypeSwitch.String())},
				{Key: "name", Value: apstra.QEStringVal("n_sys")},
			})

		var target struct {
			Items []struct {
				System struct {
					ID string `json:"id"`
				} `json:"n_sys"`
			} `json:"items"`
		}

		require.NoError(t, q.Do(ctx, &target))
		result := make([]string, len(target.Items))
		for i, item := range target.Items {
			result[i] = item.System.ID
		}
		return result
	}

	// Function which returns redundancy group node IDs.
	groupIDs := func(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient) []string {
		q := new(apstra.PathQuery).
			SetClient(bp.Client()).
			SetBlueprintId(apstra.ObjectId(bpID)).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeRedundancyGroup.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("n_grp")},
			})

		var target struct {
			Items []struct {
				Group struct {
					ID string `json:"id"`
				} `json:"n_grp"`
			} `json:"items"`
		}

		require.NoError(t, q.Do(ctx, &target))
		result := make([]string, len(target.Items))
		for i, item := range target.Items {
			result[i] = item.Group.ID
		}
		return result
	}

	// Create maps to keep track of every system and group we encounter during the following tests.
	systemIDSet := make(map[string]struct{})
	groupIDSet := make(map[string]struct{})

	t.Run("lookup_all_groups", func(t *testing.T) {
		// Begin by clearing the cache.
		clearBPToGroupToSystemsCache()
		clearBPToSystemToGroupCache()

		groupCount := 0
		for _, groupID := range groupIDs(t, ctx, bp) {
			groupCount++
			groupIDSet[groupID] = struct{}{}
			var diags diag.Diagnostics
			systemIDs := cache.LookupSystem(ctx, bp, groupID, &diags)
			require.False(t, diags.HasError())
			require.NotEmpty(t, systemIDs[0])
			require.NotEmpty(t, systemIDs[1])
			systemIDSet[systemIDs[0]] = struct{}{}
			systemIDSet[systemIDs[1]] = struct{}{}
		}
		require.Equal(t, expectedGroupCount, groupCount)
	})

	t.Run("lookup_all_systems", func(t *testing.T) {
		// Begin by clearing the cache.
		clearBPToGroupToSystemsCache()
		clearBPToSystemToGroupCache()

		redundantSystemCount := 0
		for _, systemID := range switchIDs(t, ctx, bp) {
			systemIDSet[systemID] = struct{}{}
			var diags diag.Diagnostics
			groupID := cache.LookupGroup(ctx, bp, systemID, &diags)
			require.False(t, diags.HasError())
			if groupID != nil {
				groupIDSet[*groupID] = struct{}{}
				redundantSystemCount++
			}
		}
		require.Equal(t, expectedGroupCount*2, redundantSystemCount)
	})

	require.Equal(t, expectedGroupCount, len(groupIDSet))
	require.Equal(t, expectedSystemCount, len(systemIDSet))
	require.Equal(t, expectedGroupCount, len(cache.BPToGroupToSystem[bpID]))
	require.Equal(t, expectedSystemCount, len(cache.BPToSystemToGroup[bpID]))

	t.Run("concurrent_access", func(t *testing.T) {
		// Begin by clearing the cache.
		clearBPToGroupToSystemsCache()
		clearBPToSystemToGroupCache()

		// We need to get system and group IDs by index below.
		systemIDSlice := slices.Collect(maps.Keys(systemIDSet))
		groupIDSlice := slices.Collect(maps.Keys(groupIDSet))

		var wg sync.WaitGroup
		numGoRoutines := 100
		wg.Add(numGoRoutines)

		for i := range numGoRoutines {
			go func() {
				switch {
				case i%7 == 0: // bogus group lookup every 7th request
					var diags diag.Diagnostics
					g := cache.LookupGroup(ctx, bp, fmt.Sprintf("bogus_system_%03d", i), &diags)
					require.Nil(t, g)
					require.True(t, diags.HasError())
				case i%6 == 0: // bogus system lookup every 6th request
					var diags diag.Diagnostics
					s := cache.LookupSystem(ctx, bp, fmt.Sprintf("bogus_group_%03d", i), &diags)
					require.True(t, diags.HasError())
					require.Empty(t, s[0])
					require.Empty(t, s[1])
				case i%2 == 0: // valid group lookup on even numbers (not divisible by 6 or 7)
					var diags diag.Diagnostics
					testSys := systemIDSlice[i%len(systemIDSlice)]
					result := cache.LookupGroup(ctx, bp, testSys, &diags)
					require.False(t, diags.HasError())
					if result != nil { // if we got a group ID, run it the other way.
						s := cache.LookupSystem(ctx, bp, *result, &diags)
						require.False(t, diags.HasError())
						require.Contains(t, s, testSys)
					}
				case i%2 == 1: // valid system lookup on odd numbers (not divisible by 6 or 7)
					var diags diag.Diagnostics
					testGrp := groupIDSlice[i%len(groupIDSlice)]
					result := cache.LookupSystem(ctx, bp, testGrp, &diags)
					require.False(t, diags.HasError())
					for _, s := range result { // look up the group associated with each returned sys ID
						g := cache.LookupGroup(ctx, bp, s, &diags)
						require.False(t, diags.HasError())
						require.NotNil(t, g)
						require.Equal(t, testGrp, *g)
					}
				}
				wg.Done()
			}()
		}

		wg.Wait()
	})
}
