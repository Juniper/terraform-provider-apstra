package apstraregexp_test

import (
	"regexp"
	"testing"

	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/stretchr/testify/require"
)

func Test_MatchString(t *testing.T) {
	type TestCase struct {
		r *regexp.Regexp // input regex
		s string         // input string
		e bool           // expected result
	}

	testCases := map[string]TestCase{
		"NoLeadingOrTrailingWhitespaceg_with_empty_string": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: "",
			e: true,
		},
		"NoLeadingOrTrailingWhitespaceg_with_single_char": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: "a",
			e: true,
		},
		"NoLeadingOrTrailingWhitespaceg_with_multiple": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: "abc",
			e: true,
		},
		"NoLeadingOrTrailingWhitespaceg_with_emedded_whitespace": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: "abc_d_ef_g",
			e: true,
		},
		"NoLeadingOrTrailingWhitespaceg_with_leading_whitespace": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: " abc",
			e: false,
		},
		"NoLeadingOrTrailingWhitespaceg_with_trailing_whitespace": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: "abc ",
			e: false,
		},
		"NoLeadingOrTrailingWhitespaceg_with_leading_and_trailing_whitespace": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: " abc ",
			e: false,
		},
		"NoLeadingOrTrailingWhitespaceg_with_only_single_whitespace": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: " ",
			e: false,
		},
		"NoLeadingOrTrailingWhitespaceg_with_only_multiple_whitespace": {
			r: apstraregexp.NoLeadingOrTrailingWhitespace,
			s: "  ",
			e: false,
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tCase.e, tCase.r.MatchString(tCase.s))
		})
	}
}
