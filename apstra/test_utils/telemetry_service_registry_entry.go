package testutils

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func TelemetryServiceRegistryEntryA(t testing.TB, ctx context.Context) *apstra.TelemetryServiceRegistryEntry {
	t.Helper()

	client := GetTestClient(t, ctx)
	schema := []byte(`{
        "required": ["key","value"],
        "type": "object",
        "properties": {
          	"value": {
            "type": "integer",
            "description": "0 in case of blocked, 1 in case of authorized"
          	},
			"key": {
            "required": [
              "supplicant_mac",
              "authenticated_vlan",
              "authorization_status",
              "port_status",
              "fallback_vlan_active"
            ],
            "type": "object",
            "properties": {
              "supplicant_mac": {
                "type": "string",
                "pattern": "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
              },
              "authenticated_vlan": {
                "type": "string"
              },
              "authorization_status": {
                "type": "string"
              },
              "port_status": {
                "enum": [
                  "authorized",
                  "blocked"
                ],
                "type": "string"
              },
              "fallback_vlan_active": {
                "enum": [
                  "True",
                  "False"
                ],
                "type": "string"
              }
            }
          }
		}
	}`)
	request := apstra.TelemetryServiceRegistryEntry{
		ServiceName:       acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		ApplicationSchema: schema,
		StorageSchemaPath: enum.StorageSchemaPathIbaIntegerData,
		Builtin:           false,
		Description:       "Test Telemetry Service A",
		Version:           "",
	}
	sn, err := client.CreateTelemetryServiceRegistryEntry(ctx, &request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTelemetryServiceRegistryEntry(ctx, sn)) })

	ts, err := client.GetTelemetryServiceRegistryEntry(ctx, sn)
	require.NoError(t, err)

	return ts
}
