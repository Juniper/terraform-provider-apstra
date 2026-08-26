//go:build integration

package random_test

import (
	"testing"

	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/stretchr/testify/require"
)

func TestSomeOfBounds(t *testing.T) {
	testCases := map[string]struct {
		input     []string
		args      []uint16
		want      *int
		wantMin   *int
		wantMax   *int
		wantPanic bool
	}{
		"no_bounds": {
			input:   []string{"a", "b", "c", "d"},
			wantMin: pointer.To(0),
			wantMax: pointer.To(4),
		},
		"bounded_within_slice": {
			input:   []string{"a", "b", "c", "d"},
			args:    []uint16{2, 3},
			wantMin: pointer.To(2),
			wantMax: pointer.To(3),
		},
		"bounded_exact value": {
			input: []string{"a", "b", "c", "d"},
			args:  []uint16{3, 3},
			want:  pointer.To(3),
		},
		"min_above_slice_length_panics": {
			input:     []string{"a", "b", "c", "d"},
			args:      []uint16{10},
			wantPanic: true,
		},
		"max_above_slice_length_clamps": {
			input:   []string{"a", "b", "c", "d"},
			args:    []uint16{2, 10},
			wantMin: pointer.To(2),
			wantMax: pointer.To(4),
		},
		"both_bounds_above_slice_length_panics": {
			input:     []string{"a", "b", "c", "d"},
			args:      []uint16{10, 12},
			wantPanic: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.want != nil && (tc.wantMin != nil || tc.wantMax != nil) {
				panic("test case set want along with wantMin or wantMax")
			}
			if tc.wantPanic {
				require.Panics(t, func() { random.SomeOf(tc.input, tc.args...) })
				return
			}

			got := random.SomeOf(tc.input, tc.args...)

			require.LessOrEqual(t, len(got), len(tc.input))
			require.Subset(t, tc.input, got)

			if tc.want != nil {
				require.Len(t, got, *tc.want)
			}
			if tc.wantMin != nil {
				require.LessOrEqual(t, *tc.wantMin, len(got))
			}
			if tc.wantMax != nil {
				require.GreaterOrEqual(t, *tc.wantMax, len(got))
			}
		})
	}
}
