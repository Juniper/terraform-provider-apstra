
data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_datacenter_iba_widgets" "all" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
}


data "apstra_datacenter_iba_widget" "all" {
  for_each = data.apstra_datacenter_iba_widgets.all.ids
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  id = each.key
}

output "o" {
  value = data.apstra_datacenter_iba_widget.all
}

#Output looks something like this
#
#o = {
#  "2e6466ad-ed96-4773-9b55-87c7607e30ed" = {
#    "blueprint_id" = "cff966ad-f85f-478f-bae5-b64c1e58d31f"
#    "description"  = ""
#    "id"           = "2e6466ad-ed96-4773-9b55-87c7607e30ed"
#    "label"        = "Systems with high interface utilization"
#  }
#  "555fcdd7-ce9b-4bbb-9f90-43b921afe1d2" = {
#    "blueprint_id" = "cff966ad-f85f-478f-bae5-b64c1e58d31f"
#    "description"  = ""
#    "id"           = "555fcdd7-ce9b-4bbb-9f90-43b921afe1d2"
#    "label"        = "Systems with high memory utilization"
#  }
#}