package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"testing"
)

func TagA(t testing.TB, ctx context.Context) *apstra.DesignTag {
	t.Helper()

	client := GetTestClient(t, ctx)

	tagData := &apstra.DesignTagData{
		Label:       acctest.RandString(10),
		Description: acctest.RandString(10),
	}

	id, err := client.CreateTag(ctx, &apstra.DesignTagRequest{
		Label:       tagData.Label,
		Description: tagData.Description,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTag(ctx, id)) })

	return &apstra.DesignTag{
		Id:   id,
		Data: tagData,
	}
}
