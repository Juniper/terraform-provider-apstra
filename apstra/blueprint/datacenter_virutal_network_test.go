package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"sort"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"terraform-provider-apstra/apstra/utils"
	"testing"
)

func TestGetSystemRoles(t *testing.T) {
	ctx := context.Background()

	client, err := testutils.GetTestClient()
	if err != nil {
		t.Fatal(err)
	}

	blueprintId := apstra.ObjectId("118ba0be-21d3-4eb3-bcc2-4aad4a65af57")
	bpClient, err := client.NewTwoStageL3ClosClient(ctx, blueprintId)
	if err != nil {
		t.Fatal(err)
	}

	response := new(struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	})
	err = new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeSystem.String())},
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
			{Key: "role", Value: apstra.QEStringValIsIn{"access", "leaf"}},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		}).
		SetClient(client).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bpClient.Id()).
		Do(ctx, response)
	if err != nil {
		t.Fatal(err)
	}

	switchIds := make([]string, len(response.Items))
	for i, item := range response.Items {
		switchIds[i] = item.System.Id
	}
	_, err = switchIdsToBindings(ctx, switchIds, bpClient)

	//gsr, err := getSystemRoles(ctx, []string{"-ZlNS3vcKS2tfJpYDEA"}, bpClient)
	if err != nil {
		t.Fatal(err)
	}

	//_ = gsr
}

func TestRedundancyPeersFromIds(t *testing.T) {
	g1 := redundancyGroup{
		id:        "ag1",
		role:      "access",
		memberIds: []string{"a11", "a12"},
	}
	g2 := redundancyGroup{
		id:        "ag2",
		role:      "access",
		memberIds: []string{"a21", "a22"},
	}
	td := map[string]*redundancyGroup{
		"a11": &g1,
		"a12": &g1,
		"a21": &g2,
		"a22": &g2,
	}

	type testCase struct {
		i []string // input
		e []string // expected
	}

	testCases := []testCase{
		{
			i: []string{"a21"},
			e: []string{"a21", "a22"},
		},
		{
			i: []string{"a21", "a22"},
			e: []string{"a21", "a22"},
		},
		{
			i: []string{"a11", "a22"},
			e: []string{"a11", "a12", "a21", "a22"},
		},
	}

	for i, tc := range testCases {
		r := redundancyPeersFromIds(tc.i, td)
		sort.Strings(r)
		if !utils.SlicesMatch(tc.e, r) {
			t.Fatalf("test case %d, expected %v, got %v", i, tc.e, r)
		}
	}
}

func TestAccessSwitchIdsToParentLeafIds(t *testing.T) {
	ctx := context.Background()
	accessSwitchIdsToParentLeafIds(ctx, []string{
		"sgo2busZOonZwKHmOJg",
		"BfYGZuMOphU6vYkJDqk",
		"cncW6mDsaKymcOR4FxM",
		"NKs3uQLJHIN0fWVVDJk",
		"1GzZ4F7Aq13n-8xZFI8",
	}, nil)
}
