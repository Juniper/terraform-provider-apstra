package pointer_test

import (
	"reflect"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/stretchr/testify/require"
)

func TestConvertInteger(t *testing.T) {
	t.Parallel()
	type testCase struct {
		call func() any
		want any
	}

	tests := map[string]testCase{
		"int32_to_int64": {
			call: func() any {
				return pointer.ConvertInteger(new(int64), pointer.To(int32(123)))
			},
			want: new(int64(123)),
		},
		"int64_to_int32": {
			call: func() any {
				return pointer.ConvertInteger(new(int32), pointer.To(int64(42)))
			},
			want: new(int32(42)),
		},
		"overflow_int16_to_int8": {
			call: func() any {
				return pointer.ConvertInteger(new(int8), pointer.To(int16(128)))
			},
			want: new(int8(-128)),
		},
		"overflow_uint16_to_uint8": {
			call: func() any {
				return pointer.ConvertInteger(new(uint8), pointer.To(uint16(256)))
			},
			want: new(uint8(0)),
		},
		"nil_int32_to_int8": {
			call: func() any {
				return pointer.ConvertInteger(new(int8), (*int32)(nil))
			},
			want: (*int8)(nil),
		},
		"negative_overflow_int16_to_int8": {
			call: func() any {
				return pointer.ConvertInteger(new(int8), pointer.To(int16(-129)))
			},
			want: new(int8(127)),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.want == nil {
				panic("this should never happen with proper test data. " +
					"The interface itself should never be nil. It'll always " +
					"be a (possibly nil) pointer to some type.")

			}

			got := tc.call()
			require.IsType(t, tc.want, got)

			if reflect.ValueOf(tc.want).IsNil() {
				require.Truef(t, reflect.ValueOf(got).IsNil(), "expected nil, got %v (%T)", got, got)
				return
			}

			require.NotNil(t, got, "expected non-nil result")
			require.Equalf(t, tc.want, got, "expected value %v, got %v", tc.want, got)
		})
	}
}
