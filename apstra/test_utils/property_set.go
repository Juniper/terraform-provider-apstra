package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"testing"
)

func PropertySetA(t testing.TB, ctx context.Context) *apstra.PropertySet {
	t.Helper()

	client := GetTestClient(t, ctx)

	request := apstra.PropertySetData{
		Label:  acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Values: []byte(`{"value_int": 42, "value_json": {"inner_value_str": "innerstr", "inner_value_int": 4242}, "value_str": "str"}`),
	}

	id, err := client.CreatePropertySet(ctx, &request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeletePropertySet(ctx, id)) })

	ps, err := client.GetPropertySet(ctx, id)
	require.NoError(t, err)

	return ps
}
