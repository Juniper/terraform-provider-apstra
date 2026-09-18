//go:build integration

package random

import (
	"math/rand"
	"sync"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

var (
	persistenceMutex  = new(sync.Mutex)
	persistentInts    = make(map[string]int)
	persistentStrings = make(map[string]string)
)

// PersistentIntn generates and returns a random int using rand.Intn(). The same int will be
// returned with each invocation of PersistentIntn using a given key. Note that the n element is
// only consulted when generating an int. Keys found in cache will be returned regardless of n.
func PersistentIntn(key string, n int) int {
	persistenceMutex.Lock()
	defer persistenceMutex.Unlock()

	if v, ok := persistentInts[key]; ok {
		return v
	}

	v := rand.Intn(n)
	persistentInts[key] = v
	return v
}

// PersistentString generates and returns a random string of the specified length using the
// (optional) character set string. The same string will be returned with each invocation of
// PersistentString using a given key. Note that the length and chars elements are only consulted
// when generating a string. Keys found in cache will be returned regardless of length and chars.
func PersistentString(key string, length int, chars ...string) string {
	persistenceMutex.Lock()
	defer persistenceMutex.Unlock()

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
