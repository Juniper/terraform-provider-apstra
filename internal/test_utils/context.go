//go:build integration

package testutils

import (
	"context"
	"testing"
	"time"
)

func CleanupWithFreshContext(t testing.TB, timeout time.Duration, f func(ctx context.Context) error) {
	t.Helper()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err := f(ctx)
		if err != nil {
			t.Logf("Cleanup test %q: %v", t.Name(), err)
		}
	})
}
