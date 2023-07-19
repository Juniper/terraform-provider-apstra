# This example pulls a routing policy by name and makes it
# avalilable via `output`
data "apstra_datacenter_routing_policy" "test" {
  blueprint_id = "f30a5572-b899-408a-9195-e4d5bc1118cb" # required attribute
  # id         = "CX05RiY6Zcp32h7nPE0"                  # optional attribute
  name         = "test-label-d1f12"                     # optional attribute
}

output "test_policy" { value = data.apstra_datacenter_routing_policy.test }

# the output looks like:
#
# test_policy = {
#   "aggregate_prefixes" = tolist([
#     "1.0.0.0/8",
#     "2.0.0.0/7",
#   ])
#   "blueprint_id" = "f30a5572-b899-408a-9195-e4d5bc1118cb"
#   "description" = "test-description-d1f12"
#   "expect_default_ipv4" = true
#   "expect_default_ipv6" = true
#   "export_policy" = {
#     "export_l2_edge_subnets" = true
#     "export_l3_edge_server_links" = true
#     "export_loopbacks" = true
#     "export_spine_leaf_links" = true
#     "export_spine_superspine_links" = true
#     "export_static_routes" = true
#   }
#   "extra_exports" = tolist([
#     {
#       "action" = "permit"
#       "ge_mask" = 11
#       "le_mask" = 13
#       "prefix" = "200.0.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = 12
#       "le_mask" = 14
#       "prefix" = "200.64.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = 12
#       "le_mask" = tonumber(null)
#       "prefix" = "200.128.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = tonumber(null)
#       "le_mask" = 14
#       "prefix" = "200.192.0.0/10"
#     },
#     {
#       "action" = "permit"
#       "ge_mask" = 11
#       "le_mask" = 13
#       "prefix" = "210.0.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = 12
#       "le_mask" = 14
#       "prefix" = "210.0.0.0/10"
#     },
#   ])
#   "extra_imports" = tolist([
#     {
#       "action" = "permit"
#       "ge_mask" = 11
#       "le_mask" = 13
#       "prefix" = "100.0.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = 12
#       "le_mask" = 14
#       "prefix" = "100.64.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = 11
#       "le_mask" = tonumber(null)
#       "prefix" = "100.128.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = tonumber(null)
#       "le_mask" = 11
#       "prefix" = "100.192.0.0/10"
#     },
#     {
#       "action" = "permit"
#       "ge_mask" = 11
#       "le_mask" = 13
#       "prefix" = "110.0.0.0/10"
#     },
#     {
#       "action" = "deny"
#       "ge_mask" = 12
#       "le_mask" = 14
#       "prefix" = "110.0.0.0/10"
#     },
#   ])
#   "id" = "CX05RiY6Zcp32h7nPE0"
#   "import_policy" = "default_only"
#   "name" = "test-label-d1f12"
# }
