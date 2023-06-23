package utils

import (
	"sort"
	"testing"
)

func TestSlicesMatch(t *testing.T) {

	type intTestCase struct {
		a        []int
		b        []int
		expected bool
	}

	type stringTestCase struct {
		a        []string
		b        []string
		expected bool
	}

	intTestCases := []intTestCase{
		{
			a:        []int{1, 3, 4},
			b:        []int{1, 3, 4},
			expected: true,
		},
		{
			a:        []int{1, 3, 4, 5},
			b:        []int{1, 3, 4},
			expected: false,
		},
		{
			a:        []int{1, 3, 4},
			b:        []int{1, 3, 4, 5},
			expected: false,
		},
		{
			a:        []int{0, 3, 4},
			b:        []int{1, 3, 4},
			expected: false,
		},
		{
			a:        []int{1, 3, 0},
			b:        []int{1, 3, 4},
			expected: false,
		},
	}

	for i, tc := range intTestCases {
		result := SlicesMatch(tc.a, tc.b)
		if tc.expected != result {
			t.Fatalf("int test case %d: expected %t; got %t", i, tc.expected, result)
		}

	}

	stringTestCases := []stringTestCase{
		{
			a:        []string{"foo", "bar"},
			b:        []string{"foo", "bar"},
			expected: true,
		},
		{
			a:        []string{"fOo", "bar"},
			b:        []string{"foo", "bar"},
			expected: false,
		},
		{
			a:        []string{"foo", "bar", "baz"},
			b:        []string{"foo", "bar"},
			expected: false,
		},
	}

	for i, tc := range stringTestCases {
		result := SlicesMatch(tc.a, tc.b)
		if tc.expected != result {
			t.Fatalf("string test case %d: expected %t; got %t", i, tc.expected, result)
		}

	}
}

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
		if !SlicesMatch(tc.ea, ra) {
			t.Fatalf("test case %d expected %v, got %v", i, tc.ea, ra)
		}
		if !SlicesMatch(tc.eb, rb) {
			t.Fatalf("test case %d expected %v, got %v", i, tc.eb, rb)
		}
	}
}
