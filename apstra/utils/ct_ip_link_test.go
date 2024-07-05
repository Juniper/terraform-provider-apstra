package utils

import (
	"context"
	"encoding/json"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestThing(t *testing.T) {
	ctx := context.Background()

	clientCfg := apstra.ClientCfg{
		Url:  "https://apstra-2bd238e9-290d-4f8d-9f3a-05b501becd14.aws.apstra.com",
		User: "admin",
		Pass: "WillingMockingbird8-",
	}

	client, err := clientCfg.NewClient(ctx)
	require.NoError(t, err)

	bp, err := client.NewTwoStageL3ClosClient(ctx, "22044be2-e7af-462d-847a-ce6d0b49000e")
	require.NoError(t, err)

	var diags diag.Diagnostics

	ctId := "5765435a-acdd-46e7-81dc-a884dc4478ab"
	//ctId := "d7d2daa1-4b15-4db6-8982-d4ff64a6a723"
	apIds := []string{"IZiK1TY0amaN8UfU9Gc", "UVJpVynfmBMy74Nks4Y"}

	vlanToSubinterfaces := GetCtIpLinkIdsByCtAndAps(ctx, bp, ctId, apIds, &diags)
	require.False(t, diags.HasError())

	m, err := json.MarshalIndent(vlanToSubinterfaces, "", "  ")
	require.NoError(t, err)

	log.Println(string(m))
}
