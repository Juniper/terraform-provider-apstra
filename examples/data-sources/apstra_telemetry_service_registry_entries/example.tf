# The following example shows getting a list of
# Apstra Telemetry Service Registy Entries

# Get all Entries
data "apstra_telemetry_service_registry_entries" "all" {}
output "all_entries" {
  value = data.apstra_telemetry_service_registry_entries.all
}

#Get only built-in Entries
data "apstra_telemetry_service_registry_entries" "builtin" {
  built_in = true
}
output "builtin_entries" {
  value = data.apstra_telemetry_service_registry_entries.builtin
}

#Get Only non-built-in Entries
data "apstra_telemetry_service_registry_entries" "not_builtin" {
  built_in = false
}
output "not_builtin_entries" {
  value = data.apstra_telemetry_service_registry_entries.not_builtin
}

#Output will look something like this

#all_entries = {
#  "built_in"      = tobool(null)
#  "names" = toset([
#    "TestTelemetryServiceA",
#    "TestTelemetryServiceC",
#    "TestTelemetryServiceD",
#    "arp",
#    "bgp",
#    "bgp_communities",
#    "bgp_route",
#    "blueprint_collector",
#    "disk_util",
#    "dot1x",
#    "dot1x_hosts",
#    "environment",
#    "evpn_host_flap",
#    "evpn_host_flap_count",
#    "evpn_vxlan_type1",
#    "evpn_vxlan_type3",
#    "evpn_vxlan_type4",
#    "evpn_vxlan_type5",
#    "hostname",
#    "interface",
#    "interface_counters",
#    "lag",
#    "lldp",
#    "mac",
#    "mlag",
#    "multiagent_detector",
#    "nsxt",
#    "optical_xcvr",
#    "ospf_state",
#    "poe_controller",
#    "poe_interfaces",
#    "resource_util",
#    "route",
#    "route_lookup",
#    "shared_tunnel_mode",
#    "virtual_infra",
#    "vxlan_floodlist",
#    "xcvr",
#  ])
#}
#builtin_entries = {
#  "built_in"      = true
#  "names" = toset([
#    "arp",
#    "bgp",
#    "bgp_communities",
#    "bgp_route",
#    "blueprint_collector",
#    "disk_util",
#    "dot1x",
#    "dot1x_hosts",
#    "environment",
#    "evpn_host_flap",
#    "evpn_host_flap_count",
#    "evpn_vxlan_type1",
#    "evpn_vxlan_type3",
#    "evpn_vxlan_type4",
#    "evpn_vxlan_type5",
#    "hostname",
#    "interface",
#    "interface_counters",
#    "lag",
#    "lldp",
#    "mac",
#    "mlag",
#    "multiagent_detector",
#    "nsxt",
#    "optical_xcvr",
#    "ospf_state",
#    "poe_controller",
#    "poe_interfaces",
#    "resource_util",
#    "route",
#    "route_lookup",
#    "shared_tunnel_mode",
#    "virtual_infra",
#    "vxlan_floodlist",
#    "xcvr",
#  ])
#}
#not_builtin_entries = {
#  "built_in"      = false
#  "names" = toset([
#    "TestTelemetryServiceA",
#    "TestTelemetryServiceC",
#    "TestTelemetryServiceD",
#  ])
#}