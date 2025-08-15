package blueprint_test

import (
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func TestTrimSetApplicationPointsConnectivityTemplatesRequestBasedOnError(t *testing.T) {
	compare := func(t testing.TB, a, b map[apstra.ObjectId]map[apstra.ObjectId]bool) {
		t.Helper()

		require.Equal(t, len(a), len(b))
		keysA := maps.Keys(a)
		keysB := maps.Keys(b)
		for _, k := range keysA {
			require.Contains(t, keysB, k)
			require.Equal(t, len(a[k]), len(b[k]))
			keysKA := maps.Keys(a[k])
			keysKB := maps.Keys(b[k])
			for _, ka := range keysKA {
				require.Contains(t, keysKB, ka)
				require.Equal(t, a[k][ka], b[k][ka])
			}
		}
	}

	type testCase struct {
		request       map[apstra.ObjectId]map[apstra.ObjectId]bool
		errDetail     apstra.ErrCtAssignmentFailedDetail
		expected      map[apstra.ObjectId]map[apstra.ObjectId]bool
		expectedCount int
	}

	testCases := map[string]testCase{
		"empty": {
			request:       map[apstra.ObjectId]map[apstra.ObjectId]bool{},
			errDetail:     apstra.ErrCtAssignmentFailedDetail{},
			expected:      map[apstra.ObjectId]map[apstra.ObjectId]bool{},
			expectedCount: 0,
		},
		"none_invalid": {
			request: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_1": true,
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			errDetail: apstra.ErrCtAssignmentFailedDetail{},
			expected: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_1": true,
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			expectedCount: 0,
		},
		"one_invalid_ap": {
			request: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": { // <--------    this will be removed
					"ct_id_1": false, // <-- this will be removed
					"ct_id_2": false, // <-- this will be removed
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			errDetail: apstra.ErrCtAssignmentFailedDetail{
				InvalidApplicationPointIds: []apstra.ObjectId{"ap_id_1"},
			},
			expected: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			expectedCount: 3,
		},
		"one_invalid_ct": {
			request: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_1": false, // <-- this will be removed
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			errDetail: apstra.ErrCtAssignmentFailedDetail{
				InvalidConnectivityTemplateIds: []apstra.ObjectId{"ct_id_1"},
			},
			expected: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			expectedCount: 1,
		},
		"one_invalid_ct_used_multiple_times": {
			request: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_1": false, // <-- this will be removed
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_1": false, // <-- this will be removed
				},
			},
			errDetail: apstra.ErrCtAssignmentFailedDetail{
				InvalidConnectivityTemplateIds: []apstra.ObjectId{"ct_id_1"},
			},
			expected: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
				},
			},
			expectedCount: 2,
		},
		"one_invalid_ct_cannot_be_removed": {
			request: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_1": true, // <-- this cannot be removed
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			errDetail: apstra.ErrCtAssignmentFailedDetail{
				InvalidConnectivityTemplateIds: []apstra.ObjectId{"ct_id_1"},
			},
			expected: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_1": true,
					"ct_id_2": false,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			expectedCount: 0,
		},
		"one_invalid_ap_cannot_be_removed": {
			request: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": { // <--------    this cannot be removed
					"ct_id_1": true,  // <-- this cannot be removed
					"ct_id_2": false, // <-- this will be removed
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			errDetail: apstra.ErrCtAssignmentFailedDetail{
				InvalidApplicationPointIds: []apstra.ObjectId{"ap_id_1"},
			},
			expected: map[apstra.ObjectId]map[apstra.ObjectId]bool{
				"ap_id_1": {
					"ct_id_1": true,
				},
				"ap_id_2": {
					"ct_id_3": true,
					"ct_id_4": false,
				},
			},
			expectedCount: 1,
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			i := blueprint.TrimSetApplicationPointsConnectivityTemplatesRequestBasedOnError(tCase.request, &tCase.errDetail)
			require.Equal(t, tCase.expectedCount, i)
			compare(t, tCase.expected, tCase.request)
		})
	}
}
