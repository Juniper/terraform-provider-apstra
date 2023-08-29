#reference a blueprint
data "apstra_datacenter_blueprint" "b" {
	name = "test"
}
#reference a catalog configlet
data "apstra_configlet" "cfg"{
  name = "US-MOTD"
}

#import configlet from catalog
resource "apstra_datacenter_configlet" "rc" {
 blueprint_id = data.apstra_datacenter_blueprint.b.id
  catalog_configlet_id = data.apstra_configlet.cfg.id
  condition = "role in [\"spine\"]"
  name = "helloworld"
}
locals {
    name = "DNS according to Terraforming"
    generators = [
      {
        config_style  = "junos"
        section       = "top_level_hierarchical"
        template_text = <<-EOT
        name-server {
          4.2.2.1;
          4.2.2.2;
        }
      EOT
      },
      {
        config_style           = "eos"
        section                = "system"
        template_text          = "ip name-server 4.2.2.1 4.2.2.2"
        negation_template_text = "no ip name-server 4.2.2.1 4.2.2.4"
      }
    ]
}

#create configlet in blueprint
resource "apstra_datacenter_configlet" "rc1" {
 blueprint_id = data.apstra_datacenter_blueprint.b.id
  condition = "role in [\"leaf\"]"
  name = local.name
  generators = local.generators
}

output "imported" {
  value =  apstra_datacenter_configlet.rc
}

output "created" {
  value =  apstra_datacenter_configlet.rc1
}

#Output looks like this
#created = {
#  "blueprint_id" = "5d4f7b2c-e7c5-4863-a01d-2f5b398a341f"
#  "catalog_configlet_id" = tostring(null)
#  "condition" = "role in [\"leaf\"]"
#  "generators" = tolist([
#    {
#      "config_style" = "junos"
#      "filename" = tostring(null)
#      "negation_template_text" = tostring(null)
#      "section" = "top_level_hierarchical"
#      "template_text" = <<-EOT
#      name-server {
#        4.2.2.1;
#        4.2.2.2;
#      }
#
#      EOT
#    },
#    {
#      "config_style" = "eos"
#      "filename" = tostring(null)
#      "negation_template_text" = "no ip name-server 4.2.2.1 4.2.2.4"
#      "section" = "system"
#      "template_text" = "ip name-server 4.2.2.1 4.2.2.2"
#    },
#  ])
#  "id" = "zoyTkvrg9gW5it-HnpI"
#  "name" = "DNS according to Terraforming"
#}
#imported = {
#  "blueprint_id" = "5d4f7b2c-e7c5-4863-a01d-2f5b398a341f"
#  "catalog_configlet_id" = "US-MOTD"
#  "condition" = "role in [\"spine\"]"
#  "generators" = tolist([
#    {
#      "config_style" = "nxos"
#      "filename" = tostring(null)
#      "negation_template_text" = <<-EOT
#      no banner login
#
#      EOT
#      "section" = "system"
#      "template_text" = <<-EOT
#      banner login #
#      --------------------------------------------------------
#      UNAUTHORIZED ACCESS TO THIS DEVICE IS PROHIBITED
#      You must have explicit, authorized permission to access
#      or configure this device. Unauthorized attempts and
#      actions to access or use this system may result in civil
#      and/or criminal penalties. All activities performed on
#      this device are logged and monitored.
#      --------------------------------------------------------
#      #
#
#      EOT
#    },
#    {
#      "config_style" = "eos"
#      "filename" = tostring(null)
#      "negation_template_text" = <<-EOT
#      no banner motd
#
#      EOT
#      "section" = "system"
#      "template_text" = <<-EOT
#      banner motd
#      --------------------------------------------------------
#      UNAUTHORIZED ACCESS TO THIS DEVICE IS PROHIBITED
#      You must have explicit, authorized permission to access
#      or configure this device. Unauthorized attempts and
#      actions to access or use this system may result in civil
#      and/or criminal penalties. All activities performed on
#      this device are logged and monitored.
#      --------------------------------------------------------
#      EOF
#
#      EOT
#    },
#  ])
#  "id" = "WS-pwJkupI2yXO8V1jM"
#  "name" = "helloworld"
#}
