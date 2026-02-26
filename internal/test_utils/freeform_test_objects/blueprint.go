package fftestobj

import (
	"context"
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"testing"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// TestBlueprintA creates a freeform blueprint
func TestBlueprintA(t testing.TB, ctx context.Context, client apstra.Client) apstra.FreeformClient {
	t.Helper()

	id, err := client.CreateFreeformBlueprint(ctx, acctest.RandString(6))
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, 10*time.Second, func(ctx context.Context) error {
		return client.DeleteBlueprint(ctx, id)
	})

	bp, err := client.NewFreeformClient(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, bp)

	return *bp
}

// TestBlueprintB creates a freeform blueprint containing systems and links for testing
// link aggregation between internal systems. It creates two groups of servers according
// to sysCount1 and sysCount2 and links every sysCount1 system to every sysCount2 system
// linkCount times.
// The returned values are a blueprint client and a map of link IDs to [2]string of
// system IDs associated with each link. index 0 is the id from group 1, index 1 is the
// id from group 2.
//
// with sysCount1=2, sysCount2=3 and linkCount=1 we should get:
/*
                    +------+
      +--------------+ G2-1 |
      |              +-+----+
      |                |
      |          +-----+
 +------+        |
 | G1-1 |--------|----+
 +----+-+        |     |
      |          |     |
      +-----+    |     |
            |    |   +-+----+
      +-----|----+   | G2-2 |
      |     |        +-+----+
 +----+-+   |           |
 | G1-2 |---|-----------+
 +------+   |
      |     +----------+
      |                |
      |              +-+----+
      +------------- | G2-2 |
                     +------+
*/
func TestBlueprintB(t testing.TB, ctx context.Context, client apstra.Client, sysCount1, sysCount2, linkCount int) (apstra.FreeformClient, map[string][2]string) {
	t.Helper()

	dpToImport := "Juniper_vEX"
	dp, err := client.GetDeviceProfile(ctx, dpToImport)
	require.NoError(t, err)

	if sysCount1*sysCount2*linkCount > len(dp.Ports) {
		t.Fatalf(
			"cannot link %d group 1 servers to %d group 2 servers %d times with device profile %s: %d ports requied of %d available",
			sysCount1, sysCount2, linkCount, dpToImport, sysCount1*sysCount2*linkCount, len(dp.Ports),
		)
	}

	// create the blueprint
	id, err := client.CreateFreeformBlueprint(ctx, acctest.RandString(6))
	require.NoError(t, err)

	// cleanup
	testutils.CleanupWithFreshContext(t, 10*time.Second, func(ctx context.Context) error {
		return client.DeleteBlueprint(ctx, id)
	})

	// create a client
	bp, err := client.NewFreeformClient(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, bp)

	// import the device profile
	bpDPID, err := bp.ImportDeviceProfile(ctx, apstra.ObjectId(dpToImport))
	require.NoError(t, err)

	// create systems of EP group 1
	systemsGroup1 := make([]string, sysCount1)
	for i := range systemsGroup1 {
		id, err := bp.CreateSystem(ctx, &apstra.FreeformSystemData{
			Label:           fmt.Sprintf("G1-S%d", i+1),
			Type:            apstra.SystemTypeInternal,
			DeviceProfileId: &bpDPID,
		})
		require.NoError(t, err)
		systemsGroup1[i] = id.String()
	}

	// create systems of EP group 2
	systemsGroup2 := make([]string, sysCount2)
	for i := range systemsGroup2 {
		id, err := bp.CreateSystem(ctx, &apstra.FreeformSystemData{
			Label:           fmt.Sprintf("G2-S%d", i+1),
			Type:            apstra.SystemTypeInternal,
			DeviceProfileId: &bpDPID,
		})
		require.NoError(t, err)
		systemsGroup2[i] = id.String()
	}

	// map keyed by link ID with value [2]string{idOfServerFromGroup0, idOfServerFromGroup1}
	linkIdToServerIDs := make(map[string][2]string)

	// make a full mesh of links between the two ep groups: everybody in group0 linked to everybody in group1
	for linkIdx := range linkCount {
		for sysGroup1Idx := range systemsGroup1 {
			for sysGroup2Idx := range systemsGroup2 {
				linkID, err := bp.CreateLink(ctx, &apstra.FreeformLinkRequest{
					Label: acctest.RandString(6),
					Tags:  nil,
					Endpoints: [2]apstra.FreeformEthernetEndpoint{
						{
							SystemId: apstra.ObjectId(systemsGroup1[sysGroup1Idx]),
							Interface: apstra.FreeformInterface{Data: &apstra.FreeformInterfaceData{
								IfName:           pointer.To(fmt.Sprintf("ge-0/0/%d", sysGroup2Idx+(linkIdx*sysCount1))),
								TransformationId: pointer.To(1),
							}},
						},
						{
							SystemId: apstra.ObjectId(systemsGroup2[sysGroup2Idx]),
							Interface: apstra.FreeformInterface{Data: &apstra.FreeformInterfaceData{
								IfName:           pointer.To(fmt.Sprintf("ge-0/0/%d", sysGroup1Idx+(linkIdx*len(systemsGroup1)))),
								TransformationId: pointer.To(1),
							}},
						},
					},
				})
				require.NoError(t, err)
				linkIdToServerIDs[linkID.String()] = [2]string{systemsGroup1[sysGroup1Idx], systemsGroup2[sysGroup2Idx]}
			}
		}
	}

	return *bp, linkIdToServerIDs
}
