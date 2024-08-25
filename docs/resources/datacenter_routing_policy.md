---
page_title: "apstra_datacenter_routing_policy Resource - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This resource creates a Routing Policy within a Blueprint.
---

# apstra_datacenter_routing_policy (Resource)

This resource creates a Routing Policy within a Blueprint.


## Example Usage

```terraform
# This example creates a routing policy within the blueprint named "production"

data "apstra_datacenter_blueprint" "prod" {
  name = "production"
}

resource "apstra_datacenter_routing_policy" "just_pull_every_available_lever" {
  blueprint_id  = data.apstra_datacenter_blueprint.prod.id
  name          = "nope"
  description   = "Nothing good can come from this"
  import_policy = "default_only" // "default_only" is the default. other options: "all" "extra_only"
  extra_imports = [
    { prefix = "10.0.0.0/8",                                 action = "deny"   },
    { prefix = "11.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "deny"   },
    { prefix = "12.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "deny"   },
    { prefix = "13.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "deny"   },
    { prefix = "14.0.0.0/8", ge_mask = 9,                    action = "deny"   },
    { prefix = "15.0.0.0/8", ge_mask = 32,                   action = "deny"   },
    { prefix = "16.0.0.0/8",                 le_mask = 9,    action = "deny"   },
    { prefix = "17.0.0.0/8",                 le_mask = 32,   action = "deny"   },
    { prefix = "20.0.0.0/8",                                 action = "permit" },
    { prefix = "21.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "permit" },
    { prefix = "22.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "permit" },
    { prefix = "23.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "permit" },
    { prefix = "24.0.0.0/8", ge_mask = 9,                    action = "permit" },
    { prefix = "25.0.0.0/8", ge_mask = 32,                   action = "permit" },
    { prefix = "26.0.0.0/8",                 le_mask = 9,    action = "permit" },
    { prefix = "27.0.0.0/8",                 le_mask = 32,   action = "permit" },
    { prefix = "30.0.0.0/8",                                                   }, // default action is "permit"
    { prefix = "31.0.0.0/8", ge_mask = 31,   le_mask = 32,                     }, // default action is "permit"
    { prefix = "32.0.0.0/8", ge_mask = 9,    le_mask = 10,                     }, // default action is "permit"
    { prefix = "33.0.0.0/8", ge_mask = 9,    le_mask = 32,                     }, // default action is "permit"
    { prefix = "34.0.0.0/8", ge_mask = 9,                                      }, // default action is "permit"
    { prefix = "35.0.0.0/8", ge_mask = 32,                                     }, // default action is "permit"
    { prefix = "36.0.0.0/8",                 le_mask = 9,                      }, // default action is "permit"
    { prefix = "37.0.0.0/8",                 le_mask = 32,                     }, // default action is "permit"
  ]
  extra_exports = [
    { prefix = "40.0.0.0/8",                                 action = "deny"   },
    { prefix = "41.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "deny"   },
    { prefix = "42.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "deny"   },
    { prefix = "43.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "deny"   },
    { prefix = "44.0.0.0/8", ge_mask = 9,                    action = "deny"   },
    { prefix = "45.0.0.0/8", ge_mask = 32,                   action = "deny"   },
    { prefix = "46.0.0.0/8",                 le_mask = 9,    action = "deny"   },
    { prefix = "47.0.0.0/8",                 le_mask = 32,   action = "deny"   },
    { prefix = "50.0.0.0/8",                                 action = "permit" },
    { prefix = "51.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "permit" },
    { prefix = "52.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "permit" },
    { prefix = "53.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "permit" },
    { prefix = "54.0.0.0/8", ge_mask = 9,                    action = "permit" },
    { prefix = "55.0.0.0/8", ge_mask = 32,                   action = "permit" },
    { prefix = "56.0.0.0/8",                 le_mask = 9,    action = "permit" },
    { prefix = "57.0.0.0/8",                 le_mask = 32,   action = "permit" },
    { prefix = "60.0.0.0/8",                                                   }, // default action is "permit"
    { prefix = "61.0.0.0/8", ge_mask = 31,   le_mask = 32,                     }, // default action is "permit"
    { prefix = "62.0.0.0/8", ge_mask = 9,    le_mask = 10,                     }, // default action is "permit"
    { prefix = "63.0.0.0/8", ge_mask = 9,    le_mask = 32,                     }, // default action is "permit"
    { prefix = "64.0.0.0/8", ge_mask = 9,                                      }, // default action is "permit"
    { prefix = "65.0.0.0/8", ge_mask = 32,                                     }, // default action is "permit"
    { prefix = "66.0.0.0/8",                 le_mask = 9,                      }, // default action is "permit"
    { prefix = "67.0.0.0/8",                 le_mask = 32,                     }, // default action is "permit"
  ]
  export_policy = {
#    export_spine_leaf_links       = false // default value is "false"
#    export_spine_superspine_links = false // default value is "false"
#    export_l3_edge_server_links   = false // default value is "false"
    export_l2_edge_subnets        = true
    export_loopbacks              = false // but it's okay if you type it in
    export_static_routes          = null  // you can also use "null" to get "false"
  }
  aggregate_prefixes = [
    "0.0.0.0/0",  // any prefix is okay here...
    "0.0.0.0/1",
    "0.0.0.0/2",
    "0.0.0.0/3",  // but it has to land on a "zero address"...
    "0.0.0.0/4",
    "0.0.0.0/5",
    "0.0.0.0/6",  // so, "1.0.0.0/6" would not be okay...
    "0.0.0.0/7",
    "0.0.0.0/8",
    "1.0.0.0/9",  // but "1.0.0.0/9" is fine...
    "0.0.0.0/10",
    "0.0.0.0/11",
    "0.0.0.0/12", // Burma-Shave
    "0.0.0.0/13",
    "0.0.0.0/14",
    "0.0.0.0/15",
    "0.0.0.0/16",
    "0.0.0.0/17",
    "0.0.0.0/18",
    "0.0.0.0/19",
    "0.0.0.0/20",
    "0.0.0.0/21",
    "0.0.0.0/22",
    "0.0.0.0/23",
    "0.0.0.0/24",
    "0.0.0.0/25",
    "0.0.0.0/26",
    "0.0.0.0/27",
    "0.0.0.0/28",
    "0.0.0.0/29",
    "0.0.0.0/30",
    "0.0.0.0/31",
    "0.0.0.0/32",
    "255.255.255.255/32"
  ]
  expect_default_ipv4 = true
#  expect_default_ipv6 = false // default value is "false"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID.
- `name` (String) Web UI 'name' field.

### Optional

- `aggregate_prefixes` (List of String) BGP Aggregate routes to be imported into a routing zone (VRF) on all border switches. This option can only be set on routing policies associated with routing zones, and cannot be set on per-connectivity point policies. The aggregated routes are sent to all external router peers in a SZ (VRF).
- `description` (String) Web UI 'description' field.
- `expect_default_ipv4` (Boolean) Default IPv4 route is expected to be imported via protocol session using this policy. Used for rendering route expectations.
- `expect_default_ipv6` (Boolean) Default IPv6 route is expected to be imported via protocol session using this policy. Used for rendering route expectations.
- `export_policy` (Attributes) The export policy controls export of various types of fabric prefixes. (see [below for nested schema](#nestedatt--export_policy))
- `extra_exports` (Attributes List) User defined export routes will be used in addition to any other routes specified in export policies. These policies are additive. To advertise only extra routes, disable all export types within 'export_policy', and only the extra prefixes specified here will be advertised. (see [below for nested schema](#nestedatt--extra_exports))
- `extra_imports` (Attributes List) User defined import routes will be used in addition to any routes generated by the import policies. Prefixes specified here are additive to the import policy, unless 'import_policy' is set to "extra_only", in which only these routes will be imported. (see [below for nested schema](#nestedatt--extra_imports))
- `import_policy` (String) One of '', 'all', 'default_only', 'extra_only'

### Read-Only

- `id` (String) Apstra graph node ID.

<a id="nestedatt--export_policy"></a>
### Nested Schema for `export_policy`

Optional:

- `export_l2_edge_subnets` (Boolean) Exports all virtual networks (VLANs) that have L3 addresses within a routing zone (VRF).
- `export_l3_edge_server_links` (Boolean) Exports all leaf to L3 server links within a routing zone (VRF). This will be an empty list on a layer2 based blueprint
- `export_loopbacks` (Boolean) Exports all loopbacks within a routing zone (VRF) across spine, leaf, and L3 servers.
- `export_spine_leaf_links` (Boolean) Exports all spine-supersine (fabric) links within the default routing zone (VRF)
- `export_spine_superspine_links` (Boolean) Exports all spine-leaf (fabric) links within a VRF. EVPN routing zones do not have spine-leaf addressing, so this generated list may be empty. For routing zones of type Virtual L3 Fabric, subinterfaces between spine-leaf will be included.
- `export_static_routes` (Boolean) Exports all subnets in a VRF associated with static routes from all fabric systems to external routers associated with this routing policy


<a id="nestedatt--extra_exports"></a>
### Nested Schema for `extra_exports`

Required:

- `prefix` (String) IPv4 or IPv6 network address specified in the form of network/prefixlen.

Optional:

- `action` (String) If the action is "permit", match the route. If the action is "deny", do not match the route. For composing complex policies, all prefix-list items will be processed in the order specified, top-down. This allows the user to deny a subset of a route that may otherwise be permitted.
- `ge_mask` (Number) Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.
- `le_mask` (Number) Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.


<a id="nestedatt--extra_imports"></a>
### Nested Schema for `extra_imports`

Required:

- `prefix` (String) IPv4 or IPv6 network address specified in the form of network/prefixlen.

Optional:

- `action` (String) If the action is "permit", match the route. If the action is "deny", do not match the route. For composing complex policies, all prefix-list items will be processed in the order specified, top-down. This allows the user to deny a subset of a route that may otherwise be permitted.
- `ge_mask` (Number) Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.
- `le_mask` (Number) Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.



