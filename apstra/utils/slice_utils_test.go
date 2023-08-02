package utils

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestItemInSlice(t *testing.T) {
	type testCase[T comparable] struct {
		item     T
		slice    []T
		expected bool
	}

	intTestCases := []testCase[int]{
		{item: 1, slice: []int{1, 2, 3}, expected: true},
		{item: 1, slice: []int{1, 2, 3, 1}, expected: true},
		{item: 1, slice: []int{3, 2, 1}, expected: true},
		{item: 0, slice: []int{1, 2, 3}, expected: false},
		{item: 0, slice: []int{}, expected: false},
		{item: 1, slice: []int{}, expected: false},
	}
	for i, tc := range intTestCases {
		result := ItemInSlice(tc.item, tc.slice)
		if result != tc.expected {
			t.Fatalf("int test case %d produced %t, expected %t", i, result, tc.expected)
		}
	}

	stringTestCases := []testCase[string]{
		{item: "foo", slice: []string{"foo", "bar"}, expected: true},
		{item: "foo", slice: []string{"bar", "foo"}, expected: true},
		{item: "foo", slice: []string{"foo", "bar", "foo"}, expected: true},
		{item: "foo", slice: []string{"bar", "baz"}, expected: false},
		{item: "foo", slice: []string{""}, expected: false},
		{item: "foo", slice: []string{"", ""}, expected: false},
		{item: "foo", slice: []string{}, expected: false},
		{item: "", slice: []string{"bar", "foo"}, expected: false},
		{item: "", slice: []string{"bar", "", "foo"}, expected: true},
		{item: "", slice: []string{}, expected: false},
	}
	for i, tc := range stringTestCases {
		result := ItemInSlice(tc.item, tc.slice)
		if result != tc.expected {
			t.Fatalf("string test case %d produced %t, expected %t", i, result, tc.expected)
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

func TestSwap(t *testing.T) {
	type testCase struct {
		t    []any
		e    []any
		swap []int
	}

	testCases := []testCase{
		{
			t:    []any{"a", "b", "c"},
			e:    []any{"a", "c", "b"},
			swap: []int{1, 2},
		},
		{
			t:    []any{"a", "b", "c"},
			e:    []any{"b", "a", "c"},
			swap: []int{0, 1},
		},
		{
			t:    []any{"a", "b", "c"},
			e:    []any{"a", "b", "c"},
			swap: []int{1, 1},
		},
		{
			t:    []any{5, 6, 7},
			e:    []any{7, 6, 5},
			swap: []int{0, 2},
		},
		{
			t:    []any{5, 6, 7},
			e:    []any{7, 6, 5},
			swap: []int{2, 0},
		},
	}

	for i, tc := range testCases {
		Swap(tc.swap[0], tc.swap[1], tc.t)
		if !SlicesMatch(tc.t, tc.e) {
			t.Fatalf("test case %d, expected %v got %v", i, tc.e, tc.t)
		}
	}
}

func TestRevers(t *testing.T) {
	type testCase struct {
		t []any
		e []any
	}

	testCases := []testCase{
		{
			t: []any{"a", "b", "c"},
			e: []any{"c", "b", "a"},
		},
		{
			t: []any{"a", "b", "c", "d"},
			e: []any{"d", "c", "b", "a"},
		},
		{
			t: []any{4, 5, 6},
			e: []any{6, 5, 4},
		},
		{
			t: []any{4, 5, 6, 7},
			e: []any{7, 6, 5, 4},
		},
	}

	for i, tc := range testCases {
		Reverse(tc.t)
		if !SlicesMatch(tc.t, tc.e) {
			t.Fatalf("test case %d, expected %v got %v", i, tc.e, tc.t)
		}
	}
}
