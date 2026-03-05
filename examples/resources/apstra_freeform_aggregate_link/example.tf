# The apstra_freeform_aggregate_link resource creates an
# aggregate link in a freeform blueprint.

resource "apstra_freeform_aggregate_link" "multi-headed" {
  blueprint_id    = "6791e524-fd70-4f73-ba7b-5985ecd16d68"
  name            = "lag name"
  member_link_ids = [ "nv2JYHNY9B3X1cqHRw", "drwEBZLF0SgLOg8ZKQ" ]

  # endpoint_group_1 is the system or systems on the left side of the
  # freeform LAG management UI. In this case, two endpoints suggests an
  # MLAG or ESI-LAG configuration.
  endpoint_group_1 = {
    endpoints = [
      {
        system_id       = "Bzbh3t9Qx1z5VsGynw"
        port_channel_id = 150
        lag_mode        = "lacp_active"
      },
      {
        system_id       = "PH_DDxHfy76lL34KEw"
        port_channel_id = 150
        lag_mode        = "lacp_active"
      }
    ]
  }
  # endpoint_group_2 is the system or systems on the right side of the
  # freeform LAG management UI. In this case we have a single endpoint
  # with no MLAG or ESI-LAG configuration.
  endpoint_group_2 = {
    endpoints = [
      {
        system_id       = "clZCVKkWuORs1AdmUQ"
        port_channel_id = 6
        lag_mode        = "lacp_passive"
      }
    ]
  }
}
