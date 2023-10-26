---
page_title: "apstra_anomalies Data Source - terraform-provider-apstra"
subcategory: "Reference Design: Shared"
description: |-
  This data source provides per-node summary, per-service summary and full details of anomalies in the specified Blueprint.
---

# apstra_anomalies (Data Source)

This data source provides per-node summary, per-service summary and full details of anomalies in the specified Blueprint.


## Example Usage

```terraform
# This example pulls all of the anomaly details from a blueprint with several
# cabling mistakes which produce a total of 21 anomalies.
data "apstra_anomalies" "anomalies" {
  blueprint_id = "c8fe07d6-f631-46e6-ae2a-9c6f3929e6a3"
}

output anomalies { value = data.apstra_anomalies.anomalies }

# The output looks like this:

#anomalies = {
#  "blueprint_id" = "c8fe07d6-f631-46e6-ae2a-9c6f3929e6a3"
#  "details" = toset([
#    {
#      "actual" = "{\"interfaces_up\":[],\"intf_up_count\":0}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "40403043-dba1-4546-9a7e-00d92858d590"
#      "expected" = "{\"interfaces_up\":[\"ge-0/0/0\"],\"intf_up_count\":1}"
#      "identity" = "{\"anomaly_type\":\"lag\",\"lag\":\"ae1\",\"system_id\":\"525400240557\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "lag"
#    },
#    {
#      "actual" = "{\"interfaces_up\":[],\"intf_up_count\":0}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "4bdd2de3-2022-4151-82f2-e734e57a5619"
#      "expected" = "{\"interfaces_up\":[\"ge-0/0/1\"],\"intf_up_count\":1}"
#      "identity" = "{\"anomaly_type\":\"lag\",\"lag\":\"ae1\",\"system_id\":\"525400E19D40\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "lag"
#    },
#    {
#      "actual" = "{\"neighbor_interface\":\"ge-0/0/0\",\"neighbor_name\":\"test-001-access1\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "76692c7a-2067-4372-8906-02a015825557"
#      "expected" = "{\"neighbor_interface\":\"ge-0/0/0\",\"neighbor_name\":\"test-001-leaf2\"}"
#      "identity" = "{\"anomaly_type\":\"cabling\",\"interface\":\"ge-0/0/1\",\"system_id\":\"5254008A8F34\"}"
#      "role" = "spine_leaf"
#      "severity" = "critical"
#      "type" = "cabling"
#    },
#    {
#      "actual" = "{\"neighbor_interface\":\"ge-0/0/0\",\"neighbor_name\":\"test-001-access2\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "cde14dba-daf0-4bf3-9b49-a718cfc3694d"
#      "expected" = "{\"neighbor_interface\":\"ge-0/0/0\",\"neighbor_name\":\"test-001-access1\"}"
#      "identity" = "{\"anomaly_type\":\"cabling\",\"interface\":\"ge-0/0/1\",\"system_id\":\"525400523E0A\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "cabling"
#    },
#    {
#      "actual" = "{\"neighbor_interface\":\"ge-0/0/1\",\"neighbor_name\":\"spine1\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "e9e29a10-386d-4d87-bf5c-91c0db7eb878"
#      "expected" = "{\"neighbor_interface\":\"ge-0/0/1\",\"neighbor_name\":\"test-001-leaf1\"}"
#      "identity" = "{\"anomaly_type\":\"cabling\",\"interface\":\"ge-0/0/0\",\"system_id\":\"525400240557\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "cabling"
#    },
#    {
#      "actual" = "{\"neighbor_interface\":\"ge-0/0/1\",\"neighbor_name\":\"test-001-leaf1\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "294822e8-7d8f-4978-9ccb-749ea373c6e8"
#      "expected" = "{\"neighbor_interface\":\"ge-0/0/1\",\"neighbor_name\":\"test-001-leaf2\"}"
#      "identity" = "{\"anomaly_type\":\"cabling\",\"interface\":\"ge-0/0/0\",\"system_id\":\"525400101F55\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "cabling"
#    },
#    {
#      "actual" = "{\"neighbor_interface\":\"ge-0/0/2\",\"neighbor_name\":\"spine1\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "9e7b289e-f106-4856-add7-e4566ee57314"
#      "expected" = "{\"neighbor_interface\":\"ge-0/0/1\",\"neighbor_name\":\"spine1\"}"
#      "identity" = "{\"anomaly_type\":\"cabling\",\"interface\":\"ge-0/0/0\",\"system_id\":\"525400E19D40\"}"
#      "role" = "spine_leaf"
#      "severity" = "critical"
#      "type" = "cabling"
#    },
#    {
#      "actual" = "{\"neighbor_interface\":\"ge-0/0/2\",\"neighbor_name\":\"test-001-access2\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "31d7aa87-efb3-4695-9312-0580d9c4ea27"
#      "expected" = "{\"neighbor_interface\":\"ge-0/0/0\",\"neighbor_name\":\"test-001-access2\"}"
#      "identity" = "{\"anomaly_type\":\"cabling\",\"interface\":\"ge-0/0/1\",\"system_id\":\"525400E19D40\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "cabling"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "0cf491f3-031a-4fa2-8daa-61aa23bbe045"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"addr_family\":\"evpn\",\"anomaly_type\":\"bgp\",\"destination_asn\":\"4200000002\",\"destination_ip\":\"10.60.60.2\",\"destination_name\":\"test-001-leaf2\",\"source_asn\":\"4200000000\",\"source_ip\":\"10.60.60.0\",\"system_id\":\"5254008A8F34\",\"vrf_name\":\"default\"}"
#      "role" = "spine_leaf"
#      "severity" = "critical"
#      "type" = "bgp"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "1fd5a3c3-07cf-4b97-81f2-b6e1153b48ee"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"interface\",\"interface\":\"ae1.0\",\"system_id\":\"525400240557\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "interface"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "2058be80-5be3-4787-b09b-a46935c52974"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"addr_family\":\"ipv4\",\"anomaly_type\":\"bgp\",\"destination_asn\":\"4200000000\",\"destination_ip\":\"10.60.60.8\",\"destination_name\":\"spine1\",\"source_asn\":\"4200000002\",\"source_ip\":\"10.60.60.9\",\"system_id\":\"525400E19D40\",\"vrf_name\":\"default\"}"
#      "role" = "spine_leaf"
#      "severity" = "critical"
#      "type" = "bgp"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "3dca6d8b-23ab-4636-b446-48b3979967f4"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"interface\",\"interface\":\"ae1.0\",\"system_id\":\"525400E19D40\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "interface"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "3fdd09db-832d-4186-8e8a-81ec69cf9837"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"interface\",\"interface\":\"ae1\",\"system_id\":\"525400E19D40\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "interface"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "8a8b0438-b751-46c4-a3ea-992a58fe1d0d"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"addr_family\":\"ipv4\",\"anomaly_type\":\"bgp\",\"destination_asn\":\"4200000002\",\"destination_ip\":\"10.60.60.9\",\"destination_name\":\"test-001-leaf2\",\"source_asn\":\"4200000000\",\"source_ip\":\"10.60.60.8\",\"system_id\":\"5254008A8F34\",\"vrf_name\":\"default\"}"
#      "role" = "spine_leaf"
#      "severity" = "critical"
#      "type" = "bgp"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "bba106ea-ed79-4268-81f4-dad5ce3f45b9"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"addr_family\":\"evpn\",\"anomaly_type\":\"bgp\",\"destination_asn\":\"4200000000\",\"destination_ip\":\"10.60.60.0\",\"destination_name\":\"spine1\",\"source_asn\":\"4200000002\",\"source_ip\":\"10.60.60.2\",\"system_id\":\"525400E19D40\",\"vrf_name\":\"default\"}"
#      "role" = "spine_leaf"
#      "severity" = "critical"
#      "type" = "bgp"
#    },
#    {
#      "actual" = "{\"value\":\"down\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "e73513eb-9393-4893-8306-7e0f23e4a4a2"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"interface\",\"interface\":\"ae1\",\"system_id\":\"525400240557\"}"
#      "role" = "leaf_access"
#      "severity" = "critical"
#      "type" = "interface"
#    },
#    {
#      "actual" = "{\"value\":\"missing\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "7a800b7c-605c-4efa-8d7a-852b651aa3c3"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"route\",\"destination_ip\":\"10.60.60.0/32\",\"system_id\":\"525400E19D40\"}"
#      "role" = "unknown"
#      "severity" = "critical"
#      "type" = "route"
#    },
#    {
#      "actual" = "{\"value\":\"missing\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "8fc0ad40-a4d3-4932-99c9-bca252666373"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"route\",\"destination_ip\":\"10.60.60.1/32\",\"system_id\":\"525400E19D40\"}"
#      "role" = "unknown"
#      "severity" = "critical"
#      "type" = "route"
#    },
#    {
#      "actual" = "{\"value\":\"missing\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "978c04e4-c029-4d28-8502-924d0b78a920"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"route\",\"destination_ip\":\"10.60.60.2/32\",\"system_id\":\"5254008A8F34\"}"
#      "role" = "unknown"
#      "severity" = "critical"
#      "type" = "route"
#    },
#    {
#      "actual" = "{\"value\":\"missing\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "a7141677-b4a4-470b-8027-6c84fef7fd51"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"route\",\"destination_ip\":\"10.60.60.6/31\",\"system_id\":\"525400E19D40\"}"
#      "role" = "unknown"
#      "severity" = "critical"
#      "type" = "route"
#    },
#    {
#      "actual" = "{\"value\":\"missing\"}"
#      "anomalous" = tostring(null)
#      "anomaly_id" = "b0136228-a626-4ed2-8df2-552ab32cb7d0"
#      "expected" = "{\"value\":\"up\"}"
#      "identity" = "{\"anomaly_type\":\"route\",\"destination_ip\":\"10.60.60.2/32\",\"system_id\":\"525400523E0A\"}"
#      "role" = "unknown"
#      "severity" = "critical"
#      "type" = "route"
#    },
#  ])
#  "summary_by_node" = toset([
#    {
#      "arp" = 0
#      "bgp" = 0
#      "blueprint_rendering" = 0
#      "cabling" = 1
#      "config" = 0
#      "counter" = 0
#      "deployment" = 0
#      "hostname" = 0
#      "interface" = 0
#      "lag" = 0
#      "liveness" = 0
#      "mac" = 0
#      "mlag" = 0
#      "node_name" = "test_001_access2"
#      "probe" = 0
#      "route" = 0
#      "series" = 0
#      "streaming" = 0
#      "system_id" = "525400101F55"
#      "total" = 1
#    },
#    {
#      "arp" = 0
#      "bgp" = 0
#      "blueprint_rendering" = 0
#      "cabling" = 1
#      "config" = 0
#      "counter" = 0
#      "deployment" = 0
#      "hostname" = 0
#      "interface" = 0
#      "lag" = 0
#      "liveness" = 0
#      "mac" = 0
#      "mlag" = 0
#      "node_name" = "test_001_leaf1"
#      "probe" = 0
#      "route" = 1
#      "series" = 0
#      "streaming" = 0
#      "system_id" = "525400523E0A"
#      "total" = 2
#    },
#    {
#      "arp" = 0
#      "bgp" = 0
#      "blueprint_rendering" = 0
#      "cabling" = 1
#      "config" = 0
#      "counter" = 0
#      "deployment" = 0
#      "hostname" = 0
#      "interface" = 2
#      "lag" = 1
#      "liveness" = 0
#      "mac" = 0
#      "mlag" = 0
#      "node_name" = "test_001_access1"
#      "probe" = 0
#      "route" = 0
#      "series" = 0
#      "streaming" = 0
#      "system_id" = "525400240557"
#      "total" = 4
#    },
#    {
#      "arp" = 0
#      "bgp" = 2
#      "blueprint_rendering" = 0
#      "cabling" = 1
#      "config" = 0
#      "counter" = 0
#      "deployment" = 0
#      "hostname" = 0
#      "interface" = 0
#      "lag" = 0
#      "liveness" = 0
#      "mac" = 0
#      "mlag" = 0
#      "node_name" = "spine1"
#      "probe" = 0
#      "route" = 1
#      "series" = 0
#      "streaming" = 0
#      "system_id" = "5254008A8F34"
#      "total" = 4
#    },
#    {
#      "arp" = 0
#      "bgp" = 2
#      "blueprint_rendering" = 0
#      "cabling" = 2
#      "config" = 0
#      "counter" = 0
#      "deployment" = 0
#      "hostname" = 0
#      "interface" = 2
#      "lag" = 1
#      "liveness" = 0
#      "mac" = 0
#      "mlag" = 0
#      "node_name" = "test_001_leaf2"
#      "probe" = 0
#      "route" = 3
#      "series" = 0
#      "streaming" = 0
#      "system_id" = "525400E19D40"
#      "total" = 10
#    },
#  ])
#  "summary_by_service" = toset([
#    {
#      "count" = 21
#      "role" = "all"
#      "type" = "all"
#    },
#    {
#      "count" = 2
#      "role" = "leaf_access"
#      "type" = "lag"
#    },
#    {
#      "count" = 2
#      "role" = "spine_leaf"
#      "type" = "cabling"
#    },
#    {
#      "count" = 4
#      "role" = "leaf_access"
#      "type" = "cabling"
#    },
#    {
#      "count" = 4
#      "role" = "leaf_access"
#      "type" = "interface"
#    },
#    {
#      "count" = 4
#      "role" = "spine_leaf"
#      "type" = "bgp"
#    },
#    {
#      "count" = 5
#      "role" = "unknown"
#      "type" = "route"
#    },
#  ])
#}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID.

### Read-Only

- `details` (Attributes Set) Each current Anomaly is represented by an object in this set. (see [below for nested schema](#nestedatt--details))
- `summary_by_node` (Attributes Set) Set of Anomaly summaries organized by Node. (see [below for nested schema](#nestedatt--summary_by_node))
- `summary_by_service` (Attributes Set) Set of Anomaly summaries organized by Fabric Service. (see [below for nested schema](#nestedatt--summary_by_service))

<a id="nestedatt--details"></a>
### Nested Schema for `details`

Read-Only:

- `actual` (String) Extended Anomaly attribute describing the actual value/state/condition in JSON format.
- `anomalous` (String) Extended Anomaly attribute which further contextualizes the Anomaly.
- `anomaly_id` (String) Apstra Anomaly ID.
- `expected` (String) Extended Anomaly attribute describing the expected value/state/condition in JSON format.
- `identity` (String) Extended Anomaly attribute which identifies the anomalous value/state/condition in JSON format.
- `role` (String) Anomaly role further contextualizes `type`.
- `severity` (String) Severity of Anomaly.
- `type` (String) Anomaly Type.


<a id="nestedatt--summary_by_node"></a>
### Nested Schema for `summary_by_node`

Read-Only:

- `arp` (Number) Number of ARP Anomalies related to the Node.
- `bgp` (Number) Number of BGP Anomalies related to the Node.
- `blueprint_rendering` (Number) Number of Blueprint Rendering Anomalies related to the Node.
- `cabling` (Number) Number of Cabling Anomalies related to the Node.
- `config` (Number) Number of Config Anomalies related to the Node.
- `counter` (Number) Number of Counter Anomalies related to the Node.
- `deployment` (Number) Number of Deployment Anomalies related to the Node.
- `hostname` (Number) Number of Hostname Anomalies related to the Node.
- `interface` (Number) Number of Interface Anomalies related to the Node.
- `lag` (Number) Number of LAG Anomalies related to the Node.
- `liveness` (Number) Number of Liveness Anomalies related to the Node.
- `mac` (Number) Number of MAC Anomalies related to the Node.
- `mlag` (Number) Number of MLAG Anomalies related to the Node.
- `node_name` (String) Name of the Node experiencing Anomalies.
- `probe` (Number) Number of Probe Anomalies related to the Node.
- `route` (Number) Number of Route Anomalies related to the Node.
- `series` (Number) Number of Series Anomalies related to the Node.
- `streaming` (Number) Number of Streaming Anomalies related to the Node.
- `system_id` (String) System ID of the Node experiencing Anomalies.
- `total` (Number) Total number of Anomalies related to the Node.


<a id="nestedatt--summary_by_service"></a>
### Nested Schema for `summary_by_service`

Read-Only:

- `count` (Number) Count of Anomalies related to the Fabric Service and Role.
- `role` (String) Further context about the Fabric Service Anomalies.
- `type` (String) Fabric Service experiencing Anomalies.
