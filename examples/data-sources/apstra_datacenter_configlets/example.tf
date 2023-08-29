# This example uses the `apstra_datacenter_configlets` data source to get a list
# of all imported configlets, and then uses the apstra_datacenter_configlet data source
# to inspect the results


data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_datacenter_configlets" "all" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
}

data "apstra_datacenter_configlets" "junos_only" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  supported_platforms = ["junos"]
}
output "all_blueprint_configlets" {
  value =  data.apstra_datacenter_configlets.all
}

# Use the ID numbers to pull full details of those configlets.
output "blueprint_configlets_junos" {
  value = data.apstra_datacenter_configlets.junos_only
}

#Output looks something like
#all_blueprint_configlets = {
#  "blueprint_id" = "5d4f7b2c-e7c5-4863-a01d-2f5b398a341f"
#  "ids" = toset([
#    "BtJ6dr2Rg002XLP6kec",
#    "bisaaVp66kSkdOwLe2A",
#  ])
#  "supported_platforms" = toset(null) /* of string */
#}
#blueprint_configlets_junos = {
#  "blueprint_id" = "5d4f7b2c-e7c5-4863-a01d-2f5b398a341f"
#  "ids" = toset([
#    "BtJ6dr2Rg002XLP6kec",
#  ])
#  "supported_platforms" = toset([
#    "junos",
#  ])
#}

