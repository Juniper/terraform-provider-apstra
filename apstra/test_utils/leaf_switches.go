package testutils

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/stretchr/testify/require"
)

func leafSwitches(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient) []apstra.ObjectId {
	query := new(apstra.PathQuery).
		SetBlueprintId(client.Id()).
		SetClient(client.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal("leaf")},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	var queryResponse struct {
		Items []struct {
			System struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	require.NoError(t, query.Do(ctx, &queryResponse))

	result := make([]apstra.ObjectId, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		result[i] = item.System.Id
	}

	return result
}
