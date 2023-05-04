package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func TagA(ctx context.Context) (*apstra.DesignTag, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient()
	if err != nil {
		return nil, deleteFunc, err
	}

	tagData := &apstra.DesignTagData{
		Label:       acctest.RandString(10),
		Description: acctest.RandString(10),
	}

	id, err := client.CreateTag(ctx, &apstra.DesignTagRequest{
		Label:       tagData.Label,
		Description: tagData.Description,
	})
	if err != nil {
		return nil, deleteFunc, err
	}

	deleteFunc = func(ctx context.Context) error {
		return client.DeleteTag(ctx, id)
	}

	return &apstra.DesignTag{
		Id:   id,
		Data: tagData,
	}, deleteFunc, nil
}
