package tfapstra

import (
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"golang.org/x/exp/constraints"
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

func notAllZeros[A constraints.Integer](in []A) bool {
	for i := range in {
		if in[i] != 0 {
			return true
		}
	}
	return false
}

func TestRandomIntegers(t *testing.T) {
	type uFoo uint16
	type sFoo int16

	dataI := make([]int, 50)
	dataS8 := make([]uint8, 50)
	dataS16 := make([]uint16, 50)
	dataS32 := make([]uint32, 50)
	dataS64 := make([]uint64, 50)
	dataU8 := make([]uint8, 50)
	dataU16 := make([]uint16, 50)
	dataU32 := make([]uint32, 50)
	dataU64 := make([]uint64, 50)
	dataSFoo := make([]sFoo, 50)
	dataUFoo := make([]uFoo, 50)

	FillWithRandomIntegers(dataI)
	FillWithRandomIntegers(dataS8)
	FillWithRandomIntegers(dataS16)
	FillWithRandomIntegers(dataS32)
	FillWithRandomIntegers(dataS64)
	FillWithRandomIntegers(dataU8)
	FillWithRandomIntegers(dataU16)
	FillWithRandomIntegers(dataU32)
	FillWithRandomIntegers(dataU64)
	FillWithRandomIntegers(dataSFoo)
	FillWithRandomIntegers(dataUFoo)

	if !notAllZeros(dataI) {
		t.Fail()
	}
	if !notAllZeros(dataS8) {
		t.Fail()
	}
	if !notAllZeros(dataS16) {
		t.Fail()
	}
	if !notAllZeros(dataS32) {
		t.Fail()
	}
	if !notAllZeros(dataS64) {
		t.Fail()
	}
	if !notAllZeros(dataU8) {
		t.Fail()
	}
	if !notAllZeros(dataU16) {
		t.Fail()
	}
	if !notAllZeros(dataU32) {
		t.Fail()
	}
	if !notAllZeros(dataU64) {
		t.Fail()
	}
	if !notAllZeros(dataSFoo) {
		t.Fail()
	}
	if !notAllZeros(dataUFoo) {
		t.Fail()
	}
}
