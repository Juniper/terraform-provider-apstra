# This example reports on un-mapped interfaces found in various interface
# maps. These are cases where a particular switch model has more interfaces
# than are required for the fabric role to which it might be assigned (using
# a 48 port switch where the design only requires 12 ports, etc...)

# Grab a list of interesting Interface Map IDs. In this case we're
# interested only in interface maps which link hardware to the
# "AOS-7x10-Leaf" Logical Device design element.
data "apstra_interface_maps" "imaps" {
  logical_device_id = "AOS-7x10-Leaf"
}

# Loop over the matching Interface Map IDs and grab the full details of
# those objects so that we can inspect them further.
data "apstra_interface_map" "details" {
  for_each = data.apstra_interface_maps.imaps.ids
  id       = each.key
}

# Report stage: Loop over Interface Map objects, and then the interface
# mappings found inside. Filter the mappings for instances where the
# logical device panel and port are null (un-mapped). Output map
# (dictionary) where the key is Interface Map ID, and the value is the count
# of un-mapped interfaces.
output "unmapped_interface_count" {
  value = {
    for i in data.apstra_interface_map.details : i.id => length(
      [
        for j in i.interfaces : j if j.mapping.logical_device_panel == null && j.mapping.logical_device_panel_port == null
      ]
    )
  }
}

############################################################################
# The output looks like this:
#
#   unmapped_interface_count = {
#     "Arista_vEOS__AOS-7x10-Leaf" = 0
#     "Cisco_NXOSv__AOS-7x10-Leaf" = 2
#     "Cumulus_VX__AOS-7x10-Leaf" = 17
#     "Juniper_vQFX__AOS-7x10-Leaf" = 5
#     "VS_SONiC_BUZZNIK_PLUS__AOS-7x10-Leaf" = 49
#   }
############################################################################
