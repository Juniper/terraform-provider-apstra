# This example shows how to collect the names and graph node IDs of all tags
# with description "firewall"

data "apstra_datacenter_tags" "firewall" {
  blueprint_id = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
  filters = [
    {
      description = "firewall"
    }
  ]
}

output "firewall_tags" { value = data.apstra_datacenter_tags.firewall }

# The output looks like this:
# firewall_tags = {
#   "blueprint_id" = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
#   "filters" = tolist([
#     {
#       "blueprint_id" = tostring(null)
#       "description" = "firewall"
#       "name" = tostring(null)
#     },
#    ])
#    "ids" = toset([
#     "9mF7QTOjTSspgG5FPg",
#     "SyTIdNbDbm8QhBEgOw",
#   ])
#   "names" = toset([
#     "firewall-a",
#     "firewall-b",
#   ])
# }

