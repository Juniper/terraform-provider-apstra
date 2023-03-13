# This example creates a logical device with 48x1G
# and 2x25G ports on a single panel.
resource "apstra_logical_device" "example" {
  name = "example device with 48x1G and 2x25G"
  panels = [
    {
      rows = 2
      columns = 25
      port_groups = [
        {
          port_count = 48
          port_speed = "1G"
          port_roles = ["superspine", "spine", "leaf", "peer", "access", "generic"]
        },
        {
          port_count = 2
          port_speed = "25G"
          port_roles = ["superspine", "spine", "leaf", "peer", "access", "generic"]
        },
      ]
    }
  ]
}
