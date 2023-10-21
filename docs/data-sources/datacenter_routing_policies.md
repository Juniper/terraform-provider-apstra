---
page_title: "apstra_datacenter_routing_policies Data Source - terraform-provider-apstra"
subcategory: ""
description: |-
  This data source returns Graph DB node IDs of routing_policy nodes within a Blueprint.
  Optional filters can be used select only interesting nodes.
---

# apstra_datacenter_routing_policies (Data Source)

This data source returns Graph DB node IDs of *routing_policy* nodes within a Blueprint.

Optional `filters` can be used select only interesting nodes.

## Example Usage

```terraform
# This example returns the IDs of all Routing Policies in the
# Datacenter 1 blueprint which
# - import 10.140/16 and 10.150/16
#   ...or...
# - import 10.240/16 and 10.250/16, and export loopbacks
#
# All attributes specified within a 'filter' block must match.
# If an routing policy is found to match all attributes within
# a filter block, its ID will be included in the computed `ids`
# attribute.
data "apstra_datacenter_blueprint" "DC1" {
  name = "Datacenter 1"
}

data "apstra_datacenter_routing_policies" "all" {
  blueprint_id = data.apstra_datacenter_blueprint.DC1.id
  filters = [
    {
      extra_imports = [
        { prefix = "10.140.0.0/16", action = "permit" },
        { prefix = "10.150.0.0/16", action = "permit" },
      ]
    },
    {
      extra_imports = [
        { prefix = "10.240.0.0/16", action = "permit" },
        { prefix = "10.250.0.0/16", action = "permit" },
      ]
      export_policy = {
        export_loopbacks              = true
      }
    },
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint to search.

### Optional

- `filters` (Attributes List) List of filters used to select only desired node IDs. For a nodeto match a filter, all specified attributes must match (each attribute within a filter is AND-ed together). The returned node IDs represent the nodes matched by all of the filters together (filters are OR-ed together). (see [below for nested schema](#nestedatt--filters))

### Read-Only

- `ids` (Set of String) IDs of matching `routing_policy` Graph DB nodes.

<a id="nestedatt--filters"></a>
### Nested Schema for `filters`

Optional:

- `aggregate_prefixes` (List of String) All `aggregate_prefixes` specified here are required for the filter to match, but the list need not be an *exact match*. That is, a policy containting `10.1.0.0/16` and `10.2.0.0/16` will match a filter which specifies only `10.1.0.0/16`
- `description` (String) Web UI 'description' field.
- `expect_default_ipv4` (Boolean) Default IPv4 route is expected to be imported via protocol session using this policy. Used for rendering route expectations.
- `expect_default_ipv6` (Boolean) Default IPv6 route is expected to be imported via protocol session using this policy. Used for rendering route expectations.
- `export_policy` (Attributes) The export policy controls export of various types of fabric prefixes. (see [below for nested schema](#nestedatt--filters--export_policy))
- `extra_exports` (Attributes List) All `extra_exports` specified here are required for the filter to match, using the same logic as `aggregate_prefixes`. (see [below for nested schema](#nestedatt--filters--extra_exports))
- `extra_imports` (Attributes List) All `extra_imports` specified here are required for the filter to match, using the same logic as `aggregate_prefixes`. (see [below for nested schema](#nestedatt--filters--extra_imports))
- `id` (String) Apstra graph node ID.
- `import_policy` (String) One of '', 'default_only', 'all', 'extra_only'
- `name` (String) Web UI `name` field.

Read-Only:

- `blueprint_id` (String) Not applicable in filter context. Ignore.

<a id="nestedatt--filters--export_policy"></a>
### Nested Schema for `filters.export_policy`

Optional:

- `export_l2_edge_subnets` (Boolean) Exports all virtual networks (VLANs) that have L3 addresses within a routing zone (VRF).
- `export_l3_edge_server_links` (Boolean) Exports all leaf to L3 server links within a routing zone (VRF). This will be an empty list on a layer2 based blueprint
- `export_loopbacks` (Boolean) Exports all loopbacks within a routing zone (VRF) across spine, leaf, and L3 servers.
- `export_spine_leaf_links` (Boolean) Exports all spine-supersine (fabric) links within the default routing zone (VRF)
- `export_spine_superspine_links` (Boolean) Exports all spine-leaf (fabric) links within a VRF. EVPN routing zones do not have spine-leaf addressing, so this generated list may be empty. For routing zones of type Virtual L3 Fabric, subinterfaces between spine-leaf will be included.
- `export_static_routes` (Boolean) Exports all subnets in a VRF associated with static routes from all fabric systems to external routers associated with this routing policy


<a id="nestedatt--filters--extra_exports"></a>
### Nested Schema for `filters.extra_exports`

Optional:

- `action` (String) If the action is "permit", match the route. If the action is "deny", do not match the route. For composing complex policies, all prefix-list items will be processed in the order specified, top-down. This allows the user to deny a subset of a route that may otherwise be permitted.
- `ge_mask` (Number) Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.
- `le_mask` (Number) Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.
- `prefix` (String) IPv4 or IPv6 network address specified in the form of network/prefixlen.


<a id="nestedatt--filters--extra_imports"></a>
### Nested Schema for `filters.extra_imports`

Optional:

- `action` (String) If the action is "permit", match the route. If the action is "deny", do not match the route. For composing complex policies, all prefix-list items will be processed in the order specified, top-down. This allows the user to deny a subset of a route that may otherwise be permitted.
- `ge_mask` (Number) Match less-specific prefixes from a parent prefix, up from `ge_mask` to the prefix length of the route. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `le_mask`. `ge_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.
- `le_mask` (Number) Match more-specific prefixes from a parent prefix, up until `le_mask` prefix len. Range is 0-32 for IPv4, 0-128 for IPv6. If not specified, implies the prefix-list entry should be an exact match. The option can be optionally be used in combination with `ge_mask`. `le_mask` must be longer than the subnet prefix length. If `le_mask` and `ge_mask` are both specified, then `le_mask` must be greater than `ge_mask`.
- `prefix` (String) IPv4 or IPv6 network address specified in the form of network/prefixlen.