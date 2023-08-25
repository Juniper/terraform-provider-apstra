# The apstra_datacenter_svis data source returns a map detailing each SVI
# in the given blueprint. The map is keyed by virtual network ID. Each
# value in the map is a set of SVIs attached to the virtual network.
data "apstra_datacenter_svis" "x" {
  blueprint_id = "fa6782cc-c4d5-4933-ad89-e542acd6b0c1"
}

# The read-only svi_map attribute looks like this:
#
# "svi_map" = tomap({
#   "uQO_xhAW6CAtyTIkYqw" = toset([         // virtual network ID
#     {
#       "ipv4_addr" = "10.0.1.11/24"
#       "ipv4_mode" = "forced"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name"      = "irb.4"
#       "svi_id"    = "mSB9VpYYd1ooOXShrmk"
#       "system_id" = "FaL5bEEo_d3ISL6bQcU"
#     },
#     {
#       "ipv4_addr" = "10.0.1.12/24"
#       "ipv4_mode" = "forced"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name"      = "irb.4"
#       "svi_id"    = "w3ZVUfJl5obihHT7Qyg"
#       "system_id" = "lbo0UdHNhwOSTlJ0lSA"
#     }
#   ])
#   "blTPvfoRx0btyW7Xr8Y" = toset([         // virtual network ID
#     {
#       "ipv4_addr" = "10.0.2.11/24"
#       "ipv4_mode" = "forced"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name"      = "irb.4"
#       "svi_id"    = "VgEADuZiuq4XI1GTHWQ"
#       "system_id" = "VQM_UdeyD4aKAQ9nrPM"
#     },
#     {
#       "ipv4_addr" = "10.0.2.12/24"
#       "ipv4_mode" = "forced"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name"      = "irb.4"
#       "svi_id"    = "OViDU982h-wfRkUJUW0"
#       "system_id" = "w7sTrrFE2XN5Mxy6aS7"
#     }
#   ])
# })