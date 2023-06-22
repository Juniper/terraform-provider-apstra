# This example pulls the details of all property sets
# that have been imported into a blueprint

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_datacenter_property_sets" "ps" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
}

output "o" {
  value = data.apstra_datacenter_property_sets.ps
}

#Output looks something like this
#o = {
#  "blueprint_id" = "621a5671-970f-402c-ae13-6b83834b5255"
#  "ids" = toset([
#    "a60a45fe-7466-48e1-8ca7-b6008de1c1e5",
#  ])
#}
