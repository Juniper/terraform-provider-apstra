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

func LeafSwitchGenericSystemInterfaces(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient) []apstra.ObjectId {
	t.Helper()

	query := new(apstra.PathQuery).
		SetBlueprintId(client.Id()).
		SetClient(client.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal("leaf")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeLink.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal("server")},
			{Key: "role", Value: apstra.QEStringVal("generic")},
		})

	var queryResponse struct {
		Items []struct {
			Interface struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_interface"`
		} `json:"items"`
	}

	require.NoError(t, query.Do(ctx, &queryResponse))

	result := make([]apstra.ObjectId, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		result[i] = item.Interface.Id
	}

	return result
}
