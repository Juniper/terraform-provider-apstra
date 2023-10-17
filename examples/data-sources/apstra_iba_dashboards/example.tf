# This example pulls all the iba dashboards from a blueprint

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}
data "apstra_iba_dashboards" "all" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
}

data "apstra_iba_dashboard" "all" {
  for_each = data.apstra_iba_dashboards.all.ids
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  id = each.key
}

output "od" {
  value = data.apstra_iba_dashboard.all
}

# Output looks something like this
#od = {
#  "4d6fb6a4-c140-49be-b39c-28a25fe0f948" = {
#    "blueprint_id" = "evpn-vqfx_offbox-virtual"
#    "default" = false
#    "description" = "Find issues in physical infrastructure that affect the available throughput caused by issues such as imbalanced traffic over a group of L3 (ECMP) or L2 (LAG) links."
#    "id" = "4d6fb6a4-c140-49be-b39c-28a25fe0f948"
#    "name" = "Throughput Health ESI"
#    "predefined_dashboard" = "esi_throughput_health"
#    "updated_by" = ""
#    "widget_grid" = tolist([
#      tolist([
#        "ffc10b11-0758-4a5e-897b-ec53b1978eb5",
#        "54a06aac-71ac-4001-bd87-91b57e31bdcb",
#      ]),
#      tolist([
#        "392009fd-1602-481a-b588-15db0d7f4fe6",
#      ]),
#    ])
#  }
#  "a2098e17-5a8e-4135-9f6e-73e56685fe7f" = {
#    "blueprint_id" = "evpn-vqfx_offbox-virtual"
#    "default" = false
#    "description" = "This dashboard presents the top 10 systems sorted by their aggregate interface utilization."
#    "id" = "a2098e17-5a8e-4135-9f6e-73e56685fe7f"
#    "name" = "Device Traffic Hotspots"
#    "predefined_dashboard" = "system_interface_utilization"
#    "updated_by" = ""
#    "widget_grid" = tolist([
#      tolist([
#        "6042b830-bdff-43b2-95f5-b6962eec630f",
#      ]),
#    ])
#  }
#}
