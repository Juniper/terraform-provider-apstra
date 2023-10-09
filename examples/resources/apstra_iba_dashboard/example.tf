# The following example creates a Dashboard with Widgets in a blueprint.
# The dashboards references widgets by Id.
# These are pulled by using blueprint and widget data sources.

data "apstra_datacenter_blueprint" "b" {
  id = "evpn-vqfx_offbox-virtual"
}

data "apstra_iba_widget" "i" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Device Health Summary"
}

data "apstra_iba_widget" "j" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Fabric ECMP Imbalance"
}

data "apstra_iba_widget" "k" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "LAG Imbalance"
}

data "apstra_iba_widget" "l" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Systems with high disk utilization"
}

resource "apstra_iba_dashboard" "b" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  default = false
  description = "Test Dashboard"
  name = "Test"
  widget_grid = tolist([
    tolist([
      data.apstra_iba_widget.i.id, data.apstra_iba_widget.j.id
    ]),
    tolist([
      data.apstra_iba_widget.k.id
    ]),
    tolist([
      data.apstra_iba_widget.l.id
    ]),
  ])
}
