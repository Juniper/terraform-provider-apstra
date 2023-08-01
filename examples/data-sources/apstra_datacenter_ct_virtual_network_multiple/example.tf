# Each apstra_datacenter_ct_* data source represents a Connectivity Template
# Primitive. They're stand-ins for the Primitives found in the Web UI's CT
# builder interface.
#
# These data sources do not interact with the Apstra API. Instead, they assemble
# their input fields into a JSON string presented at the `primitive` attribute
# key.
#
# Use the `primitive` output anywhere you need a primitive represented as JSON:
# - at the root of a Connectivity Template
# - as a child of another Primitive (as constrained by the accepts/produces
#   relationship between Primitives)

# Declare a "Virtual Network (Multiple)" Connectivity Template Primitive:
data "apstra_datacenter_ct_virtual_network_multiple" "hypervisor" {
  tagged_vn_ids = [
    apstra_datacenter_virtual_network.app_net_1.id,
    apstra_datacenter_virtual_network.app_net_2.id,
    apstra_datacenter_virtual_network.storage.id,
    apstra_datacenter_virtual_network.management.id,
  ]
  untagged_vn_id = apstra_datacenter_virtual_network.pxe.id
}

# The `primitive` output of this data source is the following JSON structure:
# {
#   "type": "AttachMultipleVLAN",
#   "data": {
#     "untagged_vn_id": "5Diy9UrebCGgWxSg",
#     "tagged_vn_ids": [
#       "vqMKZLmD+IYuRaGu",
#       "anSq7gCmz4whxGKY",
#       "3xBi3EqpKN2hyp9r",
#       "wJyxoX9jakVFmM6e"
#     ]
#   }
# }

# Use the `primitive` JSON when creating a Connectivity Template:
resource "apstra_datacenter_connectivity_template" "hypervisor" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name         = "hypervisor"
  tags         = [
    "prod",
    "hypervisor",
  ]
  primitives   = [
    data.apstra_datacenter_ct_virtual_network_multiple.hypervisor.primitive
  ]
}


