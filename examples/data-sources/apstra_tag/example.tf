# The following example shows how a module might accept a tag key as an input variable,then use it to retrieve the appropriate tag when templating devices within a rack type.

variable "tag_key" {}

data "apstra_tag" "selected" {
    key = var.tag_key
}

resource "apstra_rack_type" "my_rack" {
    todo = "all of this"
}
