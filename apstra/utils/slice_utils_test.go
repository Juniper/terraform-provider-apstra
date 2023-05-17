package utils

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			t.Fatalf("test case %d produced %t, expected %t", i, result, tc.expected)
		}
	}
}

func TestUniq(t *testing.T) {
	type testCase struct {
		t []any
		e []any
	}

	testCases := []testCase{
		{
			t: []any{},
			e: []any{},
		},
		{
			t: []any{"foo", "bar", "baz"},
			e: []any{"foo", "bar", "baz"},
		},
		{
			t: []any{"foo", "bar", "baz", "baz"},
			e: []any{"foo", "bar", "baz"},
		},

		{
			t: []any{"foo", "bar", "foo", "baz"},
			e: []any{"baz", "bar", "foo"},
		},
		{
			t: []any{1, 2, 3},
			e: []any{1, 2, 3},
		},
		{
			t: []any{1, 1, 2, 3},
			e: []any{3, 1, 2},
		},
	}

	for i, tc := range testCases {
		r := Uniq(tc.t)
		if !SlicesMatch(r, tc.e) {
			t.Fatalf("test case %d, expected %v, got %v", i, tc.e, r)
		}
	}
}

func TestElementsFromANotInB(t *testing.T) {
	type testCase struct {
		a []any // slice 'a' which may have extra members
		b []any // slice 'b' is the baseline against which 'a' is compared
		e []any // slice 'e' is the expected result
	}

	testCases := []testCase{
		{
			a: []any{1, 2, 3},
			b: []any{1, 2, 3},
			e: []any{},
		},
		{
			a: []any{1, 2, 3},
			b: []any{1, 2, 3, 4},
			e: []any{},
		},
		{
			a: []any{1, 2, 3, 2},
			b: []any{1, 2, 3},
			e: []any{},
		},
		{
			a: []any{1, 2, 5, 2},
			b: []any{1, 2, 3},
			e: []any{5},
		},
		{
			a: []any{"a", "b", "c"},
			b: []any{"a", "b", "c"},
			e: []any{},
		},
		{
			a: []any{"a", "d", "c"},
			b: []any{"a", "b", "c"},
			e: []any{"d"},
		},
	}

	for i, tc := range testCases {
		r := UniqueElementsFromA(tc.a, tc.b)
		if !SlicesMatch(tc.e, r) {
			t.Fatalf("test case %d, expectd %v, got %v", i, tc.e, r)
		}
	}
}

func TestUniqStringers(t *testing.T) {
	type testCase struct {
		t []fmt.Stringer
		e []fmt.Stringer
	}

	testCases := []testCase{
		{
			t: []fmt.Stringer{},
			e: []fmt.Stringer{},
		},
		{
			t: []fmt.Stringer{types.StringValue("foo"), types.StringValue("bar"), types.StringValue("baz")},
			e: []fmt.Stringer{types.StringValue("foo"), types.StringValue("bar"), types.StringValue("baz")},
		},
		{
			t: []fmt.Stringer{types.StringValue("foo"), types.StringValue("bar"), types.StringValue("baz"), types.StringValue("baz")},
			e: []fmt.Stringer{types.StringValue("foo"), types.StringValue("bar"), types.StringValue("baz")},
		},
		{
			t: []fmt.Stringer{types.StringValue("foo"), types.StringValue("bar"), types.StringValue("foo"), types.StringValue("baz")},
			e: []fmt.Stringer{types.StringValue("baz"), types.StringValue("bar"), types.StringValue("foo")},
		},
	}

	for i, tc := range testCases {
		r := UniqStringers(tc.t)
		if !SliceStringersMatch(r, tc.e) {
			t.Fatalf("test case %d, expected %v, got %v", i, tc.e, r)
		}
	}
}
