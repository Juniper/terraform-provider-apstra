package tfapstra

import (
	"terraform-provider-apstra/apstra/utils"
	"testing"
)

func TestSliceWithoutElement(t *testing.T) {
	type intTestCase struct {
		in              []int
		e               int
		expectedSlice   []int
		expectedRemoved int
	}

	type stringTestCase struct {
		in              []string
		e               string
		expectedSlice   []string
		expectedRemoved int
	}

	testCases := []any{
		intTestCase{
			in:              []int{0, 1, 2, 3, 4},
			e:               1,
			expectedSlice:   []int{0, 2, 3, 4},
			expectedRemoved: 1,
		},
		intTestCase{
			in:              []int{0, 1, 2, 3, 4},
			e:               5,
			expectedSlice:   []int{0, 1, 2, 3, 4},
			expectedRemoved: 0,
		},
		intTestCase{
			in:              []int{0, 1, 1, 1, 4},
			e:               1,
			expectedSlice:   []int{0, 4},
			expectedRemoved: 3,
		},
		intTestCase{
			in:              []int{1, 1, 1, 1, 1},
			e:               1,
			expectedSlice:   []int{},
			expectedRemoved: 5,
		},
		stringTestCase{
			in:              []string{"a", "b", "c", "d", "e"},
			e:               "b",
			expectedSlice:   []string{"a", "c", "d", "e"},
			expectedRemoved: 1,
		},
		stringTestCase{
			in:              []string{"a", "b", "c", "d", "e"},
			e:               "f",
			expectedSlice:   []string{"a", "b", "c", "d", "e"},
			expectedRemoved: 0,
		},
		stringTestCase{
			in:              []string{"a", "b", "b", "b", "e"},
			e:               "b",
			expectedSlice:   []string{"a", "e"},
			expectedRemoved: 3,
		},
		stringTestCase{
			in:              []string{"a", "a", "a", "a", "a"},
			e:               "a",
			expectedSlice:   []string{},
			expectedRemoved: 5,
		},
	}

	for i := range testCases {
		if tc, ok := testCases[i].(intTestCase); ok {
			result, removed := sliceWithoutElement(tc.in, tc.e)
			if !utils.SlicesMatch(tc.expectedSlice, result) {
				t.Fatalf("expected: %v\ngot:      %v", tc.expectedSlice, result)
			}
			if tc.expectedRemoved != removed {
				t.Fatalf("expected %d removals, got %d removals", tc.expectedRemoved, removed)
			}
			continue
		}
		if tc, ok := testCases[i].(stringTestCase); ok {
			result, removed := sliceWithoutElement(tc.in, tc.e)
			if !utils.SlicesMatch(tc.expectedSlice, result) {
				t.Fatalf("expected: %v\ngot:      %v", tc.expectedSlice, result)
			}
			if tc.expectedRemoved != removed {
				t.Fatalf("expected %d removals, got %d removals", tc.expectedRemoved, removed)
			}
			continue
		}
	}
}
