# This example creates an "over the top" DCI External Gateway.
# Note: Prior to Apstra 4.2 these were known as "Remote EVPN Gateways"

resource "apstra_datacenter_external_gateway" "example" {
  blueprint_id     = "b4c4ed6a-9c6a-4577-b3d4-78705c08a272"
  name             = "example gateway"
  ip_address       = "192.0.2.1"
  asn              = 64510
  evpn_route_types = "all" # "all" or "type5_only"
  ttl              = 10
  keepalive_time   = 3
  hold_time        = 9
  password         = "big secret"
  local_gateway_nodes = [
    "JGcTJy_jP4898Z13WHU", // use apstra_datacenter_systems data
    "Fx-fVa7t_LYp7JtQ_nU", // source to find node IDs
  ]
}
