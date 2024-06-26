---
page_title: "apstra_tag Resource - terraform-provider-apstra"
subcategory: "Design"
description: |-
  This resource creates a Tag in the Apstra Design tab.
---

# apstra_tag (Resource)

This resource creates a Tag in the Apstra Design tab.


## Example Usage

```terraform
# this example creates a tags named after enterprise teams
# responsible for various data center asset types.
locals {
  device_owners = toset([
    "research",
    "security",
    "app team 1",
    "app team 2",
  ])
}

resource "apstra_tag" "example" {
  for_each    = local.device_owners
  name        = each.key
  description = format("device maintained by %q team", each.key)
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Tag name field as seen in the web UI.

### Optional

- `description` (String) Tag description field as seen in the web UI.

### Read-Only

- `id` (String) Apstra ID of the Tag.



