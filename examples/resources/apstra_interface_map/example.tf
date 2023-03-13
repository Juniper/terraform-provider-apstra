# Interface Maps have an "interfaces" section which can get tedious because it
# involves spelling out the relationship between each Logical Device interface
# with the corresponding Device Profile interface.

# We'll create two interface maps here:
# - QFX5120-48T -> AOS-7x10-Spine (typing out all of the details)
# - QFX5120-48T -> AOS-48x10+6x40-1 (using loops)

# First example: we type out all of the interface mappings.
resource "apstra_interface_map" "the_hard_way" {
  name              = "example interface map 1"
  logical_device_id = "AOS-7x10-Spine"
  device_profile_id = "Juniper_QFX5120-48T_Junos"
  interfaces = [
    {
      "logical_device_port"     = "1/1"
      "physical_interface_name" = "xe-0/0/0"
      "transformation_id"       = 1
      # transform #1 is the only transform which can fulfill the Logical
      # Device requirement with this equipment, making it optional (the
      # provider infers the transform ID in that case), so it's omitted
      # from the interface mappings which follow.
    },
    {
      "logical_device_port"     = "1/2"
      "physical_interface_name" = "xe-0/0/1"
    },
    {
      "logical_device_port"     = "1/3"
      "physical_interface_name" = "xe-0/0/2"
    },
    {
      "logical_device_port"     = "1/4"
      "physical_interface_name" = "xe-0/0/3"
    },
    {
      "logical_device_port"     = "1/5"
      "physical_interface_name" = "xe-0/0/4"
    },
    {
      "logical_device_port"     = "1/6"
      "physical_interface_name" = "xe-0/0/5"
    },
    {
      "logical_device_port"     = "1/7"
      "physical_interface_name" = "xe-0/0/6"
    },
  ]
}

# Local variables which help build up the interface mapping
# used in the resource.
locals {
  # local.if_map spells out two ranges of mappings for our
  # second interface map: 48x10G ports and 6x40G ports.
  if_map = [
    { // map logical 1/1 - 1/48 to physical xe-0/0/0 - xe-0/0/47
      ld_panel       = 1
      ld_first_port  = 1
      phy_prefix     = "xe-0/0/"
      phy_first_port = 0
      count          = 48
    },
    { // map logical 2/1 - 2/6 to physical et-0/0/48 - et-0/0/53
      ld_panel       = 2
      ld_first_port  = 1
      phy_prefix     = "et-0/0/"
      phy_first_port = 48
      count          = 6
    },
  ]
  # local.interfaces loops over the elements of if_map (panel 1 and panel 2).
  # within each iteration, it loops 'count' times (every interface in the panel)
  # to build up the detailed mapping between logical and physical ports.
  interfaces = [
    for map in local.if_map : [
      for i in range(map.count) : {
        logical_device_port     = format("%d/%d", map.ld_panel, map.ld_first_port + i)
        physical_interface_name = format("%s%d", map.phy_prefix, map.phy_first_port + i)
      }
    ]
  ]
}

# second example: interfacemappings are calculated
# using the local variables above.
resource "apstra_interface_map" "with_loops" {
  name              = "example interface map 2"
  logical_device_id = "AOS-48x10_6x40-1"
  device_profile_id = "Juniper_QFX5120-48T_Junos"
  interfaces        = flatten([local.interfaces])
}