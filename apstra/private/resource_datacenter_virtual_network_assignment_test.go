package private_test

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestResourceDatacenterVirtualNetworkAssignment(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	bp, err := client.NewTwoStageL3ClosClient(ctx, "b604a859-24fd-457d-897f-3004f9611f6b")
	require.NoError(t, err)

	p := private.ResourceDatacenterVirtualNetworkAssignment{
		//ConfiguredLeafID:       "Ggj14cLNNdoJdQ9ucg",
		//ConfiguredLeafNodeType: apstra.NodeTypeSystem,
		//ConfiguredLeafID:       "8J-rf-OpRjinMpGI8A",
		//ConfiguredLeafNodeType: apstra.NodeTypeRedundancyGroup,
		ConfiguredLeafID:       "7n-MT5vfxkbdAflElQ",
		ConfiguredLeafNodeType: apstra.NodeTypeSystem,
	}

	var diags diag.Diagnostics
	p.FetchRedundancyGroups(ctx, "6XzOo-FbRhumeFEPOA", bp, &diags)
	require.False(t, diags.HasError())
}
