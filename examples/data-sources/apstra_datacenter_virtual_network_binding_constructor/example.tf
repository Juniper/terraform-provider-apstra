# This example creates the `bindings` data required by the
# `apstra_datacenter_virtual_network` resource when making a new Virtual Network
# for the 'blueprint_xyz' available on the switches marked with an '*' in the
# following topology:
#
#                                  blueprint_xyz
#
#  +--------------+  + -------------+   +------------+            +------------+
#  |              |  |              |   |          * |            |            |
#  |  esi-leaf-a  |  |  esi-leaf-b  |   |   leaf-c   |            |   leaf-d   |
#  |              |  |              |   |            |            |            |
#  +-+----------+-+  +-+----------+-+   +-----+------+            +-+--------+-+
#    |           \    /           |           |                     |        |
#    |            \  /            |           |                     |        |
#    |             \/             |           |                     |        |
#    |             /\             |           |                     |        |
#    |            /  \            |           |                     |        |
#    |           /    \           |           |                     |        |
#  +-+----------+-+  +-+----------+-+   +-----+------+   +----------+-+    +-+----------+
#  |            * |  |              |   |            |   |          * |    |          * |
#  | esi-access-a +--+ esi-access-b |   |  access-c  |   |  access-d  |    |  access-e  |
#  |              |  |              |   |            |   |            |    |            |
#  +--------------+  +--------------+   +------------+   +------------+    +------------+

data "apstra_datacenter_virtual_network_binding_constructor" "example" {
  blueprint_id = "blueprint_xyz"
  vlan_id      = 5 // optional; default behavior allows Apstra to choose
  switch_ids   = [ "esi-access-a", "leaf-c", "access-d", "access-e"]
}

# The output looks like the following:
#
#  bindings = {
#    // the ESI leaf group is included because of a downstream requirement
#    "esi-leaf-group-a-b" = {
#      // the ESI access group has been substituted in place of 'esi-access-a'
#      "access_ids" = [ "esi-access-group-a-b" ]
#      "vlan_id" = 5
#    },
#    // this leaf included directly without downstream switches
#    "leaf-c" = {
#      "access_ids" = []
#      "vlan_id" = 5
#    }
#    // this leaf included because of a downstream requirement
#    "leaf-d" = {
#      "access_ids" = [ "access-d", "access-e" ]
#      "vlan_id" = 5
#    }
#  }
