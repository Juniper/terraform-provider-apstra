package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"testing"
)

func TestDiagnosticWrapper(t *testing.T) {
	type testCase struct {
		d diag.Diagnostic // diagnostic to be wrapped
		w []string        // wrap messages
		e string          // expected detail output
	}

	testCases := []testCase{
		{
			d: diag.NewErrorDiagnostic("summary", "detail"),
			w: []string{"additional context 1", "additional context 2"},
			e: "additional context 2 : additional context 1 : detail",
		},
		{
			d: diag.NewErrorDiagnostic("summary", "detail"),
			w: []string{},
			e: "detail",
		},
		{
			d: diag.NewErrorDiagnostic("summary", "detail"),
			w: nil,
			e: "detail",
		},
	}

	for i, tc := range testCases {
		w := tc.d
		for j := range tc.w {
			w = WrapDiagnostic(w, tc.w[j])
		}
		if !w.Equal(tc.d) {
			t.Fatalf("test case %d failed equality: %v vs. %v", i, tc.d, w)
		}
		if tc.e != w.Detail() {
			t.Fatalf("test case %d expected %q, got %q", i, tc.e, w.Detail())
		}
	}
}
