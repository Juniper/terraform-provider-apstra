resource "apstra_rack_type" "rt" {
  name = "terraform"
  description = "terraform did this"
  fabric_connectivity_design = "l3clos"
  leaf_switches = [
    {
      name = "leafy"
      logical_device_id = "AOS-7x10-Leaf"
      spine_link_count = 1
      spine_link_speed = "10G"
    }
  ]
}
