# The following example creates a 'system' type (can be applied to 'system'
# graph nodes) Connectivity Template with two Custom Static Route primitives.
resource "apstra_datacenter_connectivity_template_system" "DC_1" {
  blueprint_id = "275769da-7b45-47d6-8f1c-49323d346bb3"
  name         = "DC 1"
  description  = "Routes to 10.1.0.0/16 in RZ A and B"
  custom_static_routes = [
    {
      name            = "RZ A"
      routing_zone_id = "XYDlrGPmbxnBCS6BH1U"
      network         = "10.1.0.0/16"
      next_hop        = "192.168.1.1"
    },
    {
      name            = "RZ B"
      routing_zone_id = "aAcbBbVe3WTJat8Y7nY"
      network         = "10.1.0.0/16"
      next_hop        = "192.168.1.1"
    },
  ]
}