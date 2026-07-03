# This example creates a switching zone within an
# existing datacenter blueprint.
resource "apstra_datacenter_switching_zone" "blue" {
  blueprint_id = "f526f3dd-d814-4a1c-bd4b-950eabf82d13"
  name         = "blue"
  mac_vrf_name = "blue"
  service_type = "vlan_aware"
  route_target = "64496:10001"
}
