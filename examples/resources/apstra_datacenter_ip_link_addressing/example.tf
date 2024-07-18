# This example configures link numbering on a logical link between a leaf
# switch and a generic system. The link identified by "link_id" was created
# as a side-effect of attaching a Connectivity Template containing one or
# more IP Link primitives.
resource "apstra_datacenter_ip_link_addressing" "x" {
  blueprint_id = "22044be2-e7af-462d-847a-ce6d0b49000e"
  link_id      = "sz:HuZK45zx7V15D4qrjz0,vlan:22,a_001_leaf1<->a(link-000000001)[1]"

  switch_ipv4_address_type = "numbered"        # none | numbered
  switch_ipv4_address      = "192.0.2.0/31"
  switch_ipv6_address_type = "numbered"        # none | link_local | numbered
  switch_ipv6_address      = "2001:db8::1/127"

  generic_ipv4_address_type = "numbered"       # none | numbered
  generic_ipv4_address      = "192.0.2.1/31"
  generic_ipv6_address_type = "numbered"       # none | link_local | numbered
  generic_ipv6_address      = "2001:db8::2/127"
}
