package random_test

import (
	"testing"

	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/stretchr/testify/require"
)

func TestRouteTarget(t *testing.T) {
	for range 100 { // Run several times to increase confidence the result is always valid.
		got := random.RouteTarget(t)

		t.Run(got, func(t *testing.T) {
			t.Parallel()

			parsed, err := datacenter.NewRouteTarget(got) // use the parser from apstra-go-sdk/datacenter
			require.NoError(t, err)

			bytes, err := parsed.MarshalText()
			require.NoError(t, err)

			require.Equal(t, got, string(bytes))
		})
	}
}
