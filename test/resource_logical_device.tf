resource "apstra_logical_device" "t" {
  name = "foo"
  panels = [
    {
      rows = 2
      columns = 3
      port_groups = [
        {
          port_count = 6
          port_speed = "10G"
          port_roles = ["spine", "leaf"]
        }
      ]
    }
  ]
}
