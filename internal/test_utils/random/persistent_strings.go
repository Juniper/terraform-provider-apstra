//go:build integration

package random

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

var (
	persistentStrings      = make(map[string]string)
	persistentStringsMutex = new(sync.Mutex)
)

// PersistentString generates and returns a random string of the specified length using the
// (optional) character set string. The same string will be returned with each invocation of
// PersistentString using a given key. Note that the length and chars elements are only consulted
// when generating a string. Keys found in cache will be returned regardless of length and chars.
func PersistentString(key string, length int, chars ...string) string {
	persistentStringsMutex.Lock()
	defer persistentStringsMutex.Unlock()

	if s, ok := persistentStrings[key]; ok {
		return s
	}

	var s string
	if len(chars) > 0 {
		s = acctest.RandStringFromCharSet(length, chars[0])
	} else {
		s = acctest.RandString(length)
	}

	persistentStrings[key] = s
	return s
}
