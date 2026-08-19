//go:build integration

package random_test

import (
	"testing"

	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/stretchr/testify/require"
)

func TestSomeOfBounds(t *testing.T) {
	testCases := map[string]struct {
		input []string
		args  []uint16
		want  int
	}{
		"no_bounds": {
			input: []string{"a", "b", "c", "d"},
			want:  -1,
		},
		"bounded_within_slice": {
			input: []string{"a", "b", "c", "d"},
			args:  []uint16{2, 3},
			want:  -1,
		},
		"bounded_exact value": {
			input: []string{"a", "b", "c", "d"},
			args:  []uint16{3, 3},
			want:  3,
		},
		"min_above_slice_length_panics": {
			input: []string{"a", "b", "c", "d"},
			args:  []uint16{10},
			want:  -2,
		},
		"max_above_slice_length_clamps": {
			input: []string{"a", "b", "c", "d"},
			args:  []uint16{3, 10},
			want:  3,
		},
		"both_bounds_above_slice_length_panics": {
			input: []string{"a", "b", "c", "d"},
			args:  []uint16{10, 12},
			want:  -2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.want == -2 {
				require.Panics(t, func() { random.SomeOf(tc.input, tc.args...) })
				return
			}

			got := random.SomeOf(tc.input, tc.args...)

			require.LessOrEqual(t, len(got), len(tc.input))
			require.Subset(t, tc.input, got)

			if tc.want >= 0 {
				require.Len(t, got, tc.want)
			}
		})
	}
}
