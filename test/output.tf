#output "d_agent_profile"                { value = data.apstra_agent_profile.d }
#output "d_agent_profiles"               { value = data.apstra_agent_profiles.d }
#output "d_asn_pool"                     { value = data.apstra_asn_pool.d }
#output "d_asn_pools"                    { value = data.apstra_asn_pools.d }
#output "d_blueprints"                   { value = data.apstra_blueprints.d }
#output "d_datacenter_blueprint_status"  { value = data.apstra_datacenter_blueprint_status.d }
#output "d_configlet"                    { value = data.apstra_configlet.d }
#output "d_ipv4_pool"                    { value = data.apstra_ipv4_pool.d }
#output "d_ipv4_pools"                   { value = data.apstra_ipv4_pools.d }
#output "d_ipv6_pool"                    { value = data.apstra_ipv6_pool.d }
#output "d_ipv6_pools"                   { value = data.apstra_ipv6_pools.d }
#output "d_logical_device"               { value = data.apstra_logical_device.d }
#output "d_datacenter_blueprint"         { value = data.apstra_datacenter_blueprint.d }
#output "d_interface_map"                { value = data.apstra_interface_map.d }
#output "d_interface_maps"               { value = data.apstra_interface_maps.imaps }
#output "d_tag"                          { value = data.apstra_tag.d }
#output "d_template_rack_based"          { value = data.apstra_template_rack_based.d }
#output "d_rack_type"                    { value = data.apstra_rack_type.d }
#output "d_vni_pool"                     { value = data.apstra_vni_pool.d }
#output "d_vni_pools"                    { value = data.apstra_vni_pools.d }

###############################################################################

#output "r_agent_profile"                { value = apstra_agent_profile.r }
#output "r_asn_pool"                     { value = apstra_asn_pool.r }
#output "r_ipv4_pool"                    { value = apstra_asn_pool.r }
#output "r_logical_device"               { value = apstra_logical_device.r }
#output "r_tag"                          { value = apstra_tag.r }
#output "r_interface_map"                { value = apstra_interface_map.r }
#output "r_rack_type"                    { value = apstra_rack_type.r }
#output "r_template_rack_based"          { value = apstra_template_rack_based.r }
#output "r_vni_pool"                     { value = apstra_vni_pool.r }
