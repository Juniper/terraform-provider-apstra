package blueprint

import (
	"context"
	"errors"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"sort"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"terraform-provider-apstra/apstra/utils"
	"testing"
)

func TestAccessSwitchIdsToParentLeafIds(t *testing.T) {
	ctx := context.Background()
	bpClient, cleanup, err := testutils.BlueprintC(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cleanup(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	type node struct {
		Id   string `json:"id"`
		Role string `json:"role"`
	}
	getNodesResposne := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err = bpClient.Client().GetNodes(ctx, bpClient.Id(), apstra.NodeTypeSystem, getNodesResposne)
	if err != nil {
		t.Fatal(errors.Join(cleanup(ctx), err))
	}

	var accessSwitchIds []string
	for _, n := range getNodesResposne.Nodes {
		if n.Role == "access" {
			accessSwitchIds = append(accessSwitchIds, n.Id)
		}
	}

	result, err := accessSwitchIdsToParentLeafIds(ctx, accessSwitchIds, bpClient)
	if err != nil {
		t.Fatal(err)
	}

	// resultData and expectedData are map[int]int representing count of access
	// switches keyed by that parent switch count.
	resultData := make(map[int]int)
	for _, v := range result {
		parentCount := len(v)
		resultData[parentCount]++
	}

	// expecting 3 switches to have 1 parent and 4 switch to have 2 parents
	expectedData := map[int]int{
		1: 3,
		2: 4,
	}

	if !utils.MapsMatch(expectedData, resultData) {
		t.Fatalf("expected %v, got %v", expectedData, resultData)
	}
}

func TestRedunancyGroupIdToRedundancyGroupInfo(t *testing.T) {
	ctx := context.Background()
	bpClient, cleanup, err := testutils.BlueprintD(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cleanup(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	type node struct {
		Id   string `json:"id"`
		Role string `json:"role"`
	}
	getNodesResposne := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err = bpClient.Client().GetNodes(ctx, bpClient.Id(), apstra.NodeTypeSystem, getNodesResposne)
	if err != nil {
		t.Fatal(errors.Join(cleanup(ctx), err))
	}

	var accessNodes, leafNodes []string
	for _, n := range getNodesResposne.Nodes {
		switch n.Role {
		case "access":
			accessNodes = append(accessNodes, n.Id)
		case "leaf":
			leafNodes = append(leafNodes, n.Id)
		}
	}
	sort.Strings(accessNodes)
	sort.Strings(leafNodes)

	rgMap, err := redunancyGroupIdToRedundancyGroupInfo(ctx, bpClient)
	if err != nil {
		t.Fatal(err)
	}

	expectedLen := 2
	if len(rgMap) != expectedLen {
		t.Fatalf("expected %d redundancy groups, got %d", expectedLen, len(rgMap))
	}

	for _, rg := range rgMap {
		sort.Strings(rg.memberIds)
		switch rg.role {
		case "access":
			if !utils.SlicesMatch(accessNodes, rg.memberIds) {
				t.Fatalf("access nodes: expected %v, got %v", accessNodes, rg.memberIds)
			}
		case "leaf":
			if !utils.SlicesMatch(leafNodes, rg.memberIds) {
				t.Fatalf("leaf nodes: expected %v, got %v", leafNodes, rg.memberIds)
			}
		}
	}
}

func TestGetSystemRoles(t *testing.T) {
	ctx := context.Background()
	bpClient, cleanup, err := testutils.BlueprintD(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cleanup(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	type node struct {
		Id   string `json:"id"`
		Role string `json:"role"`
	}
	getNodesResposne := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err = bpClient.Client().GetNodes(ctx, bpClient.Id(), apstra.NodeTypeSystem, getNodesResposne)
	if err != nil {
		t.Fatal(errors.Join(cleanup(ctx), err))
	}

	var systemIds []string
	expected := make(map[string]apstra.SystemRole)
	for _, n := range getNodesResposne.Nodes {
		systemIds = append(systemIds, n.Id)
		var role apstra.SystemRole
		err = role.FromString(n.Role)
		if err != nil {
			t.Fatal(err)
		}
		expected[n.Id] = role
	}

	result, err := getSystemRoles(ctx, systemIds, bpClient)
	if err != nil {
		t.Fatal(err)
	}

	if !utils.MapsMatch(expected, result) {
		t.Fatalf("expected: %v, got: %v", expected, result)
	}
}

//func TestSwitchIdsToBindings(t *testing.T) {
//	ctx := context.Background()
//	bpClient, cleanup, err := testutils.BlueprintE(ctx)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer func() {
//		err := cleanup(ctx)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}()
//
//	bindings, err := switchIdsToBindings(ctx, []string{}, nil, bpClient)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if len(bindings) != 0 {
//		t.Fatalf("expected 0 bindings, got %d bindings", len(bindings))
//	}
//
//	type node struct {
//		Id         string `json:"id"`
//		Role       string `json:"role"`
//		GroupLabel string `json:"group_label"`
//	}
//	getNodesResposne := &struct {
//		Nodes map[string]node `json:"nodes"`
//	}{}
//	err = bpClient.Client().GetNodes(ctx, bpClient.Id(), apstra.NodeTypeSystem, getNodesResposne)
//	if err != nil {
//		t.Fatal(errors.Join(cleanup(ctx), err))
//	}
//
//	var accessIds, leafIds []string
//	for _, n := range getNodesResposne.Nodes {
//		switch n.Role {
//		case apstra.SystemRoleAccess.String():
//			accessIds = append(accessIds, n.Id)
//		case apstra.SystemRoleLeaf.String():
//			leafIds = append(leafIds, n.Id)
//		}
//	}
//
//	bindings, err = switchIdsToBindings(ctx, accessIds, nil, bpClient)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if len(bindings) != 4 {
//		t.Fatalf("expected 4 bindings, got %d bindings", len(bindings))
//	}
//
//	bindings, err = switchIdsToBindings(ctx, append(leafIds, accessIds...), nil, bpClient)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if len(bindings) != 4 {
//		t.Fatalf("expected 4 bindings, got %d bindings", len(bindings))
//	}
//
//	var boundAccessIdCount int
//	for _, binding := range bindings {
//		boundAccessIdCount += len(binding.AccessSwitchNodeIds)
//	}
//	if boundAccessIdCount != 8 {
//		t.Fatalf("expected 5 access bindings, got %d access bindings", boundAccessIdCount)
//	}
//
//	type testCase struct {
//		groupLabelMatch  []string
//		expectedBidnings int
//		expectedAccess   int
//	}
//	testCases := []testCase{
//		{
//			groupLabelMatch:  []string{"rack type B access 2"},
//			expectedBidnings: 1,
//			expectedAccess:   2,
//		},
//		{
//			groupLabelMatch:  []string{"rack type B access 1", "rack type B leaf"},
//			expectedBidnings: 1,
//			expectedAccess:   1,
//		},
//		{
//			groupLabelMatch:  []string{"rack type B access 1", "rack type B access 2", "rack type B leaf"},
//			expectedBidnings: 1,
//			expectedAccess:   3,
//		},
//		{
//			groupLabelMatch: []string{
//				"rack type A access",
//				"rack type B access 1", "rack type B access 2", "rack type B leaf",
//			},
//			expectedBidnings: 2,
//			expectedAccess:   4,
//		},
//		{
//			groupLabelMatch: []string{
//				"rack type A access", "rack type A leaf",
//				"rack type B access 1", "rack type B access 2", "rack type B leaf",
//			},
//			expectedBidnings: 2,
//			expectedAccess:   4,
//		},
//		{
//			groupLabelMatch: []string{
//				"rack type A access", "rack type A leaf",
//				"rack type B access 1", "rack type B access 2", "rack type B leaf",
//				"rack type D access 1", "rack type D access 2", "rack type D access 3",
//			},
//			expectedBidnings: 3,
//			expectedAccess:   7,
//		},
//		{
//			groupLabelMatch: []string{
//				"rack type A leaf",
//				"rack type B leaf",
//				"rack type C leaf",
//				"rack type D leaf",
//			},
//			expectedBidnings: 4,
//			expectedAccess:   0,
//		},
//	}
//
//	for i, tc := range testCases {
//		var ids []string
//		for _, n := range getNodesResposne.Nodes {
//			if utils.SliceContains(n.GroupLabel, tc.groupLabelMatch) {
//				ids = append(ids, n.Id)
//			}
//		}
//		bindings, err = switchIdsToBindings(ctx, ids, nil, bpClient)
//		if err != nil {
//			t.Fatal(err)
//		}
//		if len(bindings) != tc.expectedBidnings {
//			t.Fatalf("interation %d, expected %d bindings, got %d", i, tc.expectedBidnings, len(bindings))
//		}
//
//		accessBindings := 0
//		for i := range bindings {
//			accessBindings += len(bindings[i].AccessSwitchNodeIds)
//		}
//		if accessBindings != tc.expectedAccess {
//			t.Fatalf("instance %d expected %d access bindings, got %d access bindings", i, tc.expectedAccess, accessBindings)
//		}
//	}
//}
