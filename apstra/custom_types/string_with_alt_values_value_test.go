package customtypes_test

import (
	"context"
	"testing"

	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestStringWithAltValues_StringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentValue  customtypes.StringWithAltValues
		givenValue    basetypes.StringValuable
		expectedMatch bool
	}{
		"equal - no alt values": {
			currentValue:  customtypes.NewStringWithAltValuesValue("foo"),
			givenValue:    customtypes.NewStringWithAltValuesValue("foo"),
			expectedMatch: true,
		},
		"equal - with alt values": {
			currentValue:  customtypes.NewStringWithAltValuesValue("foo", "bar", "baz"),
			givenValue:    customtypes.NewStringWithAltValuesValue("foo"),
			expectedMatch: true,
		},
		"semantically equal - given matches an alt value": {
			currentValue:  customtypes.NewStringWithAltValuesValue("foo", "bar", "baz", "bang"),
			givenValue:    customtypes.NewStringWithAltValuesValue("baz"),
			expectedMatch: true,
		},
		"semantically equal - current matches an alt value": {
			currentValue:  customtypes.NewStringWithAltValuesValue("baz"),
			givenValue:    customtypes.NewStringWithAltValuesValue("foo", "bar", "baz", "bang"),
			expectedMatch: true,
		},
		"not equal": {
			currentValue:  customtypes.NewStringWithAltValuesValue("foo", "bar", "baz", "bang"),
			givenValue:    customtypes.NewStringWithAltValuesValue("FOO"),
			expectedMatch: false,
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			match, diags := tCase.currentValue.StringSemanticEquals(context.Background(), tCase.givenValue)
			require.Equalf(t, tCase.expectedMatch, match, "Expected StringSemanticEquals to return: %t, but got: %t", tCase.expectedMatch, match)
			require.Nil(t, diags)
		})
	}
}
