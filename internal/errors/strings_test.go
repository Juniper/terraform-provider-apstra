package errors_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestCreateError(t *testing.T) {
	type testObj1 struct{}

	type testCase struct {
		t any
		e string
	}

	testCases := map[string]testCase{
		"test_obj_1": {
			t: testObj1{},
			e: "Failed to create testObj1",
		},
		"test_context": {
			t: context.Background(),
			e: "Failed to create backgroundCtx",
		},
		"test_bytes_buffer_ptr": {
			t: new(bytes.Buffer),
			e: "Failed to create Buffer",
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			r := errors.CreateError(tCase.t)
			require.Equal(t, tCase.e, r)
		})
	}
}

func TestReadError(t *testing.T) {
	type testObj1 struct{}

	type testCase struct {
		t any
		e string
	}

	testCases := map[string]testCase{
		"test_obj_1": {
			t: testObj1{},
			e: "Failed to read testObj1",
		},
		"test_context": {
			t: context.Background(),
			e: "Failed to read backgroundCtx",
		},
		"test_bytes_buffer_ptr": {
			t: new(bytes.Buffer),
			e: "Failed to read Buffer",
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			r := errors.ReadError(tCase.t)
			require.Equal(t, tCase.e, r)
		})
	}
}

func TestUpdateError(t *testing.T) {
	type testObj1 struct{}

	type testCase struct {
		t any
		e string
	}

	testCases := map[string]testCase{
		"test_obj_1": {
			t: testObj1{},
			e: "Failed to update testObj1",
		},
		"test_context": {
			t: context.Background(),
			e: "Failed to update backgroundCtx",
		},
		"test_bytes_buffer_ptr": {
			t: new(bytes.Buffer),
			e: "Failed to update Buffer",
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			r := errors.UpdateError(tCase.t)
			require.Equal(t, tCase.e, r)
		})
	}
}

func TestDeleteError(t *testing.T) {
	type testObj1 struct{}

	type testCase struct {
		t any
		e string
	}

	testCases := map[string]testCase{
		"test_obj_1": {
			t: testObj1{},
			e: "Failed to delete testObj1",
		},
		"test_context": {
			t: context.Background(),
			e: "Failed to delete backgroundCtx",
		},
		"test_bytes_buffer_ptr": {
			t: new(bytes.Buffer),
			e: "Failed to delete Buffer",
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			r := errors.DeleteError(tCase.t)
			require.Equal(t, tCase.e, r)
		})
	}
}
