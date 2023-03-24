# https://cloudlabs.apstra.com/labguide/Cloudlabs/4.1.2/lab1-junos/lab1-junos-4_templates.html

## Create a template using previously looked-up (data) spine info and previously
## created (resource) rack types.
#resource "apstra_template_rack_based" "lab_guide" {
#  name                     = "apstra_junos"
#  asn_allocation_scheme    = "unique"
#  overlay_control_protocol = "evpn"
#  spine = {
#    count             = 2
#    logical_device_id = data.apstra_logical_device.lab_guide_switch.id
#  }
#  rack_infos = {
#    (apstra_rack_type.lab_guide_esi.id)    = { count = 1 }
#    (apstra_rack_type.lab_guide_single.id) = { count = 1 }
#  }
#}
