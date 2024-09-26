package tfapstra

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceBlueprintIbaProbeHCL = `
resource "apstra_blueprint_iba_probe" "p_device_health" {
  blueprint_id = "%s"
  predefined_probe_id = "device_health"
  probe_config = jsonencode(
    {
      "max_cpu_utilization": 80,
      "max_memory_utilization": 80,
      "max_disk_utilization": 80,
      "duration": 660,
      "threshold_duration": 360,
      "history_duration": 604800
    }
  )
}
`
	probeStrErr = `{
    bad_json = bad
	1 : "more badness"
}`
	probeStr = `{
  "label": "Device Traffic",
  "processors": [
    {
      "name": "Live Interface Counters",
      "type": "traffic_monitor",
      "properties": {
        "speed": "functions.speed_to_bits(link.speed)",
        "graph_query": "node(\"system\", name=\"system\", system_id=not_none(),role=is_in([\"spine\", \"leaf\", \"superspine\", \"access\"])).out(\"hosted_interfaces\").node(\"interface\", name=\"iface\", if_name=not_none()).out(\"link\").node(\"link\", name=\"link\", link_type=\"ethernet\")",
        "query_group_by": [],
        "period": 120,
        "query_tag_filter": {
          "filter": {},
          "operation": "and"
        },
        "interface": "iface.if_name",
        "system_id": "system.system_id",
        "query_expansion": {},
        "enable_streaming": false
      },
      "inputs": {},
      "outputs": {
        "interface_counters_average": "Average Interface Counters",
        "out": "Live Interface Counters"
      }
    },
    {
      "name": "System Interface Counters",
      "type": "system_utilization",
      "properties": {
        "enable_streaming": false
      },
      "inputs": {
        "tx_utilization": {
          "stage": "Average Interface Counters",
          "column": "tx_utilization_average"
        },
        "tx_bps": {
          "stage": "Average Interface Counters",
          "column": "tx_bps_average"
        },
        "rx_utilization": {
          "stage": "Average Interface Counters",
          "column": "rx_utilization_average"
        },
        "rx_bps": {
          "stage": "Average Interface Counters",
          "column": "rx_bps_average"
        }
      },
      "outputs": {
        "out": "System Interface Counters"
      }
    }
  ],
  "stages": [
    {
      "name": "System Interface Counters",
      "enable_metric_logging": true,
      "retention_duration": 2592000,
      "description": "Interface data grouped per system",
      "units": {
        "aggregate_rx_utilization": "%",
        "aggregate_tx_bps": "bps",
        "aggregate_tx_utilization": "%",
        "aggregate_rx_bps": "bps",
        "max_ifc_rx_utilization": "%",
        "max_ifc_tx_utilization": "%"
      }
    },
    {
      "name": "Average Interface Counters",
      "enable_metric_logging": true,
      "retention_duration": 2592000,
      "description": "Average interface counter data",
      "units": {
        "tx_bps_average": "bps",
        "symbol_errors_per_second_average": "",
        "alignment_errors_per_second_average": "",
        "fcs_errors_per_second_average": "",
        "tx_multicast_pps_average": "pps",
        "rx_utilization_average": "%",
        "tx_error_pps_average": "pps",
        "rx_discard_pps_average": "pps",
        "rx_error_pps_average": "pps",
        "tx_discard_pps_average": "pps",
        "runts_per_second_average": "",
        "giants_per_second_average": "",
        "rx_bps_average": "bps",
        "tx_unicast_pps_average": "pps",
        "rx_broadcast_pps_average": "pps",
        "tx_broadcast_pps_average": "pps",
        "rx_unicast_pps_average": "pps",
        "rx_multicast_pps_average": "pps",
        "tx_utilization_average": "%"
      }
    },
    {
      "name": "Live Interface Counters",
      "retention_duration": 86400,
      "description": "Live interface counter data",
      "units": {
        "tx_unicast_pps": "pps",
        "runts_per_second": "",
        "rx_multicast_pps": "pps",
        "symbol_errors_per_second": "",
        "rx_broadcast_pps": "pps",
        "alignment_errors_per_second": "",
        "fcs_errors_per_second": "",
        "tx_utilization": "%",
        "tx_discard_pps": "pps",
        "tx_bps": "bps",
        "tx_multicast_pps": "pps",
        "rx_utilization": "%",
        "rx_error_pps": "pps",
        "rx_bps": "bps",
        "tx_error_pps": "pps",
        "tx_broadcast_pps": "pps",
        "rx_unicast_pps": "pps",
        "rx_discard_pps": "pps",
        "giants_per_second": ""
      }
    }
  ]
}`
	resourceBlueprintIbaProbeJsonHCL = `
	resource "apstra_blueprint_iba_probe" "p_device_traffic" {
		blueprint_id = "%s"
		probe_json =  <<-EOT
		%s
		EOT
	}`
)

func TestAccResourceProbe(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)
	clientVersion := version.Must(version.NewVersion(client.ApiVersion()))
	if !compatibility.BpIbaProbeOk.Check(clientVersion) {
		t.Skipf("skipping due to version constraint %s", compatibility.BpIbaProbeOk)
	}

	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "BPA", testutils.BlueprintA)

	type testCase struct {
		steps []resource.TestStep
	}

	testCases := map[string]testCase{
		"predefined": {
			steps: []resource.TestStep{
				{
					Config: insecureProviderConfigHCL + fmt.Sprintf(resourceBlueprintIbaProbeHCL, bpClient.Id()),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_blueprint_iba_probe.p_device_health", "id"),
						resource.TestCheckResourceAttr("apstra_blueprint_iba_probe.p_device_health", "name", "Device System Health"),
					),
				},
			},
		},
		"from_json": {
			steps: []resource.TestStep{
				{
					Config: insecureProviderConfigHCL + fmt.Sprintf(resourceBlueprintIbaProbeJsonHCL, bpClient.Id(), probeStr),
					Check: resource.ComposeAggregateTestCheckFunc(
						// Verify ID has any value set
						resource.TestCheckResourceAttrSet("apstra_blueprint_iba_probe.p_device_traffic", "id"),
						resource.TestCheckResourceAttr("apstra_blueprint_iba_probe.p_device_traffic", "name", "Device Traffic"),
					),
				},
			},
		},
		"error": {
			steps: []resource.TestStep{
				{
					Config:      insecureProviderConfigHCL + fmt.Sprintf(resourceBlueprintIbaProbeJsonHCL, bpClient.Id(), probeStrErr),
					ExpectError: regexp.MustCompile("Error: Invalid JSON String Value.*"),
					Destroy:     false,
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    tCase.steps,
			})
		})
	}
}
