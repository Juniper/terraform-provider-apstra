locals {
  if_map = [
    { // map ld 1/1 - 1/2 to et-0/0/48 - et-0/0/49
      ld_prefix = "1/"
      ld_first_port = 1
      phy_prefix = "et-0/0/"
      phy_first_port = 48
      count = 2
      transform_id = 3
    },
    { // map ld 1/3 - 1/4 to et-0/0/50 - et-0/0/50
      ld_prefix = "1/"
      ld_first_port = 3
      phy_prefix = "et-0/0/"
      phy_first_port = 52
      count = 2
      transform_id = 3
    },
    { // map ld 2/1 - 2/8 to xe-0/0/0 - xe-0/0/7
      ld_prefix = "2/"
      ld_first_port = 1
      phy_prefix = "xe-0/0/"
      phy_first_port = 0
      count = 8
      transform_id = 1
    },
  ]

  x = flatten([
    for map in local.if_map: [
      for i in range(map.count): {
        logical_device_port = format("%s%d", map.ld_prefix, map.ld_first_port + i)
        physical_interface_name = format("%s%d", map.phy_prefix, map.phy_first_port + i)
        transform_id = map.transform_id
      }
    ]
  ])
}

resource "apstra_interface_map" "if_map" {
  name = "aaa tf_map2"
  logical_device_id = "AOS-4x40_8x10-1"
  device_profile_id = "Juniper_QFX5120-48T_Junos"
  interfaces = flatten([
    for map in local.if_map: [
      for i in range(map.count): {
        logical_device_port = format("%s%d", map.ld_prefix, map.ld_first_port + i)
        physical_interface_name = format("%s%d", map.phy_prefix, map.phy_first_port + i)
        transform_id = map.transform_id
      }
    ]
  ])
#  interfaces = [
#    {
#      "logical_device_port" = "1/1"
#      "physical_interface_name" = "et-0/0/48"
#    },
#    {
#      "logical_device_port" = "1/2"
#      "physical_interface_name" = "et-0/0/49"
#    },
#    {
#      "logical_device_port" = "1/3"
#      "physical_interface_name" = "et-0/0/52"
#    },
#    {
#      "logical_device_port" = "1/4"
#      "physical_interface_name" = "et-0/0/53"
#    },
#    {
#      "logical_device_port" = "2/1"
#      "physical_interface_name" = "xe-0/0/0"
#    },
#    {
#      "logical_device_port" = "2/2"
#      "physical_interface_name" = "xe-0/0/1"
#    },
#    {
#      "logical_device_port" = "2/3"
#      "physical_interface_name" = "xe-0/0/2"
#    },
#    {
#      "logical_device_port" = "2/4"
#      "physical_interface_name" = "xe-0/0/3"
#    },
#    {
#      "logical_device_port" = "2/5"
#      "physical_interface_name" = "xe-0/0/4"
#    },
#    {
#      "logical_device_port" = "2/6"
#      "physical_interface_name" = "xe-0/0/5"
#    },
#    {
#      "logical_device_port" = "2/7"
#      "physical_interface_name" = "xe-0/0/6"
#    },
#    {
#      "logical_device_port" = "2/8"
#      "physical_interface_name" = "xe-0/0/7"
#    },
#  ]
}