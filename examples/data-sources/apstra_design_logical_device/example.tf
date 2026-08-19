# The following example creates a Logical Device then retrives it by name and by ID.

resource "apstra_design_logical_device" "example" {
  name = "My Logical Device "
  panels = [
    {
      rows    = 25
      columns = 2
      port_groups = [
        {
          port_count = 48
          port_speed = "1G"
          # port_roles = ["superspine", "spine", "leaf", "peer", "access", "generic"]
        },
        {
          port_count = 2
          port_speed = "25G"
          # port_roles = ["superspine", "spine", "leaf", "peer", "access", "generic"]
        }
      ]
    }
  ]
}

data "apstra_design_logical_device" "by_name" {
  name = apstra_design_logical_device.example.name
}

data "apstra_design_logical_device" "by_id" {
  id = apstra_design_logical_device.example.id
}
