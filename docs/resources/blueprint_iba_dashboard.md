---
page_title: "apstra_blueprint_iba_dashboard Resource - terraform-provider-apstra"
subcategory: "Reference Design: Shared"
description: |-
  This resource creates a IBA Dashboard.
  Note: Compatible only with Apstra <5.0.0
---

# apstra_blueprint_iba_dashboard (Resource)

This resource creates a IBA Dashboard.

*Note: Compatible only with Apstra <5.0.0*


## Example Usage

```terraform
# The following example creates a Dashboard with Widgets in a blueprint.
# The dashboards references widgets by Id.
# These are pulled by using blueprint and widget data sources.

data "apstra_datacenter_blueprint" "b" {
  id = "evpn-vqfx_offbox-virtual"
}

data "apstra_blueprint_iba_widget" "i" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Device Health Summary"
}

data "apstra_blueprint_iba_widget" "j" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Fabric ECMP Imbalance"
}

data "apstra_blueprint_iba_widget" "k" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "LAG Imbalance"
}

data "apstra_blueprint_iba_widget" "l" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Systems with high disk utilization"
}

resource "apstra_blueprint_iba_dashboard" "b" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  default = false
  description = "Test Dashboard"
  name = "Test"
  widget_grid = tolist([
    tolist([
      data.apstra_blueprint_iba_widget.i.id, data.apstra_blueprint_iba_widget.j.id
    ]),
    tolist([
      data.apstra_blueprint_iba_widget.k.id
    ]),
    tolist([
      data.apstra_blueprint_iba_widget.l.id
    ]),
  ])
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID. Used to identify the Blueprint that the IBA Widget belongs to.
- `name` (String) IBA Dashboard Name.
- `widget_grid` (List of List of String) Grid of Widgets to be displayed in the IBA Dashboard

### Optional

- `default` (Boolean) True if Default IBA Dashboard
- `description` (String) Description of the IBA Dashboard

### Read-Only

- `id` (String) IBA Dashboard ID.
- `predefined_dashboard` (String) Id of predefined IBA Dashboard if any
- `updated_by` (String) The user who updated the IBA Dashboard last



