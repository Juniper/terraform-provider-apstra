package sysredundancycache

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// LookupGroup returns the Redundancy Group ID for the given system ID in the given Blueprint.
//
// Possible results:
// - System exists and is part of a redundancy group     : returns a non-nil pointer to the RG ID
// - System exists and is not part of a redundancy group : returns nil
// - System does not exist, or failure during lookup     : returns nil and adds an error to diags
func LookupGroup(ctx context.Context, bp *apstra.TwoStageL3ClosClient, systemID string, diags *diag.Diagnostics) *string {
	bpID := bp.Id().String()

	unlock := rLockBP(bpID) // lock for read
	rgID, ok := bpToSystemToGroup[bpID][systemID]
	unlock() // release lock for read
	if ok {
		return rgID // Cache hit - Success!
	}

	// Cache miss - We may need to refresh the cache. Begin by acquiring a write lock.
	unlock = lockBP(bpID)
	defer unlock()

	// Check the cache one more time after acquiring the write lock, in case another thread refreshed it while we were waiting.
	rgID, ok = bpToSystemToGroup[bpID][systemID]
	if ok {
		return rgID // Cache hit - Success!
	}

	// Another cache miss - refresh the cache.
	refresh(ctx, bp, diags)
	if diags.HasError() {
		return nil
	}

	// Now that we've refreshed the cache, check it one last time.
	rgID, ok = bpToSystemToGroup[bpID][systemID]
	if ok {
		return rgID // Cache hit - Success!
	}

	// Probably a bogus system ID.
	diags.AddError("failed to find system in blueprint", fmt.Sprintf("System ID %q not found in Blueprint %q.", systemID, bpID))
	return nil
}

// LookupSystems returns a pair of System IDs representing the given redundancy group ID in the given Blueprint and a boolean indicating success.
//
// Possible results:
// - Redundancy Group exists                                     : returns the member system IDs, true
// - Redundancy Group does not exist, but no error during lookup : returns a zero-value array, false
// - Failure during lookup                                       : returns a zero-value array, false and adds an error to diags
func LookupSystems(ctx context.Context, bp *apstra.TwoStageL3ClosClient, rgID string, diags *diag.Diagnostics) ([2]string, bool) {
	bpID := bp.Id().String()

	unlock := rLockBP(bpID) // lock for read
	sysIDs, ok := bpToGroupToSystem[bpID][rgID]
	unlock() // release the lock for read
	if ok {
		return sysIDs, true // Cache hit - Success!
	}

	// Cache miss - We may need to refresh the cache. Begin by acquiring a write lock.
	unlock = lockBP(bpID)
	defer unlock()

	// Check the cache one more time after acquiring the write lock, in case another thread refreshed it while we were waiting.
	sysIDs, ok = bpToGroupToSystem[bpID][rgID]
	if ok {
		return sysIDs, true // Cache hit - Success!
	}

	// Another cache miss - refresh the cache.
	refresh(ctx, bp, diags)
	if diags.HasError() {
		return [2]string{}, false
	}

	// Now that we've refreshed the cache, check it one last time.
	sysIDs, ok = bpToGroupToSystem[bpID][rgID]
	if ok {
		return sysIDs, true // Cache hit - Success!
	}

	// Probably a bogus group ID.
	return [2]string{}, false
}
