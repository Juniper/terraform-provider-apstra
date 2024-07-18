---
page_title: "apstra_freeform_property_set Resource - terraform-provider-apstra"
subcategory: "Reference Design: Freeform"
description: |-
  This resource creates a Property Set in a Freeform Blueprint.
---

# apstra_freeform_property_set (Resource)

This resource creates a Property Set in a Freeform Blueprint.


## Example Usage

```terraform
# Create a freeform property set resource.

resource "apstra_freeform_property_set" "prop_set_foo" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  name         = "prop_set_foo"
  values       = jsonencode({
    foo   = "bar"
    clown = 2
  })
}

# Read the property set back with a data source.

data "apstra_freeform_property_set" "foods" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  name         = apstra_freeform_property_set.prop_set_foo.name
}

# Output the property set.

output "foo" {value = data.apstra_freeform_property_set.foods}

# Output should look like:
#   foo = {
#     "blueprint_id" = "freeform_blueprint-5ba09d07"
#     "id" = tostring(null)
#     "name" = "prop_set_foo"
#     "system_id" = tostring(null)
#     "values" = "{\"clown\": 2, \"foo\": \"bar\"}"
#   }
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID.
- `name` (String) Property Set name as shown in the Web UI.
- `values` (String) A map of values in the Property Set in JSON format.

### Optional

- `system_id` (String) The system ID where the Property Set is associated.

### Read-Only

- `id` (String) ID of the Property Set.


