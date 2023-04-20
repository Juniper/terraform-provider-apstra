package tfapstra

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
)

func newKeyLogWriter(fileName string) (*os.File, error) {
	absPath, err := homedir.Expand(fileName)
	if err != nil {
		return nil, fmt.Errorf("error expanding home directory '%s' - %w", fileName, err)
	}

	err = os.MkdirAll(filepath.Dir(absPath), os.FileMode(0600))
	if err != nil {
		return nil, err
	}
	return os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
}

// sliceWithoutElement returns a copy of in with all occurrences of e removed.
// the returned int indicates the number of occurrences removed.
func sliceWithoutElement[A comparable](in []A, e A) ([]A, int) {
	result := make([]A, len(in))
	var resultIdx int
	for inIdx := range in {
		if in[inIdx] == e {
			continue
		}
		result[resultIdx] = in[inIdx]
		resultIdx++
	}
	return result[:resultIdx], len(in) - resultIdx
}
