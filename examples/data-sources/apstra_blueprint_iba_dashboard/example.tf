# This example pulls one iba dashboards from a blueprint

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_blueprint_iba_dashboard" "i" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Device Health Summary"
}
output "pd" {
  value = data.apstra_blueprint_iba_dashboard.i
}

#pd = {
#  "blueprint_id" = "evpn-vqfx_offbox-virtual"
#  "default" = false
#  "description" = "The dashboard presents the data of utilization of system cpu, system memory and maximum disk utilization of a partition on every system present."
#  "id" = "ef0a6919-6c3f-46ce-aafa-83732d4474a8"
#  "name" = "Device Health Summary"
#  "predefined_dashboard" = "device_health_summary"
#  "updated_by" = ""
#  "widget_grid" = tolist([
#    tolist([
#      "e27b5fdb-f5e1-46d1-b83d-ec907a3788f8",
#    ]),
#    tolist([
#      "b7ff72b9-5e4b-465c-962a-6347b3ef45b2",
#    ]),
#    tolist([
#      "40c962df-64b9-4fe6-9d59-1cc55d64bc4d",
#    ]),
#  ])
#}
