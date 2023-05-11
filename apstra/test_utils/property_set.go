package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func PropertySetA(ctx context.Context) (*apstra.PropertySet, func(context.Context) error, error) {
	client, err := GetTestClient()
	deleteFunc := func(_ context.Context) error { return nil }
	if err != nil {
		return nil, deleteFunc, err
	}

	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	request := apstra.PropertySetData{
		Label:  name,
		Values: []byte(`{"value_int": 42, "value_json": {"inner_value_str": "innerstr", "inner_value_int": 4242}, "value_str": "str"}`),
	}

	id, err := client.CreatePropertySet(ctx, &request)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeletePropertySet(ctx, id)
	}

	ps, err := client.GetPropertySet(ctx, id)
	if err != nil {
		return nil, deleteFunc, err
	}

	return ps, deleteFunc, nil
}
