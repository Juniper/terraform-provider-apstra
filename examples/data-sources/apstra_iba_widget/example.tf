# This example pulls one iba widget from a blueprint

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_iba_widget" "i" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Fabric ECMP Imbalance"
}
output "o" {
  value = data.apstra_iba_widget.i
}

#Output looks like this
#o = {
#  "blueprint_id" = "cff966ad-f85f-478f-bae5-b64c1e58d31f"
#  "description" = "Number of systems with ECMP imbalance."
#  "id" = "d03ba5ad-77f4-4198-8901-3c03fea7e341"
#  "name" = "Fabric ECMP Imbalance"
#}
