//go:build integration

package random_test

import (
	"strings"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func TestPersistentStringGeneration(t *testing.T) {

	hasDisallowedChars := func(s, allowed string) bool {
		return strings.IndexFunc(s, func(r rune) bool {
			return !strings.ContainsRune(allowed, r)
		}) >= 0
	}

	testCases := map[string]struct {
		key    string
		length int
		chars  string
		want   string
	}{
		"len_only": {
			key:    "len_only",
			length: 12,
		},
		"set_with_multiple_chars": {
			key:    "set_with_multiple_chars",
			length: 10,
			chars:  acctest.CharSetAlphaNum,
		},
		"set_with_single_char": {
			key:    "set_with_single_char",
			length: 5,
			chars:  "x",
			want:   "xxxxx",
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			var got, got2 string
			if len(tCase.chars) > 0 {
				got = random.PersistentString(tCase.key, tCase.length, tCase.chars)
				got2 = random.PersistentString(tCase.key, tCase.length, tCase.chars)
			} else {
				got = random.PersistentString(tCase.key, tCase.length)
				got2 = random.PersistentString(tCase.key, tCase.length)
			}

			require.Equal(t, got, got2)
			require.Equal(t, tCase.length, len(got))
			if tCase.want != "" {
				require.Equal(t, tCase.want, got)
			}

			if tCase.chars != "" {
				require.False(t, hasDisallowedChars(got, tCase.chars))
			}
		})
	}
}
