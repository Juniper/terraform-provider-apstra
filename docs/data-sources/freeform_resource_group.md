---
page_title: "apstra_freeform_resource_group Data Source - terraform-provider-apstra"
subcategory: "Reference Design: Freeform"
description: |-
  This data source provides details of a specific Freeform Resource Allocation Group.
  At least one optional attribute is required.
---

# apstra_freeform_resource_group (Data Source)

This data source provides details of a specific Freeform Resource Allocation Group.

At least one optional attribute is required.


## Example Usage

```terraform
# This example defines a Freeform Resource Allocation Group in a blueprint

resource "apstra_freeform_resource_group" "test" {
  blueprint_id      = "freeform_blueprint-d8c1fabf"
  name              = "test_resource_group_fizz"
  data              =  jsonencode({
    foo   = "bar"
    clown = 2
  })
}

# here we retrieve the freeform resource_group

data "apstra_freeform_resource_group" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  id = apstra_freeform_resource_group.test.id
}

# here we build an output bock to display it

output "test_resource_group_out" {value = data.apstra_freeform_resource_group.test}

//test_resource_group_out = {
//  "blueprint_id" = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
//  "data" = "{\"clown\": 2, \"foo\": \"bar\"}"
//  "generator_id" = tostring(null)
//  "id" = "98ubU5cuRj7WsT159L4"
//  "name" = "test_resource_group_fizz"
//  "parent_id" = tostring(null)
//}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID. Used to identify the Blueprint where the Resource Allocation Group lives.

### Optional

- `id` (String) Populate this field to look up the Freeform Allocation Group by ID. Required when `name` is omitted.
- `name` (String) Populate this field to look up the Freeform Allocation Group by Name. Required when `id` is omitted.

### Read-Only

- `data` (String) Arbitrary key-value mapping that is useful in a context of this group. For example, you can store some VRF-related data there or add properties that are useful only in context of resource allocation, but not systems or interfaces.
- `generator_id` (String) ID of the group generator that created the group, if any.
- `parent_id` (String) ID of the group node that is present as a parent of the current one in a parent/child relationship. If this is a top-level (root) node, then `parent_id` will be `null`.