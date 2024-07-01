package utils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestGetCtIpLinkSubinterfaces(t *testing.T) {
	ctx := context.Background()

	clientCfg := apstra.ClientCfg{
		Url:  "https://apstra-973cfb91-ecfb-4b46-8715-ce1849d7a041.aws.apstra.com",
		User: "admin",
		Pass: "CoolDuck3%",
	}

	client, err := clientCfg.NewClient(ctx)
	require.NoError(t, err)

	bp, err := client.NewTwoStageL3ClosClient(ctx, "e38a9808-6bbf-470d-82f1-f19e33d76c98")
	require.NoError(t, err)

	var diags diag.Diagnostics

	ctId := apstra.ObjectId("0bdf84ee-8129-4da6-8b35-c93c4febccce")
	apId := apstra.ObjectId("k6L8SJevXl663_qqUlY")

	vlanToSubinterfaces := GetCtIpLinkSubinterfaces(ctx, bp, ctId, apId, &diags)
	require.False(t, diags.HasError())

	log.Println(vlanToSubinterfaces)
}
