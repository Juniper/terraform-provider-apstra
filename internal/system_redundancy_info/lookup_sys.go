package sysredundancyinfo

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// LookupSys returns the System IDs for the given redundancy group ID in the given Blueprint.
//
// Possible results:
// - Redundancy Group exists                                   : returns the member system IDs
// - Redundancy Group does not exist, or failure during lookup : adds an error to diags, returns a zero-value array
func LookupSys(ctx context.Context, bp *apstra.TwoStageL3ClosClient, rgID string, diags *diag.Diagnostics) [2]string {
	bpID := bp.Id().String()

	mutex.Lock()
	sysIDs, ok := bpsToRGsToSyss[bpID][rgID]
	mutex.Unlock()
	if ok {
		return sysIDs // Cache hit
	}

	// Cache miss - let's refresh the cache from the API and try again.
	refresh(ctx, bp, diags)
	if diags.HasError() {
		return [2]string{}
	}

	mutex.Lock()
	sysIDs, ok = bpsToRGsToSyss[bpID][rgID]
	mutex.Unlock()
	if ok {
		return sysIDs
	}

	diags.AddError("failed to find redundancy group in blueprint", fmt.Sprintf("Redundancy Group ID %q not found in Blueprint %q.", rgID, bpID))
	return [2]string{}
}
