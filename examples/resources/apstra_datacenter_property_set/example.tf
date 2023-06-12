#reference a blueprint
data "apstra_datacenter_blueprint" "b" {
	name = "raj1"
}
#reference a property set from the global catalog
data "apstra_property_set" "ps" {
	name = "MGMT VRF"
}
#import property set into the blueprint, import only one key
resource "apstra_datacenter_property_set" "r" {
	blueprint_id = data.apstra_datacenter_blueprint.b.id
	id = data.apstra_property_set.ps.id
# uncomment below to import only one key. If no keys are mentioned, all keys are imported
#	keys = toset(["junos_mgmt_vrf"])
}
output "p" {
	value = resource.apstra_datacenter_property_set.r
}
#The output looks something like this
#p = {
#	"blueprint_id" = "d6c74373-45ce-4d88-9547-ac23c2ebe61e"
#	"data" = "{\"junos_mgmt_vrf\": \"mgmt_junos\", \"mgmt_vrf\": \"management\"}"
#	"id" = "3ae45f2e-c9ed-401b-8f00-367fb9a5e0e8"
#	"keys" = toset(null) /* of string */
#	"name" = "MGMT VRF"
#	"stale" = false
#}
