# This example creates a Tag in a Datacenter Blueprint

resource "apstra_datacenter_tag" "prod" {
  blueprint_id = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
  name         = "prod"
  description = "Production Systems Only"
}
