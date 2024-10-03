//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	appSchemaJSON1 = `
{
  "properties": {
    "key": {
      "properties": {
        "authenticated_vlan": {
          "type": "string"
        },
        "authorization_status": {
          "type": "string"
        },
        "fallback_vlan_active": {
          "enum": [
            "True",
            "False"
          ],
          "type": "string"
        },
        "port_status": {
          "enum": [
            "authorized",
            "blocked"
          ],
          "type": "string"
        },
        "supplicant_mac": {
          "pattern": "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$",
          "type": "string"
        }
      },
      "required": [
        "authenticated_vlan",
        "authorization_status",
        "fallback_vlan_active",
        "port_status",
        "supplicant_mac"
      ],
      "type": "object"
    },
    "value": {
      "description": "0 in case of blocked, 1 in case of authorized",
      "type": "integer"
    }
  },
  "required": [
    "key",
    "value"
  ],
  "type": "object"
}
`
	appSchemaJSON2 = `
{
  "properties": {
    "key": {
      "properties": {
        "authenticated_vlan": {
          "type": "string"
        },
        "authorization_status": {
          "type": "string"
        },
        "fallback_vlan_active": {
          "enum": [
            "True",
            "False"
          ],
          "type": "string"
        },
        "port_status": {
          "enum": [
            "authorized",
            "blocked"
          ],
          "type": "string"
        }
      },
      "required": [
        "authenticated_vlan",
        "authorization_status",
        "fallback_vlan_active",
        "port_status"
      ],
      "type": "object"
    },
    "value": {
      "description": "0 in case of blocked, 1 in case of authorized",
      "type": "string"
    }
  },
  "required": [
    "key",
    "value"
  ],
  "type": "object"
}
`
	storageSchema1 = "iba_integer_data"
	storageSchema2 = "iba_string_data"

	resourceTelemetryServiceRegistryEntryHCL = `
resource %q %q {
  name                = %q
  description         = %s
  application_schema  = %q
  storage_schema_path = %q
}
`
)

type resourceTelemetryServiceRegistryEntry struct {
	name              string
	description       string
	applicationSchema string
	storageSchemaPath string
}

func (o resourceTelemetryServiceRegistryEntry) render(rType, rName string) string {
	return fmt.Sprintf(resourceTelemetryServiceRegistryEntryHCL,
		rType, rName,
		o.name,
		stringOrNull(o.description),
		o.applicationSchema,
		o.storageSchemaPath,
	)
}

func (o resourceTelemetryServiceRegistryEntry) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceInt64AttrJsonEq", "application_schema", o.applicationSchema)
	result.append(t, "TestCheckResourceAttr", "storage_schema_path", o.storageSchemaPath)
	result.append(t, "TestCheckResourceAttrSet", "version")
	result.append(t, "TestCheckResourceAttr", "built_in", "false")

	if o.description != "" {
		result.append(t, "TestCheckResourceAttr", "description", o.description)
	} else {
		result.append(t, "TestCheckNoResourceAttr", "description")
	}

	return result
}

func TestAccResourceTelemetryServiceRegistryEntry(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))
	log.Printf("Apstra version %s", apiVersion)

	type testStep struct {
		config resourceTelemetryServiceRegistryEntry
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceTelemetryServiceRegistryEntry{
						name:              acctest.RandString(6),
						applicationSchema: appSchemaJSON1,
						storageSchemaPath: storageSchema1,
					},
				},
				{
					config: resourceTelemetryServiceRegistryEntry{
						name:              acctest.RandString(6),
						description:       acctest.RandString(6),
						applicationSchema: appSchemaJSON2,
						storageSchemaPath: storageSchema2,
					},
				},
				{
					config: resourceTelemetryServiceRegistryEntry{
						name:              acctest.RandString(6),
						applicationSchema: appSchemaJSON1,
						storageSchemaPath: storageSchema1,
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceTelemetryServiceRegistryEntry{
						name:              acctest.RandString(6),
						description:       acctest.RandString(6),
						applicationSchema: appSchemaJSON2,
						storageSchemaPath: storageSchema2,
					},
				},
				{
					config: resourceTelemetryServiceRegistryEntry{
						name:              acctest.RandString(6),
						applicationSchema: appSchemaJSON1,
						storageSchemaPath: storageSchema1,
					},
				},
				{
					config: resourceTelemetryServiceRegistryEntry{
						name:              acctest.RandString(6),
						description:       acctest.RandString(6),
						applicationSchema: appSchemaJSON2,
						storageSchemaPath: storageSchema2,
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceTelemetryServiceRegistryEntry)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() // do not use t.Parallel() - these things don't have IDs and will likely collide
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.apiVersionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
