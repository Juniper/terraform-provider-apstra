package tfapstra_test

import (
	"sync"
	"testing"

	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
)

// TestMain exists to initialize the MakeOrFindBlueprintMutex which makes the
// MakeOrFindBlueprint function concurrency safe.
func TestMain(m *testing.M) {
	if testutils.MakeOrFindBlueprintMutex == nil {
		testutils.MakeOrFindBlueprintMutex = new(sync.Mutex)
	}

	m.Run()
}
