package customtypes_test

import (
	"context"
	"testing"

	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestStringWithEmptyEqualsNull_StringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentValue  customtypes.StringWithEmptyEqualsNull
		givenValue    basetypes.StringValuable
		expectedMatch bool
	}{
		"equal": {
			currentValue:  customtypes.NewStringWithEmptyEqualsNullValue("foo"),
			givenValue:    customtypes.NewStringWithEmptyEqualsNullValue("foo"),
			expectedMatch: true,
		},
		"semantically equal - null and empty": {
			currentValue:  customtypes.NewStringWithEmptyEqualsNullNull(),
			givenValue:    customtypes.NewStringWithEmptyEqualsNullValue(""),
			expectedMatch: true,
		},
		"semantically equal - empty and null": {
			currentValue:  customtypes.NewStringWithEmptyEqualsNullValue(""),
			givenValue:    customtypes.NewStringWithEmptyEqualsNullNull(),
			expectedMatch: true,
		},
		"not equal": {
			currentValue:  customtypes.NewStringWithEmptyEqualsNullValue("foo"),
			givenValue:    customtypes.NewStringWithEmptyEqualsNullValue("FOO"),
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
