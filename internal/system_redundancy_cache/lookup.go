package sysredundancycache

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
)

const (
	SystemNotFoundError = "system not found in cache"
	GroupNotFoundError  = "group not found in cache"
	NotFound            = "ID not found in cache"
)

// LookupGroup returns the Redundancy Group ID and the peer system ID for the given system ID in the given Blueprint.
//
// Possible results:
// - System exists and is part of a redundancy group     : returns non-nil pointers to the RG ID and the peer system ID and nil error
// - System exists and is not part of a redundancy group : returns nil, nil, nil
// - System does not exist, or failure during lookup     : returns nil, nil, error
func LookupGroup(ctx context.Context, bp *apstra.TwoStageL3ClosClient, systemID string) (*string, *string, error) {
	bpID := bp.Id().String()

	getPeer := func(groupID *string) *string {
		if groupID == nil {
			return nil
		}

		systemIDs := bpToGroupToSystems[bpID][*groupID]
		var peer string
		if systemID == systemIDs[0] {
			peer = systemIDs[1]
		} else {
			peer = systemIDs[0]
		}
		return &peer
	}

	unlock := rLockBP(bpID) // lock for read
	rgID, ok := bpToSystemToGroup[bpID][systemID]
	if ok {
		// Cache hit - Success!
		defer unlock() // Release lock for read on return.
		return rgID, getPeer(rgID), nil
	}

	// Cache miss - We may need to refresh the cache. Replace our reader lock with a writer lock.
	unlock()              // Release lock for read.
	unlock = lockBP(bpID) // Acquire a lock for write.
	defer unlock()        // Release lock for write on return.

	// Check the cache one more time after acquiring the write lock, in case another thread refreshed it while we were waiting.
	rgID, ok = bpToSystemToGroup[bpID][systemID]
	if ok {
		// Cache hit - Success!
		defer unlock() // Release lock for read on return.
		return rgID, getPeer(rgID), nil
	}

	// Another cache miss - refresh the cache.
	err := refresh(ctx, bp)
	if err != nil {
		return nil, nil, fmt.Errorf("refreshing cache: %w", err)
	}

	// Now that we've refreshed the cache, check it one last time.
	rgID, ok = bpToSystemToGroup[bpID][systemID]
	if ok {
		// Cache hit - Success!
		return rgID, getPeer(rgID), nil
	}

	// Probably a bogus system ID.
	return nil, nil, errors.New(SystemNotFoundError)
}

// LookupSystems returns a pair of System IDs representing the given redundancy group ID in the given Blueprint and a boolean indicating success.
//
// Possible results:
// - Redundancy Group exists                                 : returns the member system IDs, nil
// - Redundancy Group does not exist or failure during lookup: returns a zero-value array, error
func LookupSystems(ctx context.Context, bp *apstra.TwoStageL3ClosClient, rgID string) ([2]string, error) {
	bpID := bp.Id().String()

	unlock := rLockBP(bpID) // lock for read
	sysIDs, ok := bpToGroupToSystems[bpID][rgID]
	if ok {
		unlock()           // release the lock for read
		return sysIDs, nil // Cache hit - Success!
	}

	// Try refreshing the cache. Release the read lock before acquiring the write lock.
	unlock()              // Release the lock for read.
	unlock = lockBP(bpID) // Acquire a lock for write.
	defer unlock()        // Release the lock for write on return.

	// Check the cache one more time after acquiring the write lock, in case another thread refreshed it while we were waiting.
	sysIDs, ok = bpToGroupToSystems[bpID][rgID]
	if ok {
		return sysIDs, nil // Cache hit - Success!
	}

	// Another cache miss - refresh the cache.
	err := refresh(ctx, bp)
	if err != nil {
		return [2]string{}, fmt.Errorf("refreshing cache: %w", err)
	}

	// Now that we've refreshed the cache, check it one last time.
	if sysIDs, ok = bpToGroupToSystems[bpID][rgID]; ok {
		return sysIDs, nil // Cache hit - Success!
	}

	// Probably a bogus group ID.
	return [2]string{}, errors.New(GroupNotFoundError)
}

func LookupNodeType(ctx context.Context, bp *apstra.TwoStageL3ClosClient, id string) (apstra.NodeType, error) {
	bpID := bp.Id().String()

	typeFromCache := func() *apstra.NodeType {
		if slices.Contains(slices.Collect(maps.Keys(bpToGroupToSystems[string(bp.Id())])), id) {
			return pointer.To(apstra.NodeTypeRedundancyGroup)
		}
		if slices.Contains(slices.Collect(maps.Keys(bpToSystemToGroup[string(bp.Id())])), id) {
			return pointer.To(apstra.NodeTypeSystem)
		}
		return nil
	}

	unlock := rLockBP(bpID) // lock for read

	// Check the cache
	if t := typeFromCache(); t != nil {
		unlock()
		return *t, nil
	}

	// Try refreshing the cache. Release the read lock before acquiring the write lock.
	unlock()              // Release the lock for read.
	unlock = lockBP(bpID) // Acquire a lock for write.
	defer unlock()        // Release the lock for write on return.

	// Check the cache one more time after acquiring the write lock, in case another thread refreshed it while we were waiting.
	if t := typeFromCache(); t != nil {
		return *t, nil
	}

	// Another cache miss - refresh the cache.
	err := refresh(ctx, bp)
	if err != nil {
		return apstra.NodeTypeNone, fmt.Errorf("refreshing cache: %w", err)
	}

	// Now that we've refreshed the cache, check it one last time.
	if t := typeFromCache(); t != nil {
		return *t, nil
	}

	return apstra.NodeTypeNone, errors.New(NotFound)
}
