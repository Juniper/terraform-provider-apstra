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

data "apstra_datacenter_configlet" "junos_configlets" {
  for_each = data.apstra_datacenter_configlets.junos_only.ids
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  id = each.key
}

data "apstra_datacenter_configlet" "all_configlets" {
  for_each = data.apstra_datacenter_configlets.all.ids
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  id = each.key
}

output "all_blueprint_configlets" {
  value =  data.apstra_datacenter_configlets.all
}

# Use the ID numbers to pull full details of those configlets.
output "blueprint_configlets_junos" {
  value = data.apstra_datacenter_configlet.junos_configlets
}

# Loop over all configlets. Within each configlet, filter out Junos-style
# generators. Count the generators.
output "all_blueprint_configlet_names" {
  value = [for c in data.apstra_datacenter_configlet.all_configlets : c.name]
}

#Output looks something like this
#all_blueprint_configlets = {
#"blueprint_id" = "5d4f7b2c-e7c5-4863-a01d-2f5b398a341f"
#"ids" = toset([
#"BtJ6dr2Rg002XLP6kec",
#"bisaaVp66kSkdOwLe2A",
#])
#"supported_platforms" = toset(null) /* of string */
#}
#blueprint_configlets_junos = {
#"BtJ6dr2Rg002XLP6kec" = {
#"blueprint_id" = "5d4f7b2c-e7c5-4863-a01d-2f5b398a341f"
#"catalog_configlet_id" = tostring(null)
#"condition" = "role in [\"leaf\"]"
#"generators" = tolist([
#{
#"config_style" = "junos"
#"filename" = tostring(null)
#"negation_template_text" = tostring(null)
#"section" = "top_level_hierarchical"
#"template_text" = <<-EOT
#        name-server {
#          4.2.2.1;
#          4.2.2.2;
#        }
#
#        EOT
#},
#{
#"config_style" = "eos"
#"filename" = tostring(null)
#"negation_template_text" = "no ip name-server 4.2.2.1 4.2.2.4"
#"section" = "system"
#"template_text" = "ip name-server 4.2.2.1 4.2.2.2"
#},
#])
#"id" = "BtJ6dr2Rg002XLP6kec"
#"name" = "DNS according to Terraforming"
#}
#}
#all_blueprint_configlet_names = [
#"DNS according to Terraforming",
#"helloworld",
#]
