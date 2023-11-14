---
page_title: "apstra_datacenter_external_gateway Data Source - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This resource returns details of a DCI External Gateway within a Datacenter Blueprint.
  At least one optional attribute is required.
---

# apstra_datacenter_external_gateway (Data Source)

This resource returns details of a DCI External Gateway within a Datacenter Blueprint.

At least one optional attribute is required.


## Example Usage

```terraform
# This example pulls details of the external gateway named "DC2A"
# from blueprint "007723b7-a387-4bb3-8a5e-b5e9f265de0d"

data "apstra_datacenter_external_gateway" "example" {
  blueprint_id = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  name         = "DC2A"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra ID of the Blueprint in which the External Gateway should be created.

### Optional

- `id` (String) Apstra Object ID.
- `name` (String) External Gateway name

### Read-Only

- `asn` (Number) External Gateway AS Number
- `evpn_route_types` (String) EVPN route types. Valid values are: ["all", "type5_only"]. Default: "all"
- `hold_time` (Number) BGP hold time (seconds).
- `ip_address` (String) External Gateway IP address
- `keepalive_time` (Number) BGP keepalive time (seconds).
- `local_gateway_nodes` (Set of String) Set of IDs of switch nodes which will be configured to peer with the External Gateway
- `password` (String, Sensitive) BGP TCP authentication password
- `ttl` (Number) BGP Time To Live. Omit to use device defaults.