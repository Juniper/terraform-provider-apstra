# This example assigns two connectivity templates to the switch port
# identified by the ID "FkYtMBdeoJ5urBaIEi8"
#
# Note that as of release v0.27.0 there's not an easy way to determine
# interface IDs. New data sources which facilitate interface ID lookup
# are coming soon (tm)

resource "apstra_datacenter_connectivity_template_assignment" "a" {
  blueprint_id              = "b726704d-f80e-4733-9103-abd6ccd8752c"
  application_point_id      = "FkYtMBdeoJ5urBaIEi8"
  connectivity_template_ids = [
    "bcbcb35f-8f23-4bfb-916e-1b21d07d6904",
    "1f8ac61f-6996-42bb-a34f-4f4a50a7111a",
  ]
}
