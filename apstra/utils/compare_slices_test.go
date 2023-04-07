package utils

import "testing"

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
