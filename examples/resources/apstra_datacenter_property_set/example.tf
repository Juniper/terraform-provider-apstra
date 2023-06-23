#reference a blueprint
data "apstra_datacenter_blueprint" "b" {
	name = "demo1"
}
#reference a property set from the global catalog
data "apstra_property_set" "ps" {
	name = "MGMT VRF"
}
#import property set into the blueprint, import only one key
resource "apstra_datacenter_property_set" "r" {
	blueprint_id = data.apstra_datacenter_blueprint.b.id
	id = data.apstra_property_set.ps.id
	keys = data.apstra_property_set.ps.keys
}
output "p" {
	value = resource.apstra_datacenter_property_set.r
}
#The output looks something like this
#p = {
#	"blueprint_id" = "d6c74373-45ce-4d88-9547-ac23c2ebe61e"
#	"data" = "{\"junos_mgmt_vrf\": \"mgmt_junos\", \"mgmt_vrf\": \"management\"}"
#	"id" = "3ae45f2e-c9ed-401b-8f00-367fb9a5e0e8"
#	"keys" =  toset([
#   "junos_mgmt_vrf",
#   "mgmt_vrf",
#   ])
#	"name" = "MGMT VRF"
#	"stale" = false
#}
