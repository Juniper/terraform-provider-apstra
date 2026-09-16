package sysredundancyinfo

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
	unlock() // release loc for read
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

// LookupSystem returns the System IDs for the given redundancy group ID in the given Blueprint.
//
// Possible results:
// - Redundancy Group exists                                   : returns the member system IDs
// - Redundancy Group does not exist, or failure during lookup : adds an error to diags, returns a zero-value array
func LookupSystem(ctx context.Context, bp *apstra.TwoStageL3ClosClient, rgID string, diags *diag.Diagnostics) [2]string {
	bpID := bp.Id().String()

	unlock := rLockBP(bpID) // lock for read
	sysIDs, ok := bpToGroupToSystem[bpID][rgID]
	unlock() // release the lock for read
	if ok {
		return sysIDs // Cache hit - Success!
	}

	// Cache miss - We may need to refresh the cache. Begin by acquiring a write lock.
	unlock = lockBP(bpID)
	defer unlock()

	// Check the cache one more time after acquiring the write lock, in case another thread refreshed it while we were waiting.
	sysIDs, ok = bpToGroupToSystem[bpID][rgID]
	if ok {
		return sysIDs // Cache hit - Success!
	}

	// Another cache miss - refresh the cache.
	refresh(ctx, bp, diags)
	if diags.HasError() {
		return [2]string{}
	}

	// Now that we've refreshed the cache, check it one last time.
	sysIDs, ok = bpToGroupToSystem[bpID][rgID]
	if ok {
		return sysIDs // Cache hit - Success!
	}

	// Probably a bogus group ID.
	diags.AddError("failed to find redundancy group in blueprint", fmt.Sprintf("Redundancy Group ID %q not found in Blueprint %q.", rgID, bpID))
	return [2]string{}
}
