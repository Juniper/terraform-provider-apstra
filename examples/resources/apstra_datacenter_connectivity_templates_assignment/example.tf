# This example assigns two connectivity templates to the switch port
# identified by the ID "FkYtMBdeoJ5urBaIEi8"
#
# Data sources like these can be used to find node IDs to use in
# the `application_point_id` attribute:
# - apstra_datacenter_svis_map
# - apstra_datacenter_interfaces_by_link_tag
# - apstra_datacenter_interfaces_by_system

resource "apstra_datacenter_connectivity_templates_assignment" "a" {
  blueprint_id              = "b726704d-f80e-4733-9103-abd6ccd8752c"
  application_point_id      = "FkYtMBdeoJ5urBaIEi8"
  connectivity_template_ids = [
    "bcbcb35f-8f23-4bfb-916e-1b21d07d6904",
    "1f8ac61f-6996-42bb-a34f-4f4a50a7111a",
  ]
}
