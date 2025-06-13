# This example shows how to pull details (the description) of a Tag in a
# Datacenter Blueprint

data "apstra_datacenter_tag" "prod" {
  blueprint_id = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
  name         = "prod"
}

output "prod_tag" { value = data.apstra_datacenter_tag.prod }

# The output looks like this:
# prod_tag = {
#   "blueprint_id" = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
#   "description" = "Production Systems Only"
#   "name" = "prod"
# }
