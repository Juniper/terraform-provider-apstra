# This example discovers DCI Layer 3 Policy for the "green" VRF

locals { blueprint_id = "4ef73591-d23d-4f66-9ab2-16fbf156bfac" }

# Discover all DCI Interconnect domains (we expect exactly one)
data "apstra_datacenter_interconnect_domains" "dci" {
  blueprint_id = local.blueprint_id
}

# Discover the ID of RZ "green"
data "apstra_datacenter_routing_zone" "green" {
  blueprint_id = local.blueprint_id
  name         = "green"
}

data "apstra_datacenter_interconnect_domain_layer_3_policy" "green_dci" {
  blueprint_id           = local.blueprint_id
  interconnect_domain_id = one(data.apstra_datacenter_interconnect_domains.dci.ids)
  routing_zone_id        = data.apstra_datacenter_routing_zone.green.id
}

output "green_dci_l3_policy" { value = data.apstra_datacenter_interconnect_domain_layer_3_policy.green_dci }

# The output looks like this:
#
# green_dci_l3_policy = {
#   "blueprint_id" = "4ef73591-d23d-4f66-9ab2-16fbf156bfac"
#   "enabled_for_type_5" = true
#   "interconnect_domain_id" = "uiRYxcMfPZtjRUbwUg"
#   "route_target" = "64513:101"
#   "routing_policy_id" = "887-9ECW9Zbr8-DBYg"
#   "routing_zone_id" = "8-NPmOaw3Mw9VQ8P_w"
# }
