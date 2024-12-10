package customtypes_test

import (
	"context"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStringWithAltValuesType_ValueFromTerraform(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in          tftypes.Value
		expectation attr.Value
		expectedErr string
	}{
		"true": {
			in:          tftypes.NewValue(tftypes.String, "foo"),
			expectation: customtypes.NewStringWithAltValuesValue("foo"),
		},
		"unknown": {
			in:          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: customtypes.NewStringWithAltValuesUnknown(),
		},
		"null": {
			in:          tftypes.NewValue(tftypes.String, nil),
			expectation: customtypes.NewStringWithAltValuesNull(),
		},
		"wrongType": {
			in:          tftypes.NewValue(tftypes.Number, 123),
			expectedErr: "can't unmarshal tftypes.Number into *string, expected string",
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := customtypes.StringWithAltValuesType{}.ValueFromTerraform(ctx, tCase.in)
			if tCase.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tCase.expectedErr, err.Error())
				return
			}

			require.Truef(t, got.Equal(tCase.expectation), "values not equal %s, %s", tCase.expectation, got)
		})
	}

}
