# This example uses the `apstra_datacenter_systems` data source and leaf
# switch tags to discover leaf switches in rack 11.
# It then provisions a 4x10G server with two 20G link aggregation groups

# We'll just pretend the blueprint ID is "abc"
locals {
  blueprint_id     = "abc"
  rack_11_leaf_ids = sort(tolist(data.apstra_datacenter_systems.rack_11_leafs.ids))
}

# The IDs field will include all systems matching the tag-based filter
data "apstra_datacenter_systems" "rack_11_leafs" {
  blueprint_id = local.blueprint_id
  filter = {
    tag_ids = ["leaf", "rack 11"]
  }
}

resource "apstra_datacenter_generic_system" "example" {
  blueprint_id      = local.blueprint_id
  name              = "Terraform Did This"
  hostname          = "terraformdidthis.example.com"
  tags              = ["terraform"]
  links = [
    {
      tags                          = ["10G", "terraform", "bond0"]
      lag_mode                      = "lacp_active"
      target_switch_id              = local.rack_11_leaf_ids[0] // first switch
      target_switch_if_name         = "xe-0/0/6"
      target_switch_if_transform_id = 1
      group_label                   = "bond0"
    },
    {
      tags                          = ["10G", "terraform", "bond0"]
      lag_mode                      = "lacp_active"
      target_switch_id              = local.rack_11_leaf_ids[1] // second switch
      target_switch_if_name         = "xe-0/0/6"
      target_switch_if_transform_id = 1
      group_label                   = "bond0"
    },
    {
      tags                          = ["10G", "terraform", "bond1"]
      lag_mode                      = "lacp_active"
      target_switch_id              = local.rack_11_leaf_ids[0] // first switch
      target_switch_if_name         = "xe-0/0/7"
      target_switch_if_transform_id = 1
      group_label                   = "bond1"
    },
    {
      tags                          = ["10G", "terraform", "bond1"]
      lag_mode                      = "lacp_active"
      target_switch_id              = local.rack_11_leaf_ids[1] // second switch
      target_switch_if_name         = "xe-0/0/7"
      target_switch_if_transform_id = 1
      group_label                   = "bond1"
    },
  ]
}
