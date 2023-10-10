# This example discovers the IDs of all interfaces associated
# with the switch labeled "leaf_1"

# First find the IDs of all switches (just one, obviously)
# that sport the label "leaf_1"
data "apstra_datacenter_systems" "only_leaf_1" {
  blueprint_id = "fa6782cc-c4d5-4933-ad89-e542acd6b0c1"
  filter = {
    label = "leaf_1"
  }
}

# Next, use the discovered IDs when querying for interfaces
# by system ID. The one() function here does two things:
# - it asserts that exactly one result must appear in the 'ids' attribute
# - it extracts that single result from the set
data "apstra_datacenter_interfaces_by_system" "test" {
  blueprint_id = "fa6782cc-c4d5-4933-ad89-e542acd6b0c1"
  system_id = one(data.apstra_datacenter_systems.only_leaf_1.ids)
}

# The interesting output of apstra_datacenter_interfaces_by_system.test
# is in the 'if_map' attribute:
#
#   "if_map" = tomap({
#     "ae1" = "Pyu5ONSkPaJ36mwaRqQ"
#     "ae2" = "oaMT0oZSnMZcY-RGY6U"
#     "lo0.0" = "l2_esi_2x_links_001_leaf1_loopback"
#     "lo0.2" = "sz:7l7VRPThwuk4BOS_Q1s,l2_esi_2x_links_001_leaf1_loopback"
#     "xe-0/0/0" = "dtjh9C35R-vJxWySSdc"
#     "xe-0/0/1" = "iYJddO9DZ6qsJ9g6l6I"
#     "xe-0/0/2" = "gZuiJl7UuITShYNviC4"
#     "xe-0/0/3" = "07pZMOrbUYMIQy951bw"
#     "xe-0/0/4" = "pQ06xbk_k5s74360gmM"
#     "xe-0/0/5" = "wjYng7paCYXJ_dYPdtk"
#   })
