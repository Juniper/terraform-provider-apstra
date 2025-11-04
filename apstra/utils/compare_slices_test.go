package utils

import (
	"slices"
	"sort"
	"testing"
)

func TestDiffSliceSets(t *testing.T) {
	type testCase struct {
		a  []int
		b  []int
		ea []int
		eb []int
	}

	testCases := []testCase{
		{
			a:  []int{},
			b:  []int{},
			ea: []int{},
			eb: []int{},
		},
		{
			a:  []int{1},
			b:  []int{1},
			ea: []int{},
			eb: []int{},
		},
		{
			a:  []int{1},
			b:  []int{2},
			ea: []int{2},
			eb: []int{1},
		},
		{
			a:  []int{},
			b:  []int{1},
			ea: []int{1},
			eb: []int{},
		},
		{
			a:  []int{1},
			b:  []int{},
			ea: []int{},
			eb: []int{1},
		},
		{
			a:  []int{1, 2, 3},
			b:  []int{},
			ea: []int{},
			eb: []int{3, 2, 1},
		},
		{
			a:  []int{},
			b:  []int{1, 2, 3},
			ea: []int{3, 1, 2},
			eb: []int{},
		},
		{
			a:  []int{1, 2, 3},
			b:  []int{7, 8, 9},
			ea: []int{7, 8, 9},
			eb: []int{1, 2, 3},
		},
		{
			a:  []int{1, 2, 3},
			b:  []int{3, 4, 5},
			ea: []int{5, 4},
			eb: []int{2, 1},
		},
	}

	for i, tc := range testCases {
		ra, rb := DiffSliceSets(tc.a, tc.b)
		sort.Ints(ra)
		sort.Ints(rb)
		sort.Ints(tc.ea)
		sort.Ints(tc.eb)
		if !slices.Equal(tc.ea, ra) {
			t.Fatalf("test case %d expected %v, got %v", i, tc.ea, ra)
		}
		if !slices.Equal(tc.eb, rb) {
			t.Fatalf("test case %d expected %v, got %v", i, tc.eb, rb)
		}
	}
}
