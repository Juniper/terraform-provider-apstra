# This example creates a routing zone within an
# existing datacenter blueprint.
resource "apstra_datacenter_routing_zone" "blue" {
  name              = "vrf blue"
  blueprint_id      = "<blueprint-id-goes-here>"
  vlan_id           = 5                                     # optional
  vni_id            = 5000                                  # optional
#  routing_policy_id = "<routing-policy-node-id-goes-here>" # optional
}
