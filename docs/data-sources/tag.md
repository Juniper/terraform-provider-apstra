---
page_title: "apstra_tag Data Source - terraform-provider-apstra"
subcategory: "Design"
description: |-
  This data source provides details of a specific Tag.
  At least one optional attribute is required.
---

# apstra_tag (Data Source)

This data source provides details of a specific Tag.

At least one optional attribute is required.


## Example Usage

```terraform
# The following example shows how a module might accept a tag key as an
# input variable,then use it to retrieve the appropriate tag when
# templating devices within a rack type.

variable "tag_key" {}

data "apstra_tag" "selected" {
    key = var.tag_key
}

resource "apstra_rack_type" "my_rack" {
    todo = "all of this"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) Apstra ID of the Tag. Required when `name` is omitted.
- `name` (String) Web UI name of the Tag. Required when `id` is omitted.

### Read-Only

- `description` (String) The description of the returned Tag.
