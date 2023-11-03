package tfapstra

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"golang.org/x/exp/constraints"
	"math/rand"
	"strings"
	"unsafe"
)

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

func FillWithRandomIntegers[A constraints.Integer](a []A) {
	if len(a) == 0 {
		return
	}

	var nominal A
	maxSize := unsafe.Sizeof(uint64(0))
	nominalSize := unsafe.Sizeof(nominal)
	if nominalSize > maxSize {
		panic(fmt.Sprintf("FillWithRandomIntegers got unexpectedly large integer type %v", nominal))
	}

	for i := 0; i < len(a); i++ {
		a[i] = A(rand.Uint64())
	}
}

// SplitImportId splits string 'in' into len(fieldNames) strings using the first character as
// a separator for the fields. fieldNames is used mainly for its length. The values in fieldNames
// are ignored unless there's a need to print an error message.
func SplitImportId(_ context.Context, in string, fieldNames []string, diags *diag.Diagnostics) []string {
	if len(in) < 2 {
		diags.AddError("invalid import ID", "import ID minimum length is 2")
		return nil
	}

	sep := in[:1]
	parts := strings.Split(in[:], sep)[1:]

	if len(fieldNames) != len(parts) {
		form := "<separator><" + strings.Join(fieldNames, "><separator><") + ">"
		diags.AddError(
			fmt.Sprintf("cannot parse import ID: %q", in),
			fmt.Sprintf("ID string for resource import must take this form:\n\n"+
				"  %s\n\n"+
				"where <separator> is any single character not found in any of the delimited fields. "+
				"Expected %d parts after splitting on '%s', got %d parts", form, len(fieldNames), sep, len(parts)))
		return nil
	}

	return parts
}
