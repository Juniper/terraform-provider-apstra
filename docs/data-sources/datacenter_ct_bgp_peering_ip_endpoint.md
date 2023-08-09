---
page_title: "apstra_datacenter_ct_bgp_peering_ip_endpoint Data Source - terraform-provider-apstra"
subcategory: ""
description: |-
  This data source composes a Connectivity Template Primitive as a JSON string, suitable for use in the primitives attribute of an apstra_datacenter_connectivity_template resource or the child_primitives attribute of a Different Connectivity Template Primitive.
---

# apstra_datacenter_ct_bgp_peering_ip_endpoint (Data Source)

This data source composes a Connectivity Template Primitive as a JSON string, suitable for use in the `primitives` attribute of an `apstra_datacenter_connectivity_template` resource or the `child_primitives` attribute of a Different Connectivity Template Primitive.

## Example Usage

```terraform
# Each apstra_datacenter_ct_* data source represents a Connectivity Template
# Primitive. They're stand-ins for the Primitives found in the Web UI's CT
# builder interface.
#
# These data sources do not interact with the Apstra API. Instead, they assemble
# their input fields into a JSON string presented at the `primitive` attribute
# key.
#
# Use the `primitive` output anywhere you need a primitive represented as JSON:
# - at the root of a Connectivity Template
# - as a child of another Primitive (as constrained by the accepts/produces
#   relationship between Primitives)

# Declare a "BGP Peering (Generic System)" Connectivity Template Primitive:
data "apstra_datacenter_ct_bgp_peering_ip_endpoint" "a" {
  bfd_enabled      = true
  ttl              = 1
  password         = "big secret"
  ipv4_address     = "192.168.10.5"
}

# This data source's `primitive` attribute produces JSON like this:
# {
#   "type": "AttachIpEndpointWithBgpNsxt",
#   "data": {
#     "neighbor_asn": null,
#     "neighbor_asn_dynaimc": true,
#     "ipv4_afi_enabled": true,
#     "ipv6_afi_enabled": false,
#     "ttl": 1,
#     "bfd_enabled": true,
#     "password": "big secret",
#     "keepalive_time": 2,
#     "hold_time": 3,
#     "local_asn": null,
#     "ipv4_address": "192.168.10.5",
#     "ipv6_address": null,
#     "child_primitives": null
#   }
# }

# Use the `primitive` element (JSON string) when declaring a parent primitive:
data "apstra_datacenter_ct_ip_link" "ip_link_with_bgp" {
  routing_zone_id      = "Zplm0niOFCCCfjaXkXo"
  vlan_id              = 3
  ipv4_addressing_type = "numbered"
  ipv6_addressing_type = "link_local"
  child_primitives = [
    data.apstra_datacenter_ct_bgp_peering_ip_endpoint.a.primitive
  ]
}

# The IP Link data source's `primitive` field has the BGP data
# source (child primitive) as an embedded string:
# {
#   "type": "AttachLogicalLink",
#   "data": {
#     "routing_zone_id": "Zplm0niOFCCCfjaXkXo",
#     "tagged": true,
#     "vlan_id": 3,
#     "ipv4_addressing_type": "numbered",
#     "ipv6_addressing_type": "link_local",
#     "child_primitives": [
#       "{\"type\":\"AttachIpEndpointWithBgpNsxt\",\"data\":{\"neighbor_asn\":null,\"neighbor_asn_dynaimc\":true,\"ipv4_afi_enabled\":true,\"ipv6_afi_enabled\":false,\"ttl\":1,\"bfd_enabled\":true,\"password\":\"big secret\",\"keepalive_time\":null,\"hold_time\":null,\"local_asn\":null,\"ipv4_address\":\"192.168.10.5\",\"ipv6_address\":null,\"child_primitives\":null}}"
#     ]
#   }
# }

# Finally, use the IP Link's `primitive` element in a Connectivity Template:
resource "apstra_datacenter_connectivity_template" "t" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name         = "bgp"
  tags         = [
    "test",
    "terraform",
  ]
  primitives   = [
    data.apstra_datacenter_ct_ip_link.ip_link_with_bgp.primitive
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `bfd_enabled` (Boolean) Enable BFD.
- `child_primitives` (Set of String) Set of JSON strings describing Connectivity Template Primitives which are children of this Connectivity Template Primitive. Use the `primitive` attribute of other Connectivity Template Primitives data sources here.
- `hold_time` (Number) BGP hold time (seconds).
- `ipv4_address` (String) IPv4 address of peer
- `ipv6_address` (String) IPv6 address of peer
- `keepalive_time` (Number) BGP keepalive time (seconds).
- `local_asn` (Number) This feature is configured on a per-peer basis. It allows a router to appear to be a member of a second autonomous system (AS) by prepending a local-as AS number, in addition to its real AS number, announced to its eBGP peer, resulting in an AS path length of two.
- `neighbor_asn` (Number) Neighbor ASN. Omit for *Neighbor ASN Type Dynamic*.
- `password` (String)
- `ttl` (Number) BGP Time To Live. Omit to use device defaults.

### Read-Only

- `primitive` (String) JSON output for use in the `primitives` field of an `apstra_datacenter_connectivity_template` resource or a different Connectivity Template JsonPrimitive data source