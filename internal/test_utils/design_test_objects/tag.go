//go:build integration

package designtestobjects

import (
	"context"
	"testing"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/design"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func RandomTag(ctx context.Context, t testing.TB, client *apstra.Client) string {
	t.Helper()

	id, err := client.CreateTag2(ctx, design.Tag{
		Label:       acctest.RandStringFromCharSet(8, acctest.CharSetAlpha),
		Description: acctest.RandStringFromCharSet(15, acctest.CharSetAlpha),
	})
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, 10*time.Second, func(ctx context.Context) error {
		return client.DeleteTag2(ctx, id)
	})

	return id
}
