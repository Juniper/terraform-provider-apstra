package utils

import (
	"testing"
)

func TestItemInSlice(t *testing.T) {
	type testCase struct {
		item     any
		slice    []any
		expected bool
	}

	testCases := []testCase{
		{item: 1, slice: []any{1, 2, 3}, expected: true},
		{item: 1, slice: []any{1, 2, 3, 1}, expected: true},
		{item: 1, slice: []any{3, 2, 1}, expected: true},
		{item: 0, slice: []any{1, 2, 3}, expected: false},
		{item: 0, slice: []any{}, expected: false},
		{item: 1, slice: []any{}, expected: false},
		{item: "foo", slice: []any{"foo", "bar"}, expected: true},
		{item: "foo", slice: []any{"bar", "foo"}, expected: true},
		{item: "foo", slice: []any{"foo", "bar", "foo"}, expected: true},
		{item: "foo", slice: []any{"bar", "baz"}, expected: false},
		{item: "foo", slice: []any{""}, expected: false},
		{item: "foo", slice: []any{"", ""}, expected: false},
		{item: "foo", slice: []any{}, expected: false},
		{item: "", slice: []any{"bar", "foo"}, expected: false},
		{item: "", slice: []any{"bar", "", "foo"}, expected: true},
		{item: "", slice: []any{}, expected: false},
	}

	var result bool
	for i, tc := range testCases {
		switch tc.item.(type) {
		case int:
			item := tc.item.(int)
			slice := make([]int, len(tc.slice))
			for j := range tc.slice {
				slice[j] = tc.slice[j].(int)
			}
			result = ItemInSlice(item, slice)
		case string:
			item := tc.item.(string)
			slice := make([]string, len(tc.slice))
			for j := range tc.slice {
				slice[j] = tc.slice[j].(string)
			}
			result = ItemInSlice(item, slice)
		}
		if result != tc.expected {
			t.Fatalf("test case %d produced %t, epexcted %t", i, result, tc.expected)
		}
	}
}
