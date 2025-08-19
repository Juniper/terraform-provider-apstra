package testutils

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func DatacenterConnectivityTemplateA(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient, szId apstra.ObjectId, tag int) apstra.ObjectId {
	t.Helper()

	ct := apstra.ConnectivityTemplate{
		Id:          nil,
		Label:       acctest.RandString(10),
		Description: acctest.RandString(10),
		Subpolicies: []*apstra.ConnectivityTemplatePrimitive{
			{
				Label: acctest.RandString(10),
				Attributes: &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
					Label:              acctest.RandString(10),
					SecurityZone:       &szId,
					Tagged:             true,
					Vlan:               utils.ToPtr(apstra.Vlan(tag)),
					IPv4AddressingType: apstra.CtPrimitiveIPv4AddressingTypeNumbered,
				},
			},
		},
	}

	require.NoError(t, ct.SetIds())
	require.NoError(t, ct.SetUserData())
	require.NoError(t, bp.CreateConnectivityTemplate(ctx, &ct))

	return *ct.Id
}

func DatacenterIpLinkConnectivityTemplateVlans(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient, ctId apstra.ObjectId) []int {
	t.Helper()

	query := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeEpEndpointPolicy.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(ctId)},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpSubpolicy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpEndpointPolicy.QEEAttribute()}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpFirstSubpolicy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeEpEndpointPolicy.QEEAttribute(),
			{Key: "policy_type_name", Value: apstra.QEStringVal("AttachLogicalLink")},
			{Key: "name", Value: apstra.QEStringVal("n_ep_endpoint_policy")},
		})

	var queryResponse struct {
		Items []struct {
			EpEndpointPolicy struct {
				Attributes json.RawMessage `json:"attributes"`
			} `json:"n_ep_endpoint_policy"`
		} `json:"items"`
	}

	require.NoError(t, query.Do(ctx, &queryResponse))
	require.Greater(t, len(queryResponse.Items), 0)

	result := make([]int, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		var rawString string
		require.NoError(t, json.Unmarshal(item.EpEndpointPolicy.Attributes, &rawString))

		var attributes struct {
			VlanId int `json:"vlan_id"`
		}
		require.NoError(t, json.Unmarshal([]byte(rawString), &attributes))
		result[i] = attributes.VlanId
	}

	return result
}
