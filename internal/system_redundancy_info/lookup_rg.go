package sysredundancyinfo

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// LookupRG returns the Redundancy Group ID for the given system ID in the given Blueprint.
//
// Possible results:
// - System exists and is part of a redundancy group     : returns a non-nil pointer to the RG ID
// - System exists and is not part of a redundancy group : returns nil
// - System does not exist, or failure during lookup     : returns nil and adds an error to diags
func LookupRG(ctx context.Context, bp *apstra.TwoStageL3ClosClient, sysID string, diags *diag.Diagnostics) *string {
	bpID := bp.Id().String()

	mutex.Lock()
	rgID, ok := bpsToSysToRGs[bpID][sysID]
	mutex.Unlock()
	if ok {
		return rgID // Cache hit
	}

	// Cache miss - let's refresh the cache from the API and try again.
	refresh(ctx, bp, diags)
	if diags.HasError() {
		return nil
	}

	mutex.Lock()
	rgID, ok = bpsToSysToRGs[bpID][sysID]
	mutex.Unlock()
	if ok {
		return rgID
	}

	diags.AddError("failed to find system in blueprint", fmt.Sprintf("System ID %q not found in Blueprint %q.", sysID, bpID))
	return nil
}
