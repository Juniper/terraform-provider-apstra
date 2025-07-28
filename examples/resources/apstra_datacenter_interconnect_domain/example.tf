# This example creates an Interconnect Domain

resource "apstra_datacenter_interconnect_domain" "example" {
  blueprint_id = "00e15432-5012-4777-9f8a-da811bb8d896"
  name         = "DC1"
  route_target = "100:100"
}
