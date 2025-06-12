# This example pulls connectivity template status information from a blueprint
data "apstra_datacenter_connectivity_templates_status" "example" {
  blueprint_id = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
}

output "ct_status" { value = data.apstra_datacenter_connectivity_templates_status.example }

# The result looks like this:
# ct_status = {
#  "blueprint_id" = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
#  "connectivity_templates" = tomap({
#    "0e7577cf-78f0-4b00-94a5-01e48a61a0b8" = {
#      "application_point_count" = 3
#      "description" = "The first CT"
#      "id" = "0e7577cf-78f0-4b00-94a5-01e48a61a0b8"
#      "name" = "CT One"
#      "status" = "assigned"
#      "tags" = toset([
#        "prod",
#        "east",
#      ])
#    }
#    "7f4428e8-0712-4e67-a76a-29ac76cfa2bf" = {
#      "application_point_count" = 0
#      "description" = "The second CT"
#      "id" = "7f4428e8-0712-4e67-a76a-29ac76cfa2bf"
#      "name" = "CT Two"
#      "status" = "ready"
#      "tags" = toset(null) /* of string */
#    }
#  })
#}